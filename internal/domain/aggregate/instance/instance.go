package instance_aggregate

import (
	"encoding/json"
	"time"

	errors "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/status"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowInstance 工作流实例
type WorkflowInstance struct {
	ID           string                `gorm:"primaryKey" json:"id"`
	WorkflowID   string                `json:"workflow_id"`
	Status       status.InstanceStatus `json:"status"`
	Input        json.RawMessage       `gorm:"type:jsonb" json:"input"`
	Output       json.RawMessage       `gorm:"type:jsonb" json:"output"`
	ErrorMessage string                `json:"error_message"`
	StartedAt    time.Time             `json:"started_at"`
	CompletedAt  *time.Time            `json:"completed_at"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	DeletedAt    gorm.DeletedAt        `gorm:"index" json:"-"`
}

// NewWorkflowInstance 创建新工作流实例
func NewWorkflowInstance(workflowID, input string) *WorkflowInstance {
	now := time.Now()
	return &WorkflowInstance{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     status.InstanceStatusRunning,
		Input:      json.RawMessage(input),
		StartedAt:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Complete 完成实例
func (wi *WorkflowInstance) Complete(output string) error {
	if wi.Status != status.InstanceStatusRunning {
		return errors.ErrInvalidInstanceStatusTransition
	}
	now := time.Now()
	wi.Status = status.InstanceStatusCompleted
	wi.Output = json.RawMessage(output)
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// Fail 实例失败
func (wi *WorkflowInstance) Fail(errorMsg string) error {
	if wi.Status != status.InstanceStatusRunning {
		return errors.ErrInvalidInstanceStatusTransition
	}
	now := time.Now()
	wi.Status = status.InstanceStatusFailed
	wi.ErrorMessage = errorMsg
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// Cancel 取消实例
func (wi *WorkflowInstance) Cancel() error {
	if wi.Status != status.InstanceStatusRunning {
		return errors.ErrCannotCancelCompletedInstance
	}
	now := time.Now()
	wi.Status = status.InstanceStatusCancelled
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// TableName 指定表名
func (wi *WorkflowInstance) TableName() string {
	return "workflow_instances"
}
