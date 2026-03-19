package outbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk"
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

	// 收集所有租户信息
	type tenantInfo struct {
		id int
		db *gorm.DB
	}
	var tenants []tenantInfo

	sdk.Runtime.GetTenantDBs(func(tenantID int, db *gorm.DB) bool {
		tenants = append(tenants, tenantInfo{id: tenantID, db: db})
		return true
	})

	// 如果没有租户，使用单租户模式（默认租户 ID: 1）
	if len(tenants) == 0 {
		// 单租户模式：创建一个 Scheduler（使用默认租户 ID: 1）
		logger.Info("Starting Outbox Scheduler in single-tenant mode (using default tenant ID: 1)")

		db := sdk.Runtime.GetTenantDB(1)
		if db == nil {
			logger.Fatal("Failed to get database connection for single-tenant mode")
		}

		// ✅ 单租户模式：注册默认租户 1 并使用租户专属 ACK Channel
		adapter, ok := m.publisher.(*outboxadapters.EventBusAdapter)
		if !ok {
			logger.Fatal("EventPublisher is not an EventBusAdapter, tenant ACK not supported")
		}

		// 注册默认租户 1（缓冲区大小 10000）
		if err := adapter.RegisterTenant(1, 10000); err != nil {
			logger.Fatalf("Failed to register default tenant ACK channel: %v", err)
		}
		logger.Info("✅ Registered default tenant ACK channel (tenant ID: 1)")

		// 创建 Scheduler 和 Publisher
		scheduler, publisher := m.createScheduler(1, db)

		// 获取默认租户的 ACK Channel
		tenantACKChan := adapter.GetTenantPublishResultChannel(1)
		if tenantACKChan == nil {
			logger.Fatal("Failed to get tenant ACK channel for default tenant '1'")
		}

		// 使用租户专属 ACK Channel
		publisher.StartACKListenerWithChannel(ctx, tenantACKChan)
		logger.Info("✅ ACK listener started for single-tenant mode (using tenant-specific ACK channel for 1)")

		// 启动调度器
		if err := scheduler.Start(ctx); err != nil {
			logger.Fatalf("Failed to start outbox scheduler: %v", err)
		}

		m.schedulers = append(m.schedulers, scheduler)
		m.publishers = append(m.publishers, publisher)
		logger.Info("✅ Outbox scheduler started successfully (single-tenant mode with tenant ID: 1)")
	} else {
		// 多租户模式：为每个租户创建一个 Scheduler
		logger.Info("Starting Outbox Schedulers in multi-tenant mode with staggered polling")

		tenantCount := len(tenants)
		if tenantCount == 0 {
			logger.Warn("No tenants found in multi-tenant mode")
			m.running = true
			return nil
		}

		// 第二步：为每个租户注册 ACK Channel
		// 将 m.publisher 转换为 *EventBusAdapter 以访问多租户方法
		adapter, ok := m.publisher.(*outboxadapters.EventBusAdapter)
		if !ok {
			logger.Fatal("EventPublisher is not an EventBusAdapter, multi-tenant ACK not supported")
		}

		for _, tenant := range tenants {
			// ✅ 注册租户 ACK Channel（缓冲区大小 10000）
			if err := adapter.RegisterTenant(tenant.id, 10000); err != nil {
				logger.Errorf("Failed to register tenant ACK channel for %d: %v", tenant.id, err)
				return err
			}
			logger.Infof("✅ Registered tenant ACK channel for tenant: %d", tenant.id)
		}

		// 第三步：动态计算错开时间
		// 策略：将轮询间隔（1秒）平均分配给所有租户
		// 例如：3个租户 → 333ms, 5个租户 → 200ms, 10个租户 → 100ms
		pollInterval := 1 * time.Second
		staggerInterval := pollInterval / time.Duration(tenantCount)

		logger.Infof("Calculated stagger interval: %dms for %d tenant(s)",
			staggerInterval.Milliseconds(), tenantCount)

		// 第四步：为每个租户创建并启动 Scheduler
		for i, tenant := range tenants {
			startDelay := time.Duration(i) * staggerInterval

			logger.Infof("Creating Outbox Scheduler for tenant: %d (index: %d, start delay: %dms)",
				tenant.id, i, startDelay.Milliseconds())

			// 创建 Scheduler 和 Publisher
			scheduler, publisher := m.createScheduler(tenant.id, tenant.db)

			// ✅ 获取租户专属的 ACK Channel（通过 EventBusAdapter）
			tenantACKChan := adapter.GetTenantPublishResultChannel(tenant.id)
			if tenantACKChan == nil {
				logger.Errorf("Failed to get ACK channel for tenant %d", tenant.id)
				return fmt.Errorf("failed to get ACK channel for tenant %d", tenant.id)
			}

			// ✅ 启动 ACK 监听器（使用租户专属 Channel）
			publisher.StartACKListenerWithChannel(ctx, tenantACKChan)
			logger.Infof("✅ ACK listener started for tenant: %d (using tenant-specific ACK channel)", tenant.id)

			// 错开启动时间：在单独的 goroutine 中延迟启动
			currentTenantID := tenant.id // 捕获当前租户 ID
			currentDelay := startDelay   // 捕获当前延迟
			currentIndex := i            // 捕获当前索引
			go func() {
				// 延迟启动
				if currentDelay > 0 {
					time.Sleep(currentDelay)
				}

				// 启动调度器
				if err := scheduler.Start(ctx); err != nil {
					logger.Errorf("Failed to start outbox scheduler for tenant %d: %v", currentTenantID, err)
					return
				}

				logger.Infof("✅ Outbox scheduler started successfully for tenant: %d (index: %d, delayed by %dms)",
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

	// ✅ 注销所有租户的 ACK Channel（单租户和多租户模式都需要）
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
func (m *OutboxSchedulerManager) createScheduler(tenantID int, db *gorm.DB) (*outbox.OutboxScheduler, *outbox.OutboxPublisher) {
	// 创建租户专属的 Repository
	repo := gormadapter.NewGormOutboxRepository(db)

	// 创建调度器配置
	schedulerConfig := &outbox.SchedulerConfig{
		PollInterval:        1 * time.Second,  // 轮询间隔 1 秒（jxt-core 最小值）
		BatchSize:           100,              // 每次处理 100 个事件
		TenantID:            tenantID,         // 租户 ID（int 类型）
		CleanupInterval:     1 * time.Hour,    // 清理间隔 1 小时
		CleanupRetention:    24 * time.Hour,   // 保留 24 小时
		HealthCheckInterval: 30 * time.Second, // 健康检查间隔 30 秒
		EnableHealthCheck:   false,            // 禁用健康检查
		EnableCleanup:       true,             // 启用自动清理
		EnableMetrics:       false,            // 禁用指标收集（jxt-core v1.1.20 有 bug）
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
