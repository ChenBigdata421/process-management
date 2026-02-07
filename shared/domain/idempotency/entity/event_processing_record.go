// Package entity 定义幂等性相关的领域实体
package entity

import (
	"time"
)

// ProcessingStatus 处理状态
type ProcessingStatus string

const (
	// StatusProcessing 处理中
	StatusProcessing ProcessingStatus = "PROCESSING"
	// StatusSuccess 处理成功
	StatusSuccess ProcessingStatus = "SUCCESS"
	// StatusFailed 处理失败
	StatusFailed ProcessingStatus = "FAILED"
)

// String 返回状态字符串
func (s ProcessingStatus) String() string {
	return string(s)
}

// IsValid 检查状态是否有效
func (s ProcessingStatus) IsValid() bool {
	switch s {
	case StatusProcessing, StatusSuccess, StatusFailed:
		return true
	default:
		return false
	}
}

// EventProcessingRecord 事件处理记录实体
//
// 用于记录事件的处理状态，实现幂等性保障：
// - 同一事件只处理一次
// - 支持事件处理状态追踪
// - 支持失败事件重试
type EventProcessingRecord struct {
	ID           int64
	EventID      string
	EventType    string
	AggregateID  string
	TenantID     string
	HandlerName  string
	Status       ProcessingStatus
	ErrorMessage string
	ProcessedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewEventProcessingRecord 创建新的事件处理记录
func NewEventProcessingRecord(eventID, eventType, aggregateID, tenantID, handlerName string) *EventProcessingRecord {
	now := time.Now()
	return &EventProcessingRecord{
		EventID:     eventID,
		EventType:   eventType,
		AggregateID: aggregateID,
		TenantID:    tenantID,
		HandlerName: handlerName,
		Status:      StatusProcessing,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsProcessed 检查是否已处理成功
func (r *EventProcessingRecord) IsProcessed() bool {
	return r.Status == StatusSuccess
}

// IsProcessing 检查是否正在处理中
func (r *EventProcessingRecord) IsProcessing() bool {
	return r.Status == StatusProcessing
}

// IsFailed 检查是否处理失败
func (r *EventProcessingRecord) IsFailed() bool {
	return r.Status == StatusFailed
}

// IsTimeout 检查是否超时（基于 CreatedAt 判断）
func (r *EventProcessingRecord) IsTimeout(timeout time.Duration) bool {
	return r.Status == StatusProcessing && time.Since(r.CreatedAt) > timeout
}

// MarkSuccess 标记为处理成功
func (r *EventProcessingRecord) MarkSuccess() {
	now := time.Now()
	r.Status = StatusSuccess
	r.ProcessedAt = &now
	r.UpdatedAt = now
}

// MarkFailed 标记为处理失败
func (r *EventProcessingRecord) MarkFailed(errMsg string) {
	now := time.Now()
	r.Status = StatusFailed
	r.ErrorMessage = errMsg
	r.ProcessedAt = &now
	r.UpdatedAt = now
}

// TableName 返回表名（供 GORM 使用）
func (EventProcessingRecord) TableName() string {
	return "event_processing_records"
}
