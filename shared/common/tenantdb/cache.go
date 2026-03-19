// shared/common/tenantdb/cache.go
package tenantdb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	toolsConfig "github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/tenant/provider"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Cache 租户数据缓存
type Cache struct {
	provider    *provider.Provider
	etcdClient  *clientv3.Client
	ctx         context.Context
	cancel      context.CancelFunc
	serviceCode string // "evidence-command" 或 "evidence-query"
	done        chan struct{}
}

// global 全局缓存实例（单例模式）
var global *Cache
var globalMu sync.RWMutex

// Setup 初始化租户数据缓存
//
// 在 database.Setup() 之前调用
// serviceCode: "evidence-command" 或 "evidence-query"
func Setup(serviceCode string) (*Cache, error) {
	// 1. 检查是否已经初始化（单例）
	globalMu.RLock()
	if global != nil && global.GetServiceCode() == serviceCode {
		globalMu.RUnlock()
		return global, nil
	}
	globalMu.RUnlock()

	// 2. 检查 ETCD 配置
	if toolsConfig.EtcdConfig == nil || !toolsConfig.EtcdConfig.Enabled {
		log.Println("[tenantdb] ETCD 未启用，跳过租户缓存初始化")
		return nil, nil
	}

	// 3. 创建 ETCD 客户端
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   toolsConfig.EtcdConfig.Hosts,
		DialTimeout: time.Duration(toolsConfig.EtcdConfig.DialTimeout) * time.Second,
		Username:    toolsConfig.EtcdConfig.Username,
		Password:    toolsConfig.EtcdConfig.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ETCD 客户端失败: %w", err)
	}

	// 4. 创建 Provider（带重试和缓存）
	ctx, cancel := context.WithCancel(context.Background())
	prov, err := provider.NewProviderWithRetry(etcdClient,
		provider.WithNamespace(toolsConfig.EtcdConfig.Namespace),
		provider.WithConfigTypes(provider.ConfigTypeDatabase),
		provider.WithCache(provider.NewFileCache()),
	)
	if err != nil {
		cancel()
		etcdClient.Close()
		return nil, fmt.Errorf("创建 Provider 失败: %w", err)
	}

	// 5. 启动 ETCD Watch（后台监听）
	if err := prov.StartWatch(ctx); err != nil {
		cancel()
		prov.StopWatch()
		etcdClient.Close()
		return nil, fmt.Errorf("启动 Provider Watch 失败: %w", err)
	}

	// 6. 从 ETCD 加载所有租户数据
	log.Println("[tenantdb] 等待 Provider 从 ETCD 加载租户数据...")
	loadCtx, cancelLoad := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelLoad()
	if err := prov.LoadAll(loadCtx); err != nil {
		log.Printf("[tenantdb] 警告：从 ETCD 加载租户数据失败: %v", err)
		// 不返回错误，继续执行
	}

	// 等待 Provider 完成加载
	time.Sleep(2 * time.Second)

	tenantCount := len(prov.GetAllTenantIDs())
	log.Printf("[tenantdb] 租户数据缓存初始化成功，共加载 %d 个租户", tenantCount)

	// 7. 创建缓存实例并设置全局
	cache := &Cache{
		provider:    prov,
		etcdClient:  etcdClient,
		ctx:         ctx,
		cancel:      cancel,
		serviceCode: serviceCode,
		done:        make(chan struct{}),
	}

	globalMu.Lock()
	global = cache
	globalMu.Unlock()

	return cache, nil
}

// GetGlobalCache 获取全局缓存实例
func GetGlobalCache() *Cache {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return global
}

// GetTenantIDs 获取所有租户 ID 列表
func (c *Cache) GetTenantIDs() []int {
	if c == nil || c.provider == nil {
		return []int{}
	}
	return c.provider.GetAllTenantIDs()
}

// GetTenantMeta 获取租户元数据
func (c *Cache) GetTenantMeta(tenantID int) (*provider.TenantMeta, error) {
	if c == nil || c.provider == nil {
		return nil, ErrCacheNotInitialized
	}
	meta, ok := c.provider.GetTenantMeta(tenantID)
	if !ok {
		return nil, fmt.Errorf("租户 %d 的元数据不存在", tenantID)
	}
	return meta, nil
}

// GetDatabaseConfig 获取租户数据库配置
func (c *Cache) GetDatabaseConfig(tenantID int) (*provider.ServiceDatabaseConfig, error) {
	if c == nil || c.provider == nil {
		return nil, ErrCacheNotInitialized
	}

	cfg, ok := c.provider.GetServiceDatabaseConfig(tenantID, c.serviceCode)
	if !ok {
		return nil, fmt.Errorf("租户 %d 服务 %s 的数据库配置不存在", tenantID, c.serviceCode)
	}

	// 处理密码解密
	password := cfg.Password
	if cfg.HasEncryptedPassword() {
		encryptionKey := os.Getenv("ENCRYPTION_KEY")
		if encryptionKey == "" {
			return nil, fmt.Errorf("密码已加密但未设置 ENCRYPTION_KEY 环境变量")
		}
		decrypted, err := cfg.DecryptPassword(encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("解密数据库密码失败: %w", err)
		}
		password = decrypted
	}

	// 返回解密后的配置副本
	result := *cfg
	result.Password = password
	return &result, nil
}

// Count 获取缓存的租户数量
func (c *Cache) Count() int {
	if c == nil || c.provider == nil {
		return 0
	}
	return len(c.provider.GetAllTenantIDs())
}

// Close 关闭缓存
func (c *Cache) Close() error {
	if c == nil {
		return nil
	}
	if c.cancel != nil {
		c.cancel()
	}
	if c.provider != nil {
		c.provider.StopWatch()
	}
	if c.etcdClient != nil {
		return c.etcdClient.Close()
	}
	close(c.done)
	return nil
}

// GetServiceCode 获取服务代码
func (c *Cache) GetServiceCode() string {
	if c == nil {
		return ""
	}
	return c.serviceCode
}

// GetTenantIDByDomain implements DomainLookuper interface.
// It looks up the tenant ID by domain name using the internal Provider.
// Returns (0, false) if cache is nil or provider is nil.
// Time complexity: O(1) lookup in domain index.
func (c *Cache) GetTenantIDByDomain(domain string) (int, bool) {
	if c == nil {
		return 0, false
	}
	if c.provider == nil {
		return 0, false
	}
	return c.provider.GetTenantIDByDomain(domain)
}

// GetProvider returns the internal Provider instance.
// Used by domain-based tenant identification middleware.
// Returns nil if Cache is nil (defensive nil check).
func (c *Cache) GetProvider() *provider.Provider {
	if c == nil {
		return nil
	}
	return c.provider
}

// ErrCacheNotInitialized 缓存未初始化错误
var ErrCacheNotInitialized = errors.New("tenant cache not initialized")
