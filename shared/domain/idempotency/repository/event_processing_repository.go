// Package repository 定义幂等性相关的仓储接口
package repository

import (
	"context"
	"time"

	"jxt-evidence-system/process-management/shared/domain/idempotency/entity"
)

// EventProcessingRepository 事件处理记录仓储接口
//
// 提供事件处理记录的持久化操作，用于实现幂等性保障
type EventProcessingRepository interface {
	// FindByEventIDAndHandler 根据事件ID和处理器名称查询记录
	// 返回 nil, nil 表示记录不存在
	FindByEventIDAndHandler(ctx context.Context, eventID, handlerName string) (*entity.EventProcessingRecord, error)

	// Create 创建记录
	// 如果记录已存在（唯一键冲突），返回 ErrDuplicateKey 错误
	Create(ctx context.Context, record *entity.EventProcessingRecord) error

	// UpdateStatus 更新记录状态
	UpdateStatus(ctx context.Context, eventID, handlerName string, status entity.ProcessingStatus, errMsg string) error

	// DeleteByEventIDAndHandler 删除记录（用于重置失败事件）
	// 返回删除的记录数
	DeleteByEventIDAndHandler(ctx context.Context, eventID, handlerName string) (int64, error)

	// UpdateTimeoutProcessingToFailed 将超时的 PROCESSING 状态更新为 FAILED
	// timeout: 超时阈值，超过该时间的 PROCESSING 记录将被更新为 FAILED
	// 返回更新的记录数
	UpdateTimeoutProcessingToFailed(ctx context.Context, timeout time.Duration) (int64, error)
}
