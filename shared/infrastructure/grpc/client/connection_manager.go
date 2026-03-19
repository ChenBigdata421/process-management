package grpc_client

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
)

// ConnectionManager 专注于gRPC连接管理
// 职责：连接创建、健康检查、连接状态管理、自动重连
type ConnectionManager struct {
	client zrpc.Client
	config config.GRPCClientConfig

	// 连接状态管理
	isHealthy bool
	lastCheck time.Time
	mu        sync.RWMutex

	// 连接初始化
	conn *grpc.ClientConn
	once sync.Once

	// 重连机制
	reconnectAttempts   int32         // 重连次数
	maxReconnectCount   int32         // 最大重连次数
	reconnectInterval   time.Duration // 重连间隔
	healthCheckInterval time.Duration // 健康检查间隔
	isReconnecting      bool          // 是否正在重连
	reconnectMu         sync.Mutex

	// 停止信号
	stopCh chan struct{}
}

// ReconnectConfig 重连配置
type ReconnectConfig struct {
	MaxRetries     int32         `json:"max_retries" yaml:"max_retries" default:"5"`           // 最大重连次数
	RetryInterval  time.Duration `json:"retry_interval" yaml:"retry_interval" default:"10s"`   // 重连间隔
	HealthInterval time.Duration `json:"health_interval" yaml:"health_interval" default:"30s"` // 健康检查间隔
}

// getDefaultReconnectConfig 获取默认重连配置
func getDefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxRetries:     5,
		RetryInterval:  10 * time.Second,
		HealthInterval: 30 * time.Second,
	}
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(cfg config.GRPCClientConfig, etcdConfig config.ETCDConfig) (*ConnectionManager, error) {
	logger.Infof("创建gRPC连接管理器，配置: ServiceKey=%s, Timeout=%d",
		cfg.ServiceKey, cfg.Timeout)

	// 构建zrpc配置 - 使用服务发现模式
	conf := zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: etcdConfig.Hosts,
			Key:   cfg.ServiceKey,
		},
		Timeout: cfg.Timeout,
	}

	client, err := zrpc.NewClient(conf)
	if err != nil {
		logger.Errorf("创建zrpc客户端失败: %v", err)
		return nil, err
	}

	// 获取重连配置（可以考虑从配置文件读取）
	reconnectCfg := getDefaultReconnectConfig()

	manager := &ConnectionManager{
		client:              client,
		config:              cfg,
		isHealthy:           true,
		maxReconnectCount:   reconnectCfg.MaxRetries,
		reconnectInterval:   reconnectCfg.RetryInterval,
		healthCheckInterval: reconnectCfg.HealthInterval,
		stopCh:              make(chan struct{}),
	}

	// 启动健康检查
	go manager.startHealthCheck()

	logger.Infof("gRPC连接管理器创建完成，重连配置: 最大重试%d次, 重试间隔%v, 健康检查间隔%v",
		reconnectCfg.MaxRetries, reconnectCfg.RetryInterval, reconnectCfg.HealthInterval)
	return manager, nil
}

// GetConnection 获取gRPC连接（懒加载）
func (m *ConnectionManager) GetConnection() *grpc.ClientConn {
	m.once.Do(func() {
		m.conn = m.client.Conn()
		logger.Info("gRPC连接初始化完成")
	})
	return m.conn
}

// IsHealthy 检查连接是否健康
func (m *ConnectionManager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isHealthy
}

// GetConnectionState 获取连接状态
func (m *ConnectionManager) GetConnectionState() connectivity.State {
	conn := m.GetConnection()
	return conn.GetState()
}

// 健康检查
func (m *ConnectionManager) startHealthCheck() {
	// 使用配置的健康检查间隔
	ticker := time.NewTicker(m.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.CheckHealth()
		case <-m.stopCh:
			logger.Info("健康检查停止")
			return
		}
	}
}

// CheckHealth 公开的健康检查方法，允许外部主动触发健康检查
func (m *ConnectionManager) CheckHealth() {
	m.checkHealth()
}

func (m *ConnectionManager) checkHealth() {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.GetConnectionState()
	isHealthy := state == connectivity.Ready || state == connectivity.Idle
	wasHealthy := m.isHealthy
	m.isHealthy = isHealthy
	m.lastCheck = time.Now()

	if !isHealthy {
		logger.Warnf("gRPC连接不健康，状态: %v", state)

		// 如果之前是健康的，现在变成不健康，启动重连
		if wasHealthy {
			logger.Info("检测到连接状态从健康变为不健康，启动重连机制")
			go m.attemptReconnect()
		}
	} else if !wasHealthy && isHealthy {
		// 连接恢复健康
		logger.Info("gRPC连接已恢复健康")
		m.resetReconnectState()
	}
}

// attemptReconnect 尝试重连
func (m *ConnectionManager) attemptReconnect() {
	m.reconnectMu.Lock()
	defer m.reconnectMu.Unlock()

	// 如果已经在重连中，直接返回
	if m.isReconnecting {
		return
	}
	m.isReconnecting = true
	defer func() { m.isReconnecting = false }()

	logger.Info("开始尝试重连gRPC连接")

	for m.reconnectAttempts < m.maxReconnectCount {
		m.reconnectAttempts++

		logger.Infof("第 %d/%d 次重连尝试", m.reconnectAttempts, m.maxReconnectCount)

		// 等待重连间隔
		select {
		case <-time.After(m.reconnectInterval):
		case <-m.stopCh:
			logger.Info("重连被停止")
			return
		}

		// 重新创建连接
		if err := m.recreateConnection(); err != nil {
			logger.Errorf("重连失败: %v", err)
			continue
		}

		// 检查新连接状态
		state := m.GetConnectionState()

		if state == connectivity.Ready || state == connectivity.Idle {
			logger.Infof("重连成功，连接状态: %v", state)
			m.resetReconnectState()
			return
		}

		logger.Warnf("重连后连接状态仍不正常: %v", state)
	}

	logger.Errorf("达到最大重连次数 %d，停止重连", m.maxReconnectCount)
}

// recreateConnection 重新创建连接
func (m *ConnectionManager) recreateConnection() error {
	// 先关闭旧连接
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}

	// 重新创建zrpc客户端
	conf := zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: config.EtcdConfig.Hosts,
			Key:   m.config.ServiceKey,
		},
		Timeout: m.config.Timeout,
	}

	client, err := zrpc.NewClient(conf)
	if err != nil {
		return err
	}

	m.client = client

	// 重置once，强制重新初始化连接
	m.once = sync.Once{}

	return nil
}

// resetReconnectState 重置重连状态
func (m *ConnectionManager) resetReconnectState() {
	m.reconnectAttempts = 0
	m.isReconnecting = false

	m.mu.Lock()
	m.isHealthy = true
	m.mu.Unlock()
}

// IsReconnecting 检查是否正在重连
func (m *ConnectionManager) IsReconnecting() bool {
	m.reconnectMu.Lock()
	defer m.reconnectMu.Unlock()
	return m.isReconnecting
}

// GetReconnectAttempts 获取重连次数
func (m *ConnectionManager) GetReconnectAttempts() int32 {
	m.reconnectMu.Lock()
	defer m.reconnectMu.Unlock()
	return m.reconnectAttempts
}

// Close 关闭连接
func (m *ConnectionManager) Close() error {
	// 停止健康检查
	close(m.stopCh)

	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// ============================================
// 网络级别重试机制 - 对业务层透明
// ============================================

// ExecuteWithNetworkRetry 执行gRPC调用，自动处理网络级别的重试
// 这个方法对业务层透明，自动处理网络连接问题
func (m *ConnectionManager) ExecuteWithNetworkRetry(ctx context.Context, operation func() error) error {
	maxNetworkRetries := 3 // 网络级别最多重试3次

	for attempt := 0; attempt < maxNetworkRetries; attempt++ {
		// 等待连接健康（如果正在重连中）
		if err := m.waitForHealthyConnection(ctx, 5*time.Second); err != nil {
			logger.Warn("等待连接健康超时", "attempt", attempt+1, "error", err)
			if attempt == maxNetworkRetries-1 {
				return fmt.Errorf("连接不可用: %w", err)
			}
			continue
		}

		// 执行操作
		err := operation()
		if err == nil {
			// 成功，直接返回
			return nil
		}

		// 判断是否为网络错误
		if !m.isNetworkError(err) {
			// 非网络错误（如业务错误、认证错误等），直接返回
			logger.Debug("非网络错误，不进行重试", "error", err)
			return err
		}

		// 网络错误，触发健康检查
		logger.Warn("检测到网络错误，触发连接检查",
			"error", err,
			"attempt", attempt+1,
			"maxRetries", maxNetworkRetries)

		m.CheckHealth()

		// 如果不是最后一次尝试，等待后重试
		if attempt < maxNetworkRetries-1 {
			retryDelay := time.Duration(attempt+1) * time.Second // 1s, 2s, 3s
			logger.Info("网络重试等待中", "delay", retryDelay, "nextAttempt", attempt+2)

			timer := time.NewTimer(retryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
				// 继续下一次重试
			}
		}
	}

	return fmt.Errorf("网络重试失败，已达最大重试次数 %d", maxNetworkRetries)
}

// waitForHealthyConnection 等待连接变为健康状态
func (m *ConnectionManager) waitForHealthyConnection(ctx context.Context, timeout time.Duration) error {
	if m.IsHealthy() {
		return nil
	}

	logger.Info("连接不健康，等待恢复", "timeout", timeout)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeoutTimer.C:
			return fmt.Errorf("等待连接健康超时: %v", timeout)
		case <-ticker.C:
			if m.IsHealthy() {
				logger.Info("连接已恢复健康")
				return nil
			}
		}
	}
}

// isNetworkError 判断是否为网络相关错误
func (m *ConnectionManager) isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// 检查常见的网络错误模式
	networkErrorPatterns := []string{
		"connection refused",
		"connection reset by peer",
		"connection closed",
		"the client connection is closing",
		"transport is closing",
		"rpc error: code = Unavailable",
		"rpc error: code = DeadlineExceeded",
		"rpc error: code = Canceled desc = grpc: the client connection is closing",
		"dial tcp",
		"no such host",
		"network is unreachable",
		"timeout",
		"context deadline exceeded",
	}

	for _, pattern := range networkErrorPatterns {
		if strings.Contains(strings.ToLower(errMsg), pattern) {
			return true
		}
	}

	// 检查gRPC状态码
	if grpcStatus, ok := status.FromError(err); ok {
		switch grpcStatus.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
			return true
		}
	}

	return false
}
