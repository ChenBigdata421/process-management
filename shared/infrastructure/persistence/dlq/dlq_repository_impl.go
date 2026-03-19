package dlq

import (
	"context"
	"errors"
	"fmt"
	"jxt-evidence-system/process-management/shared/common/global"
	"jxt-evidence-system/process-management/shared/domain/dlq/entity"
	"jxt-evidence-system/process-management/shared/domain/dlq/repository"
	"time"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"gorm.io/gorm"
)

// GormDLQRepository GORM 实现的死信队列仓储（延迟获取数据库连接）
type GormDLQRepository struct{}

// NewGormDLQRepository 创建 GORM DLQ 仓储实例（数据库连接在使用时动态获取）
func NewGormDLQRepository() repository.DLQRepository {
	return &GormDLQRepository{}
}

// getDB 从上下文获取租户ID并返回对应的数据库连接
func (r *GormDLQRepository) getDB(ctx context.Context) (*gorm.DB, error) {
	tenantID, ok := ctx.Value(global.TenantIDKey).(int)
	if !ok || tenantID <= 0 {
		return nil, errors.New("tenant id not exist in context")
	}
	db := sdk.Runtime.GetTenantDB(tenantID)
	if db == nil {
		return nil, fmt.Errorf("database not found for tenant: %d", tenantID)
	}
	return db.WithContext(ctx), nil
}

// Save 保存死信消息
func (r *GormDLQRepository) Save(ctx context.Context, message *entity.DeadLetterMessage) error {
	db, err := r.getDB(ctx)
	if err != nil {
		return err
	}
	return db.Create(message).Error
}

// FindByID 根据ID查询死信消息
func (r *GormDLQRepository) FindByID(ctx context.Context, id int64) (*entity.DeadLetterMessage, error) {
	db, err := r.getDB(ctx)
	if err != nil {
		return nil, err
	}
	var message entity.DeadLetterMessage
	err = db.Where("id = ?", id).First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("死信消息不存在: id=%d", id)
		}
		return nil, err
	}
	return &message, nil
}

// FindByEventIDAndHandler 根据事件ID和处理器名称查询
func (r *GormDLQRepository) FindByEventIDAndHandler(ctx context.Context, eventID, handlerName string) (*entity.DeadLetterMessage, error) {
	db, err := r.getDB(ctx)
	if err != nil {
		return nil, err
	}
	var message entity.DeadLetterMessage
	err = db.
		Where("event_id = ? AND handler_name = ?", eventID, handlerName).
		First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 不存在时返回 nil，不报错
		}
		return nil, err
	}
	return &message, nil
}

// List 分页查询死信消息
func (r *GormDLQRepository) List(ctx context.Context, filter *repository.DLQFilter) ([]*entity.DeadLetterMessage, int64, error) {
	db, err := r.getDB(ctx)
	if err != nil {
		return nil, 0, err
	}
	var messages []*entity.DeadLetterMessage
	var total int64

	query := db.Model(&entity.DeadLetterMessage{})

	// 应用过滤条件
	if filter.HandlerName != "" {
		query = query.Where("handler_name = ?", filter.HandlerName)
	}
	if filter.ErrorType != "" {
		query = query.Where("error_type = ?", filter.ErrorType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.TenantID != "" {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if filter.PageNum > 0 && filter.PageSize > 0 {
		offset := (filter.PageNum - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// 按创建时间倒序
	query = query.Order("created_at DESC")

	if err := query.Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// UpdateStatus 更新状态
func (r *GormDLQRepository) UpdateStatus(ctx context.Context, id int64, status entity.DLQStatus, resolvedBy string) error {
	db, err := r.getDB(ctx)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == entity.DLQStatusResolved {
		now := time.Now()
		updates["resolved_at"] = &now
		updates["resolved_by"] = resolvedBy
	}

	return db.
		Model(&entity.DeadLetterMessage{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// IncrementRetryCount 增加重试次数
func (r *GormDLQRepository) IncrementRetryCount(ctx context.Context, id int64) error {
	db, err := r.getDB(ctx)
	if err != nil {
		return err
	}
	return db.
		Model(&entity.DeadLetterMessage{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + ?", 1)).Error
}

// Delete 删除死信消息
func (r *GormDLQRepository) Delete(ctx context.Context, id int64) error {
	db, err := r.getDB(ctx)
	if err != nil {
		return err
	}
	return db.
		Where("id = ?", id).
		Delete(&entity.DeadLetterMessage{}).Error
}
