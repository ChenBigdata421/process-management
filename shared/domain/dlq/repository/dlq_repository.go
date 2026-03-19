package repository

import (
	"context"
	"jxt-evidence-system/process-management/shared/domain/dlq/entity"
	"time"
)

// DLQRepository 死信队列仓储接口
type DLQRepository interface {
	// Save 保存死信消息
	Save(ctx context.Context, message *entity.DeadLetterMessage) error

	// FindByID 根据ID查询死信消息
	FindByID(ctx context.Context, id int64) (*entity.DeadLetterMessage, error)

	// FindByEventIDAndHandler 根据事件ID和处理器名称查询
	FindByEventIDAndHandler(ctx context.Context, eventID, handlerName string) (*entity.DeadLetterMessage, error)

	// List 分页查询死信消息
	List(ctx context.Context, filter *DLQFilter) ([]*entity.DeadLetterMessage, int64, error)

	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, id int64, status entity.DLQStatus, resolvedBy string) error

	// IncrementRetryCount 增加重试次数
	IncrementRetryCount(ctx context.Context, id int64) error

	// Delete 删除死信消息
	Delete(ctx context.Context, id int64) error
}

// DLQFilter 死信队列查询过滤器
type DLQFilter struct {
	HandlerName string           // 处理器名称
	ErrorType   string           // 错误类型
	Status      entity.DLQStatus // 状态
	TenantID    string           // 租户ID
	StartTime   *time.Time       // 开始时间
	EndTime     *time.Time       // 结束时间
	PageNum     int              // 页码
	PageSize    int              // 每页大小
}
