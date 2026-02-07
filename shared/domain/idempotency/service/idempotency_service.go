package service

import (
	"context"
	"errors"
	"time"

	"jxt-evidence-system/process-management/shared/domain/idempotency/entity"
	"jxt-evidence-system/process-management/shared/domain/idempotency/repository"
)

// 默认 PROCESSING 超时时间
const defaultProcessingTimeout = 5 * time.Minute

// ErrDuplicateKey 唯一键冲突错误（需要由具体的持久化层实现定义）
var ErrDuplicateKey = errors.New("duplicate key violation")

// IdempotencyService 幂等性服务接口
type IdempotencyService interface {
	// IsProcessed 检查事件是否已处理（状态为 SUCCESS）
	IsProcessed(ctx context.Context, eventID, handlerName string) (bool, error)

	// TryMarkProcessing 尝试标记事件为处理中（原子操作）
	// 返回 true 表示成功占位，可以继续处理
	// 返回 false 表示已被其他实例占位，应跳过
	// 注意：会检测超时的 PROCESSING 记录并自动回收
	TryMarkProcessing(ctx context.Context, eventID, eventType, aggregateID, tenantID, handlerName string) (bool, error)

	// MarkSuccess 标记事件处理成功
	MarkSuccess(ctx context.Context, eventID, handlerName string) error

	// MarkFailed 标记事件处理失败
	MarkFailed(ctx context.Context, eventID, handlerName string, errMsg string) error

	// ResetFailed 重置失败的事件，允许重新处理（管理接口）
	// 返回重置的记录数
	ResetFailed(ctx context.Context, eventID, handlerName string) (int64, error)

	// RecoverTimeoutProcessing 回收超时的 PROCESSING 状态记录
	// timeout: 超时阈值，超过该时间的 PROCESSING 记录将被重置为 FAILED
	// 返回回收的记录数
	RecoverTimeoutProcessing(ctx context.Context, timeout time.Duration) (int64, error)
}

// idempotencyService 幂等性服务实现
type idempotencyService struct {
	repo              repository.EventProcessingRepository
	processingTimeout time.Duration
}

// NewIdempotencyService 创建幂等性服务
func NewIdempotencyService(repo repository.EventProcessingRepository) IdempotencyService {
	return &idempotencyService{
		repo:              repo,
		processingTimeout: defaultProcessingTimeout,
	}
}

// IsProcessed 检查事件是否已处理成功
func (s *idempotencyService) IsProcessed(ctx context.Context, eventID, handlerName string) (bool, error) {
	record, err := s.repo.FindByEventIDAndHandler(ctx, eventID, handlerName)
	if err != nil {
		return false, err
	}
	if record == nil {
		return false, nil
	}
	return record.IsProcessed(), nil
}

// TryMarkProcessing 尝试标记事件为处理中
func (s *idempotencyService) TryMarkProcessing(ctx context.Context, eventID, eventType, aggregateID, tenantID, handlerName string) (bool, error) {
	// 首先尝试创建新记录
	record := entity.NewEventProcessingRecord(eventID, eventType, aggregateID, tenantID, handlerName)
	err := s.repo.Create(ctx, record)

	if err == nil {
		// 创建成功，成功占位
		return true, nil
	}

	// 检查是否是唯一键冲突
	if !errors.Is(err, ErrDuplicateKey) {
		// 其他错误，直接返回
		return false, err
	}

	// 唯一键冲突，查询现有记录
	existingRecord, err := s.repo.FindByEventIDAndHandler(ctx, eventID, handlerName)
	if err != nil {
		return false, err
	}

	if existingRecord == nil {
		// 理论上不应该发生，但为安全起见处理
		return false, nil
	}

	// 根据现有记录状态判断
	switch existingRecord.Status {
	case entity.StatusSuccess:
		// 已成功处理，跳过
		return false, nil

	case entity.StatusFailed:
		// 已失败，需要通过 ResetFailed 接口重置后才能重试
		return false, nil

	case entity.StatusProcessing:
		// 检查是否超时
		if existingRecord.IsTimeout(s.processingTimeout) {
			// 超时，自动回收为 FAILED
			err := s.repo.UpdateStatus(ctx, eventID, handlerName, entity.StatusFailed, "PROCESSING 超时，自动回收")
			if err != nil {
				return false, err
			}
			// 删除记录后重新占位
			_, err = s.repo.DeleteByEventIDAndHandler(ctx, eventID, handlerName)
			if err != nil {
				return false, err
			}
			// 重新创建记录
			newRecord := entity.NewEventProcessingRecord(eventID, eventType, aggregateID, tenantID, handlerName)
			err = s.repo.Create(ctx, newRecord)
			if err != nil {
				if errors.Is(err, ErrDuplicateKey) {
					// 并发情况下被其他实例抢占
					return false, nil
				}
				return false, err
			}
			return true, nil
		}
		// 未超时，正在处理中
		return false, nil

	default:
		return false, nil
	}
}

// MarkSuccess 标记事件处理成功
func (s *idempotencyService) MarkSuccess(ctx context.Context, eventID, handlerName string) error {
	return s.repo.UpdateStatus(ctx, eventID, handlerName, entity.StatusSuccess, "")
}

// MarkFailed 标记事件处理失败
func (s *idempotencyService) MarkFailed(ctx context.Context, eventID, handlerName string, errMsg string) error {
	return s.repo.UpdateStatus(ctx, eventID, handlerName, entity.StatusFailed, errMsg)
}

// ResetFailed 重置失败事件
func (s *idempotencyService) ResetFailed(ctx context.Context, eventID, handlerName string) (int64, error) {
	// 先查询确认是 FAILED 状态
	record, err := s.repo.FindByEventIDAndHandler(ctx, eventID, handlerName)
	if err != nil {
		return 0, err
	}
	if record == nil {
		return 0, nil
	}
	if record.Status != entity.StatusFailed {
		return 0, errors.New("only FAILED status can be reset")
	}

	return s.repo.DeleteByEventIDAndHandler(ctx, eventID, handlerName)
}

// RecoverTimeoutProcessing 回收超时的 PROCESSING 状态记录
func (s *idempotencyService) RecoverTimeoutProcessing(ctx context.Context, timeout time.Duration) (int64, error) {
	return s.repo.UpdateTimeoutProcessingToFailed(ctx, timeout)
}
