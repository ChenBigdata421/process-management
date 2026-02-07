package outbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox"
	outboxadapters "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters"
	gormadapter "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters/gorm"
	"gorm.io/gorm"
)

// OutboxSchedulerManager 管理单个或多个 Outbox Scheduler
type OutboxSchedulerManager struct {
	schedulers []*outbox.OutboxScheduler
	publishers []*outbox.OutboxPublisher // ✅ 保存所有 Publisher 的引用（用于启动 ACK 监听器）
	publisher  outbox.EventPublisher
	mapper     outbox.TopicMapper
	mu         sync.RWMutex
	running    bool
}

// NewOutboxSchedulerManager 创建 Scheduler 管理器
func NewOutboxSchedulerManager(publisher outbox.EventPublisher, mapper outbox.TopicMapper) *OutboxSchedulerManager {
	return &OutboxSchedulerManager{
		schedulers: make([]*outbox.OutboxScheduler, 0),
		publishers: make([]*outbox.OutboxPublisher, 0),
		publisher:  publisher,
		mapper:     mapper,
	}
}

// Start 启动所有 Scheduler
// 根据多租户配置开关，创建单个（非多租户）或多个 Scheduler（每个租户一个）
func (m *OutboxSchedulerManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil // 已经在运行
	}

	// 检查 EventPublisher 是否已初始化
	if m.publisher == nil {
		logger.Warn("EventPublisher is not initialized, Outbox Scheduler will not start. This is expected if EventBus (Kafka/NATS) is not configured.")
		m.running = false
		return nil // 不启动调度器，但不报错
	}

	// 检查是否启用多租户
	if !config.TenantsConfig.Enabled {
		// 单租户模式：创建一个 Scheduler（使用默认租户 "*"）
		logger.Info("Starting Outbox Scheduler in single-tenant mode (using default tenant '*')")

		db := sdk.Runtime.GetTenantDB("*")
		if db == nil {
			logger.Fatal("Failed to get database connection for single-tenant mode")
		}

		// ✅ 单租户模式：注册默认租户 "*" 并使用租户专属 ACK Channel
		adapter, ok := m.publisher.(*outboxadapters.EventBusAdapter)
		if !ok {
			logger.Fatal("EventPublisher is not an EventBusAdapter, tenant ACK not supported")
		}

		// 注册默认租户 "*"（缓冲区大小 10000）
		if err := adapter.RegisterTenant("*", 10000); err != nil {
			logger.Fatalf("Failed to register default tenant ACK channel: %v", err)
		}
		logger.Info("✅ Registered default tenant ACK channel (tenant ID: *)")

		// 创建 Scheduler 和 Publisher
		scheduler, publisher := m.createScheduler("*", db)

		// 获取默认租户的 ACK Channel
		tenantACKChan := adapter.GetTenantPublishResultChannel("*")
		if tenantACKChan == nil {
			logger.Fatal("Failed to get tenant ACK channel for default tenant '*'")
		}

		// 使用租户专属 ACK Channel
		publisher.StartACKListenerWithChannel(ctx, tenantACKChan)
		logger.Info("✅ ACK listener started for single-tenant mode (using tenant-specific ACK channel for '*')")

		// 启动调度器
		if err := scheduler.Start(ctx); err != nil {
			logger.Fatalf("Failed to start outbox scheduler: %v", err)
		}

		m.schedulers = append(m.schedulers, scheduler)
		m.publishers = append(m.publishers, publisher)
		logger.Info("✅ Outbox scheduler started successfully (single-tenant mode with tenant '*')")
	} else {
		// 多租户模式：为每个租户创建一个 Scheduler
		logger.Info("Starting Outbox Schedulers in multi-tenant mode with staggered polling")

		// 第一步：收集所有租户信息
		type tenantInfo struct {
			id string
			db *gorm.DB
		}
		var tenants []tenantInfo

		sdk.Runtime.GetTenantDBs(func(tenantID string, db *gorm.DB) bool {
			tenants = append(tenants, tenantInfo{id: tenantID, db: db})
			return true
		})

		tenantCount := len(tenants)
		if tenantCount == 0 {
			logger.Warn("No tenants found in multi-tenant mode")
			m.running = true
			return nil
		}

		// 第二步：为每个租户注册 ACK Channel
		adapter, ok := m.publisher.(*outboxadapters.EventBusAdapter)
		if !ok {
			logger.Fatal("EventPublisher is not an EventBusAdapter, multi-tenant ACK not supported")
		}

		for _, tenant := range tenants {
			// ✅ 注册租户 ACK Channel（缓冲区大小 10000）
			if err := adapter.RegisterTenant(tenant.id, 10000); err != nil {
				logger.Errorf("Failed to register tenant ACK channel for %s: %v", tenant.id, err)
				return err
			}
			logger.Infof("✅ Registered tenant ACK channel for tenant: %s", tenant.id)
		}

		// 第三步：动态计算错开时间
		pollInterval := 1 * time.Second
		staggerInterval := pollInterval / time.Duration(tenantCount)

		logger.Infof("Calculated stagger interval: %dms for %d tenant(s)",
			staggerInterval.Milliseconds(), tenantCount)

		// 第四步：为每个租户创建并启动 Scheduler
		for i, tenant := range tenants {
			startDelay := time.Duration(i) * staggerInterval

			logger.Infof("Creating Outbox Scheduler for tenant: %s (index: %d, start delay: %dms)",
				tenant.id, i, startDelay.Milliseconds())

			// 创建 Scheduler 和 Publisher
			scheduler, publisher := m.createScheduler(tenant.id, tenant.db)

			// ✅ 获取租户专属的 ACK Channel
			tenantACKChan := adapter.GetTenantPublishResultChannel(tenant.id)
			if tenantACKChan == nil {
				logger.Errorf("Failed to get ACK channel for tenant %s", tenant.id)
				return fmt.Errorf("failed to get ACK channel for tenant %s", tenant.id)
			}

			// ✅ 启动 ACK 监听器（使用租户专属 Channel）
			publisher.StartACKListenerWithChannel(ctx, tenantACKChan)
			logger.Infof("✅ ACK listener started for tenant: %s (using tenant-specific ACK channel)", tenant.id)

			// 错开启动时间
			currentTenantID := tenant.id
			currentDelay := startDelay
			currentIndex := i
			go func() {
				if currentDelay > 0 {
					time.Sleep(currentDelay)
				}

				if err := scheduler.Start(ctx); err != nil {
					logger.Errorf("Failed to start outbox scheduler for tenant %s: %v", currentTenantID, err)
					return
				}

				logger.Infof("✅ Outbox scheduler started successfully for tenant: %s (index: %d, delayed by %dms)",
					currentTenantID, currentIndex, currentDelay.Milliseconds())
			}()

			m.schedulers = append(m.schedulers, scheduler)
			m.publishers = append(m.publishers, publisher)
		}

		logger.Infof("✅ Scheduled %d Outbox Scheduler(s) to start with dynamic staggered delays (interval: %dms)",
			len(m.schedulers), staggerInterval.Milliseconds())
	}

	m.running = true
	return nil
}

// Stop 停止所有 Scheduler
func (m *OutboxSchedulerManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil // 未运行
	}

	logger.Infof("Stopping %d Outbox Scheduler(s)...", len(m.schedulers))

	// 停止所有调度器
	var stopErrors []error
	for i, scheduler := range m.schedulers {
		if err := scheduler.Stop(ctx); err != nil {
			logger.Errorf("Failed to stop scheduler %d: %v", i, err)
			stopErrors = append(stopErrors, err)
		}
	}

	// ✅ 注销所有租户的 ACK Channel
	adapter, ok := m.publisher.(*outboxadapters.EventBusAdapter)
	if ok {
		registeredTenants := adapter.GetRegisteredTenants()
		for _, tenantID := range registeredTenants {
			if err := adapter.UnregisterTenant(tenantID); err != nil {
				logger.Errorf("Failed to unregister tenant %s: %v", tenantID, err)
			} else {
				logger.Infof("✅ Unregistered tenant ACK channel for tenant: %s", tenantID)
			}
		}
	}

	m.schedulers = nil
	m.publishers = nil
	m.running = false

	if len(stopErrors) > 0 {
		logger.Warnf("Some schedulers failed to stop: %d errors", len(stopErrors))
	} else {
		logger.Info("✅ All Outbox Schedulers stopped successfully")
	}

	return nil
}

// IsRunning 检查是否正在运行
func (m *OutboxSchedulerManager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// GetSchedulerCount 获取 Scheduler 数量
func (m *OutboxSchedulerManager) GetSchedulerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.schedulers)
}

// createScheduler 创建单个 Scheduler 和 Publisher
func (m *OutboxSchedulerManager) createScheduler(tenantID string, db *gorm.DB) (*outbox.OutboxScheduler, *outbox.OutboxPublisher) {
	// 创建租户专属的 Repository
	repo := gormadapter.NewGormOutboxRepository(db)

	// 创建调度器配置
	schedulerConfig := &outbox.SchedulerConfig{
		PollInterval:        1 * time.Second,  // 轮询间隔 1 秒
		BatchSize:           100,              // 每次处理 100 个事件
		TenantID:            tenantID,         // 租户 ID
		CleanupInterval:     1 * time.Hour,    // 清理间隔 1 小时
		CleanupRetention:    24 * time.Hour,   // 保留 24 小时
		HealthCheckInterval: 30 * time.Second, // 健康检查间隔 30 秒
		EnableHealthCheck:   false,            // 禁用健康检查
		EnableCleanup:       true,             // 启用自动清理
		EnableMetrics:       false,            // 禁用指标收集
		EnableRetry:         true,             // 启用失败重试
		RetryInterval:       30 * time.Second, // 重试间隔 30 秒
		MaxRetries:          3,                // 最大重试次数 3 次
		EnableDLQ:           false,            // 禁用死信队列
	}

	// 创建发布器配置
	publisherConfig := &outbox.PublisherConfig{
		MaxRetries:     5,                // 最大重试次数
		RetryDelay:     2 * time.Second,  // 重试延迟
		PublishTimeout: 30 * time.Second, // 发布超时
		EnableMetrics:  true,             // 启用指标收集
	}

	// ✅ 先创建 Publisher（用于启动 ACK 监听器）
	publisher := outbox.NewOutboxPublisher(
		repo,
		m.publisher,
		m.mapper,
		publisherConfig,
	)

	// 创建调度器
	scheduler := outbox.NewScheduler(
		outbox.WithRepository(repo),
		outbox.WithEventPublisher(m.publisher),
		outbox.WithTopicMapper(m.mapper),
		outbox.WithSchedulerConfig(schedulerConfig),
		outbox.WithPublisherConfig(publisherConfig),
	)

	return scheduler, publisher
}
