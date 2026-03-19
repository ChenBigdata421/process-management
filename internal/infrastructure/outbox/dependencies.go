package outbox

import (
	"jxt-evidence-system/process-management/shared/common/di"
	event_repository "jxt-evidence-system/process-management/shared/domain/event/repository"
	"sync"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox"
	gormadapter "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters/gorm"
)

var (
	registrations = make([]func(), 0)
	registerOnce  sync.Once
)

func init() {
	registrations = append(registrations, registerOutboxDependencies)
}

// RegisterDependencies 注册所有 outbox 相关依赖
func RegisterDependencies() {
	registerOnce.Do(func() {
		for _, f := range registrations {
			f()
		}
	})
}

// registerOutboxDependencies 注册 outbox 依赖
func registerOutboxDependencies() {
	// 1. 注册 evidence-management 的 OutboxRepository（使用适配器）
	// 这个给业务服务（MediaService 等）使用
	if err := di.Provide(func() event_repository.OutboxRepository {
		db := sdk.Runtime.GetTenantDB(1) // 单租户模式（默认租户 ID: 1）
		return NewOutboxRepositoryAdapter(db)
	}); err != nil {
		logger.Fatalf("Failed to provide OutboxRepository: %v", err)
	}

	// 2. 注册 jxt-core 的 OutboxRepository（给 Scheduler 使用）
	if err := di.Provide(func() outbox.OutboxRepository {
		db := sdk.Runtime.GetTenantDB(1) // 单租户模式（默认租户 ID: 1）
		return gormadapter.NewGormOutboxRepository(db)
	}); err != nil {
		logger.Fatalf("Failed to provide jxt-core OutboxRepository: %v", err)
	}

	// 3. 注册 EventPublisher（使用 jxt-core EventBusAdapter）
	if err := di.Provide(func() outbox.EventPublisher {
		return NewEventBusAdapter()
	}); err != nil {
		logger.Fatalf("Failed to provide EventPublisher: %v", err)
	}

	// 4. 注册 TopicMapper
	if err := di.Provide(func() outbox.TopicMapper {
		return NewTopicMapper()
	}); err != nil {
		logger.Fatalf("Failed to provide TopicMapper: %v", err)
	}

	// 5. 注册 OutboxSchedulerManager（管理单个或多个 Scheduler）
	if err := di.Provide(func(
		publisher outbox.EventPublisher,
		mapper outbox.TopicMapper,
	) *OutboxSchedulerManager {
		return NewOutboxSchedulerManager(publisher, mapper)
	}); err != nil {
		logger.Fatalf("Failed to provide OutboxSchedulerManager: %v", err)
	}

	logger.Info("Outbox dependencies registered successfully")
}
