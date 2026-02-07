package event_repository

import (
	"context"
	"jxt-evidence-system/process-management/shared/common/transaction"

	jxtevent "github.com/ChenBigdata421/jxt-core/sdk/pkg/domain/event"
	jxtoutbox "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox"
)

// 类型别名：使用 jxt-core 的类型
type (
	// OutboxEvent 使用 jxt-core 的 OutboxEvent
	OutboxEvent = jxtoutbox.OutboxEvent

	// OutboxEventStatus 使用 jxt-core 的 EventStatus
	OutboxEventStatus = jxtoutbox.EventStatus
)

// OutboxRepository outbox事件仓储接口
// 注意：已迁移到 jxt-core，此接口主要用于适配器模式
type OutboxRepository interface {
	// SaveInTx 在事务中保存outbox事件
	SaveInTx(ctx context.Context, tx transaction.Transaction, event jxtevent.EnterpriseEvent) error
	Save(ctx context.Context, event jxtevent.EnterpriseEvent) error
	// FindUnpublishedEvents 查找未发布的事件
	FindUnpublishedEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)

	// FindUnpublishedEventsByTenant 查找指定租户未发布的事件
	FindUnpublishedEventsByTenant(ctx context.Context, tenantID string, limit int) ([]*OutboxEvent, error)

	// FindUnpublishedEventsByTenantWithDelay 查找指定租户创建时间超过指定延迟的未发布事件
	// 用于调度器避让机制，防止与立即发布产生竞态
	FindUnpublishedEventsByTenantWithDelay(ctx context.Context, tenantID string, delaySeconds int, limit int) ([]*OutboxEvent, error)

	// FindUnpublishedEventsByAggregateType 查找指定聚合类型未发布的事件
	FindUnpublishedEventsByAggregateType(ctx context.Context, aggregateType string, limit int) ([]*OutboxEvent, error)

	// FindUnpublishedEventsByAggregateID 查找指定聚合根ID未发布的事件
	FindUnpublishedEventsByAggregateID(ctx context.Context, aggregateID string) ([]*OutboxEvent, error)

	// FindUnpublishedEventsByTenantAndAggregateID 查找指定租户和聚合根ID未发布的事件
	FindUnpublishedEventsByTenantAndAggregateID(ctx context.Context, tenantID string, aggregateID string) ([]*OutboxEvent, error)

	// FindUnpublishedEventsByEventIDs 查找指定事件ID列表的未发布事件
	FindUnpublishedEventsByEventIDs(ctx context.Context, eventIDs []string) ([]*OutboxEvent, error)

	// FindUnpublishedEventsByTenantAndEventIDs 查找指定租户和事件ID列表的未发布事件
	FindUnpublishedEventsByTenantAndEventIDs(ctx context.Context, tenantID string, eventIDs []string) ([]*OutboxEvent, error)

	// UpdateStatus 更新事件状态
	UpdateStatus(ctx context.Context, eventID string, status OutboxEventStatus, errorMsg string) error

	// IncrementRetry 增加重试次数
	IncrementRetry(ctx context.Context, eventID string, errorMsg string) error

	// MarkAsPublished 标记事件为已发布
	MarkAsPublished(ctx context.Context, eventID string) error

	// MarkAsMaxRetry 标记事件为超过最大重试次数
	MarkAsMaxRetry(ctx context.Context, eventID string, errorMsg string) error

	// CountUnpublishedEvents 统计未发布事件数量
	CountUnpublishedEvents(ctx context.Context) (int64, error)

	// CountUnpublishedEventsByTenant 统计指定租户未发布事件数量
	CountUnpublishedEventsByTenant(ctx context.Context, tenantID string) (int64, error)

	// FindByID 根据ID查找事件
	FindByID(ctx context.Context, eventID string) (*OutboxEvent, error)

	// Delete 删除事件（用于已成功发布的事件清理）
	Delete(ctx context.Context, eventID string) error

	// DeleteOldPublishedEvents 删除旧的已发布事件（用于定期清理）
	DeleteOldPublishedEvents(ctx context.Context, beforeTime int64) error

	// FindEventsForRetry 查找需要重试的事件
	FindEventsForRetry(ctx context.Context, maxRetries int, limit int) ([]*OutboxEvent, error)
}
