package outbox

import (
	"context"
	"fmt"

	jxtoutbox "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox"
	gormadapter "github.com/ChenBigdata421/jxt-core/sdk/pkg/outbox/adapters/gorm"
	"gorm.io/gorm"

	persistence "jxt-evidence-system/process-management/internal/infrastructure/persistence/gorm"
	"jxt-evidence-system/process-management/shared/common/transaction"
	event_repository "jxt-evidence-system/process-management/shared/domain/event/repository"

	jxtevent "github.com/ChenBigdata421/jxt-core/sdk/pkg/domain/event"
)

// OutboxRepositoryAdapter 适配器，将 evidence-management 的 OutboxRepository 接口适配到 jxt-core 的 OutboxRepository
type OutboxRepositoryAdapter struct {
	jxtRepo *gormadapter.GormOutboxRepository
}

// NewOutboxRepositoryAdapter 创建 OutboxRepository 适配器
func NewOutboxRepositoryAdapter(db *gorm.DB) event_repository.OutboxRepository {
	jxtRepo := gormadapter.NewGormOutboxRepository(db).(*gormadapter.GormOutboxRepository)
	return &OutboxRepositoryAdapter{
		jxtRepo: jxtRepo,
	}
}

// convertEventToJxtOutboxEvent 将 evidence-management 的 Event 转换为 jxt-core 的 OutboxEvent
//
// 【优化方案】
//
// 设计原则：
// 1. OutboxEvent.Payload 存储完整的 DomainEvent（包含所有领域信息）
// 2. Envelope 字段从 OutboxEvent 字段获取（技术目的：ACK处理、顺序处理）
// 3. DomainEvent 信息从 Payload 获取（业务目的：领域事件处理）
// 4. 确保 Envelope 和 DomainEvent 信息强制一致
//
// 实现方式：
// - 传入完整的 DomainEvent 对象作为 payload
// - 由 jxt-core 负责序列化 DomainEvent 为 JSON（只序列化一次）
// - 覆盖关键字段确保一致性
// - 添加一致性验证
//
// 优点：
// - 避免双重序列化问题
// - Query端获得完整的领域信息
// - 确保 Envelope 和 DomainEvent 信息一致
// - 符合信息重复但用途不同的设计原则
func convertEventToJxtOutboxEvent(event jxtevent.EnterpriseEvent) (*jxtoutbox.OutboxEvent, error) {
	aggregateID := fmt.Sprintf("%v", event.GetAggregateID())

	// ✅ 直接传入完整的 DomainEvent 对象
	// jxt-core 会自动调用 jxtevent.MarshalDomainEvent() 进行序列化
	jxtEvent, err := jxtoutbox.NewOutboxEvent(
		event.GetTenantId(),
		aggregateID,
		event.GetAggregateType(),
		event.GetEventType(),
		event, // ✅ 传入完整的 DomainEvent 对象
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jxt outbox event: %w", err)
	}

	// ✅ 确保Envelope和DomainEvent信息一致
	// 覆盖jxt-core自动生成的字段，使用DomainEvent的原始值
	jxtEvent.ID = event.GetEventID()             // EventID一致性
	jxtEvent.CreatedAt = event.GetOccurredAt()   // 时间戳一致性
	jxtEvent.UpdatedAt = event.GetOccurredAt()   // 更新时间一致性
	jxtEvent.Version = int64(event.GetVersion()) // 版本一致性（jxt-core使用int64）

	// ✅ 验证关键字段一致性（防御性编程）
	if err := validateConsistency(jxtEvent, event); err != nil {
		return nil, fmt.Errorf("envelope-domainevent consistency check failed: %w", err)
	}

	// 生成幂等性键：{TenantID}:{AggregateType}:{AggregateID}:{EventType}:{EventID}
	jxtEvent.IdempotencyKey = fmt.Sprintf("%s:%s:%s:%s:%s",
		event.GetTenantId(),
		event.GetAggregateType(),
		aggregateID,
		event.GetEventType(),
		event.GetEventID(),
	)

	return jxtEvent, nil
}

// validateConsistency 验证Envelope和DomainEvent的关键信息一致性
func validateConsistency(jxtEvent *jxtoutbox.OutboxEvent, event jxtevent.EnterpriseEvent) error {
	// 验证TenantID一致性
	if jxtEvent.TenantID != event.GetTenantId() {
		return fmt.Errorf("tenantID mismatch: envelope=%s, domainEvent=%s",
			jxtEvent.TenantID, event.GetTenantId())
	}

	// 验证EventType一致性
	if jxtEvent.EventType != event.GetEventType() {
		return fmt.Errorf("eventType mismatch: envelope=%s, domainEvent=%s",
			jxtEvent.EventType, event.GetEventType())
	}

	// 验证AggregateID一致性
	expectedAggregateID := fmt.Sprintf("%v", event.GetAggregateID())
	if jxtEvent.AggregateID != expectedAggregateID {
		return fmt.Errorf("aggregateID mismatch: envelope=%s, domainEvent=%s",
			jxtEvent.AggregateID, expectedAggregateID)
	}

	// 验证AggregateType一致性
	if jxtEvent.AggregateType != event.GetAggregateType() {
		return fmt.Errorf("aggregateType mismatch: envelope=%s, domainEvent=%s",
			jxtEvent.AggregateType, event.GetAggregateType())
	}

	// 验证EventID一致性
	if jxtEvent.ID != event.GetEventID() {
		return fmt.Errorf("eventID mismatch: envelope=%s, domainEvent=%s",
			jxtEvent.ID, event.GetEventID())
	}

	return nil
}

// convertJxtOutboxEventToEvent 将 jxt-core 的 OutboxEvent 转换为 evidence-management 的 OutboxEvent
// 由于 event_repository.OutboxEvent 是 jxtoutbox.OutboxEvent 的类型别名，直接返回即可
func convertJxtOutboxEventToEvent(jxtEvent *jxtoutbox.OutboxEvent) *event_repository.OutboxEvent {
	return jxtEvent
}

// SaveInTx 在事务中保存事件
func (a *OutboxRepositoryAdapter) SaveInTx(ctx context.Context, tx transaction.Transaction, event jxtevent.EnterpriseEvent) error {
	// 转换为 jxt-core OutboxEvent
	jxtEvent, err := convertEventToJxtOutboxEvent(event)
	if err != nil {
		return err
	}

	// 从 transaction.Transaction 中提取 *gorm.DB
	gormTx := persistence.GetTx(tx)
	if gormTx == nil {
		return fmt.Errorf("invalid transaction type, cannot extract *gorm.DB")
	}

	// 使用 jxt-core 的 SaveInTx
	return a.jxtRepo.SaveInTx(ctx, gormTx, jxtEvent)
}

// Save 保存事件到 outbox
func (a *OutboxRepositoryAdapter) Save(ctx context.Context, event jxtevent.EnterpriseEvent) error {
	// 转换为 jxt-core OutboxEvent
	jxtEvent, err := convertEventToJxtOutboxEvent(event)
	if err != nil {
		return err
	}

	// 使用 jxt-core 的 Save
	return a.jxtRepo.Save(ctx, jxtEvent)
}

// FindUnpublishedEvents 查找未发布的事件
func (a *OutboxRepositoryAdapter) FindUnpublishedEvents(ctx context.Context, limit int) ([]*event_repository.OutboxEvent, error) {
	jxtEvents, err := a.jxtRepo.FindPendingEvents(ctx, limit, "")
	if err != nil {
		return nil, err
	}

	events := make([]*event_repository.OutboxEvent, len(jxtEvents))
	for i, jxtEvent := range jxtEvents {
		events[i] = convertJxtOutboxEventToEvent(jxtEvent)
	}

	return events, nil
}

// FindUnpublishedEventsByTenant 查找指定租户未发布的事件
func (a *OutboxRepositoryAdapter) FindUnpublishedEventsByTenant(ctx context.Context, tenantID string, limit int) ([]*event_repository.OutboxEvent, error) {
	jxtEvents, err := a.jxtRepo.FindPendingEvents(ctx, limit, tenantID)
	if err != nil {
		return nil, err
	}

	events := make([]*event_repository.OutboxEvent, len(jxtEvents))
	for i, jxtEvent := range jxtEvents {
		events[i] = convertJxtOutboxEventToEvent(jxtEvent)
	}

	return events, nil
}

// FindUnpublishedEventsByTenantWithDelay 查找指定租户创建时间超过指定延迟的未发布事件
func (a *OutboxRepositoryAdapter) FindUnpublishedEventsByTenantWithDelay(ctx context.Context, tenantID string, delaySeconds int, limit int) ([]*event_repository.OutboxEvent, error) {
	jxtEvents, err := a.jxtRepo.FindPendingEventsWithDelay(ctx, tenantID, delaySeconds, limit)
	if err != nil {
		return nil, err
	}

	events := make([]*event_repository.OutboxEvent, len(jxtEvents))
	for i, jxtEvent := range jxtEvents {
		events[i] = convertJxtOutboxEventToEvent(jxtEvent)
	}

	return events, nil
}

// FindUnpublishedEventsByAggregateType 根据聚合类型查找未发布事件
func (a *OutboxRepositoryAdapter) FindUnpublishedEventsByAggregateType(ctx context.Context, aggregateType string, limit int) ([]*event_repository.OutboxEvent, error) {
	jxtEvents, err := a.jxtRepo.FindByAggregateType(ctx, aggregateType, limit)
	if err != nil {
		return nil, err
	}

	events := make([]*event_repository.OutboxEvent, len(jxtEvents))
	for i, jxtEvent := range jxtEvents {
		events[i] = convertJxtOutboxEventToEvent(jxtEvent)
	}

	return events, nil
}

// FindUnpublishedEventsByAggregateID 根据聚合根ID查找未发布事件
func (a *OutboxRepositoryAdapter) FindUnpublishedEventsByAggregateID(ctx context.Context, aggregateID string) ([]*event_repository.OutboxEvent, error) {
	jxtEvents, err := a.jxtRepo.FindByAggregateID(ctx, aggregateID, "")
	if err != nil {
		return nil, err
	}

	events := make([]*event_repository.OutboxEvent, len(jxtEvents))
	for i, jxtEvent := range jxtEvents {
		events[i] = convertJxtOutboxEventToEvent(jxtEvent)
	}

	return events, nil
}

// FindUnpublishedEventsByTenantAndAggregateID 查找指定租户和聚合根ID的未发布事件
func (a *OutboxRepositoryAdapter) FindUnpublishedEventsByTenantAndAggregateID(ctx context.Context, tenantID string, aggregateID string) ([]*event_repository.OutboxEvent, error) {
	jxtEvents, err := a.jxtRepo.FindByAggregateID(ctx, aggregateID, tenantID)
	if err != nil {
		return nil, err
	}

	events := make([]*event_repository.OutboxEvent, len(jxtEvents))
	for i, jxtEvent := range jxtEvents {
		events[i] = convertJxtOutboxEventToEvent(jxtEvent)
	}

	return events, nil
}

// FindUnpublishedEventsByEventIDs 根据事件ID列表查找未发布事件
func (a *OutboxRepositoryAdapter) FindUnpublishedEventsByEventIDs(ctx context.Context, eventIDs []string) ([]*event_repository.OutboxEvent, error) {
	events := make([]*event_repository.OutboxEvent, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		jxtEvent, err := a.jxtRepo.FindByID(ctx, eventID)
		if err != nil {
			continue // 跳过未找到的事件
		}
		if jxtEvent.Status == jxtoutbox.EventStatusPending {
			events = append(events, convertJxtOutboxEventToEvent(jxtEvent))
		}
	}
	return events, nil
}

// FindUnpublishedEventsByTenantAndEventIDs 查找指定租户和事件ID列表的未发布事件
func (a *OutboxRepositoryAdapter) FindUnpublishedEventsByTenantAndEventIDs(ctx context.Context, tenantID string, eventIDs []string) ([]*event_repository.OutboxEvent, error) {
	events := make([]*event_repository.OutboxEvent, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		jxtEvent, err := a.jxtRepo.FindByID(ctx, eventID)
		if err != nil {
			continue // 跳过未找到的事件
		}
		if jxtEvent.TenantID == tenantID && jxtEvent.Status == jxtoutbox.EventStatusPending {
			events = append(events, convertJxtOutboxEventToEvent(jxtEvent))
		}
	}
	return events, nil
}

// UpdateStatus 更新事件状态
func (a *OutboxRepositoryAdapter) UpdateStatus(ctx context.Context, eventID string, status event_repository.OutboxEventStatus, errorMsg string) error {
	jxtEvent, err := a.jxtRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}

	jxtEvent.Status = status
	jxtEvent.LastError = errorMsg

	return a.jxtRepo.Update(ctx, jxtEvent)
}

// IncrementRetry 增加重试次数
func (a *OutboxRepositoryAdapter) IncrementRetry(ctx context.Context, eventID string, errorMsg string) error {
	return a.jxtRepo.IncrementRetry(ctx, eventID, errorMsg)
}

// MarkAsPublished 标记事件为已发布
func (a *OutboxRepositoryAdapter) MarkAsPublished(ctx context.Context, eventID string) error {
	return a.jxtRepo.MarkAsPublished(ctx, eventID)
}

// MarkAsFailed 标记事件为失败
func (a *OutboxRepositoryAdapter) MarkAsFailed(ctx context.Context, eventID string, errorMsg string) error {
	return a.jxtRepo.MarkAsFailed(ctx, eventID, fmt.Errorf("%s", errorMsg))
}

// MarkAsMaxRetry 标记事件为超过最大重试次数
func (a *OutboxRepositoryAdapter) MarkAsMaxRetry(ctx context.Context, eventID string, errorMsg string) error {
	return a.jxtRepo.MarkAsMaxRetry(ctx, eventID, errorMsg)
}

// CountUnpublishedEvents 统计未发布事件数量
func (a *OutboxRepositoryAdapter) CountUnpublishedEvents(ctx context.Context) (int64, error) {
	return a.jxtRepo.Count(ctx, jxtoutbox.EventStatusPending, "")
}

// CountUnpublishedEventsByTenant 统计指定租户未发布事件数量
func (a *OutboxRepositoryAdapter) CountUnpublishedEventsByTenant(ctx context.Context, tenantID string) (int64, error) {
	return a.jxtRepo.Count(ctx, jxtoutbox.EventStatusPending, tenantID)
}

// FindByID 根据ID查找事件
func (a *OutboxRepositoryAdapter) FindByID(ctx context.Context, eventID string) (*event_repository.OutboxEvent, error) {
	jxtEvent, err := a.jxtRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return convertJxtOutboxEventToEvent(jxtEvent), nil
}

// Delete 删除事件
func (a *OutboxRepositoryAdapter) Delete(ctx context.Context, eventID string) error {
	return a.jxtRepo.Delete(ctx, eventID)
}

// DeleteOldPublishedEvents 删除旧的已发布事件
func (a *OutboxRepositoryAdapter) DeleteOldPublishedEvents(ctx context.Context, beforeTime int64) error {
	// jxt-core 使用 time.Time，需要转换
	// beforeTime 是 Unix 时间戳（秒）
	// 这里暂时不实现，因为 jxt-core 的调度器会自动清理
	return nil
}

// FindEventsForRetry 查找需要重试的事件
func (a *OutboxRepositoryAdapter) FindEventsForRetry(ctx context.Context, maxRetries int, limit int) ([]*event_repository.OutboxEvent, error) {
	jxtEvents, err := a.jxtRepo.FindEventsForRetry(ctx, maxRetries, limit)
	if err != nil {
		return nil, err
	}

	events := make([]*event_repository.OutboxEvent, len(jxtEvents))
	for i, jxtEvent := range jxtEvents {
		events[i] = convertJxtOutboxEventToEvent(jxtEvent)
	}

	return events, nil
}
