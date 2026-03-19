package entity

import "time"

// DeadLetterMessage 死信消息实体
// Command 和 Query 共用同一个实体定义，确保表结构一致
type DeadLetterMessage struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID      string     `gorm:"column:event_id;type:varchar(100);not null;index:idx_event_id" json:"event_id"`
	Topic        string     `gorm:"column:topic;type:varchar(100);not null;index:idx_topic" json:"topic"`
	EventType    string     `gorm:"column:event_type;type:varchar(50)" json:"event_type"`
	AggregateID  string     `gorm:"column:aggregate_id;type:varchar(100)" json:"aggregate_id"`
	TenantID     int        `gorm:"column:tenant_id;type:int;index:idx_tenant_id" json:"tenant_id"`
	HandlerName  string     `gorm:"column:handler_name;type:varchar(100);not null;index:idx_handler_name" json:"handler_name"`
	Payload      []byte     `gorm:"column:payload;type:text;not null" json:"payload"`
	ErrorMessage string     `gorm:"column:error_message;type:text" json:"error_message"`
	ErrorType    string     `gorm:"column:error_type;type:varchar(20);index:idx_error_type" json:"error_type"`
	RetryCount   int        `gorm:"column:retry_count;default:0" json:"retry_count"`
	Status       string     `gorm:"column:status;type:varchar(20);default:'PENDING';index:idx_status" json:"status"`
	ResolvedAt   *time.Time `gorm:"column:resolved_at" json:"resolved_at,omitempty"`
	ResolvedBy   string     `gorm:"column:resolved_by;type:varchar(100)" json:"resolved_by,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime;index:idx_created_at" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DeadLetterMessage) TableName() string {
	return "dead_letter_queue"
}

// DLQStatus 死信队列状态
type DLQStatus string

const (
	DLQStatusPending  DLQStatus = "PENDING"  // 待处理
	DLQStatusRetrying DLQStatus = "RETRYING" // 重试中
	DLQStatusResolved DLQStatus = "RESOLVED" // 已解决
	DLQStatusFailed   DLQStatus = "FAILED"   // 最终失败
)
