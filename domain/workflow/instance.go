package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InstanceStatus 工作流实例状态
type InstanceStatus string

const (
	InstanceStatusRunning   InstanceStatus = "running"
	InstanceStatusCompleted InstanceStatus = "completed"
	InstanceStatusFailed    InstanceStatus = "failed"
	InstanceStatusCancelled InstanceStatus = "cancelled"
)

// WorkflowInstance 工作流实例
type WorkflowInstance struct {
	ID           string          `gorm:"primaryKey" json:"id"`
	WorkflowID   string          `json:"workflow_id"`
	Status       InstanceStatus  `json:"status"`
	Input        json.RawMessage `gorm:"type:jsonb" json:"input"`
	Output       json.RawMessage `gorm:"type:jsonb" json:"output"`
	ErrorMessage string          `json:"error_message"`
	StartedAt    time.Time       `json:"started_at"`
	CompletedAt  *time.Time      `json:"completed_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
}

// NewWorkflowInstance 创建新工作流实例
func NewWorkflowInstance(workflowID, input string) *WorkflowInstance {
	now := time.Now()
	return &WorkflowInstance{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     InstanceStatusRunning,
		Input:      json.RawMessage(input),
		StartedAt:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Complete 完成实例
func (wi *WorkflowInstance) Complete(output string) error {
	if wi.Status != InstanceStatusRunning {
		return ErrInvalidInstanceStatusTransition
	}
	now := time.Now()
	wi.Status = InstanceStatusCompleted
	wi.Output = json.RawMessage(output)
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// Fail 实例失败
func (wi *WorkflowInstance) Fail(errorMsg string) error {
	if wi.Status != InstanceStatusRunning {
		return ErrInvalidInstanceStatusTransition
	}
	now := time.Now()
	wi.Status = InstanceStatusFailed
	wi.ErrorMessage = errorMsg
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// Cancel 取消实例
func (wi *WorkflowInstance) Cancel() error {
	if wi.Status != InstanceStatusRunning {
		return ErrCannotCancelCompletedInstance
	}
	now := time.Now()
	wi.Status = InstanceStatusCancelled
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// TableName 指定表名
func (wi *WorkflowInstance) TableName() string {
	return "workflow_instances"
}
