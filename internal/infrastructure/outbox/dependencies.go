package outbox

import (
	"jxt-evidence-system/process-management/shared/common/di"
	event_repository "jxt-evidence-system/process-management/shared/domain/event/repository"
	"sync"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox"
)

var (
	registrations = make([]func(), 0)
	registerOnce  = sync.Once{}
)

// jiyuanjie 负责依赖注入所有的 application service 实例
func RegisterDependencies() {
	//遍历所有application service的依赖注入方法
	registerOnce.Do(func() {
		for _, f := range registrations {
			f()
		}
	})
}

func init() {
	registrations = append(registrations,
		registerOutboxRepositoryDependency,
		registerEventPublisherDependency,
		registerTopicMapperDependency,
		registerOutboxSchedulerManagerDependency,
	)
}

// registerOutboxRepositoryDependency 注册 OutboxRepository 依赖
func registerOutboxRepositoryDependency() {
	if err := di.Provide(func() event_repository.OutboxRepository {
		db := sdk.Runtime.GetTenantDB("*") // 默认使用 "*"，业务代码应该从上下文获取正确的 DB
		return NewOutboxRepositoryAdapter(db)
	}); err != nil {
		logger.Fatalf("Failed to provide OutboxRepository: %v", err)
	}
	logger.Info("OutboxRepository dependency registered successfully")
}

// registerEventPublisherDependency 注册 EventPublisher 依赖
func registerEventPublisherDependency() {
	if err := di.Provide(func() outbox.EventPublisher {
		return NewEventBusAdapter()
	}); err != nil {
		logger.Fatalf("Failed to provide EventPublisher: %v", err)
	}
	logger.Info("EventPublisher dependency registered successfully")
}

// registerTopicMapperDependency 注册 TopicMapper 依赖
func registerTopicMapperDependency() {
	if err := di.Provide(func() outbox.TopicMapper {
		return NewTopicMapper()
	}); err != nil {
		logger.Fatalf("Failed to provide TopicMapper: %v", err)
	}
	logger.Info("TopicMapper dependency registered successfully")
}

// registerOutboxSchedulerManagerDependency 注册 OutboxSchedulerManager 依赖
func registerOutboxSchedulerManagerDependency() {
	if err := di.Provide(func(
		publisher outbox.EventPublisher,
		mapper outbox.TopicMapper,
	) *OutboxSchedulerManager {
		return NewOutboxSchedulerManager(publisher, mapper)
	}); err != nil {
		logger.Fatalf("Failed to provide OutboxSchedulerManager: %v", err)
	}
	logger.Info("OutboxSchedulerManager dependency registered successfully")
}
