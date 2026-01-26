package instance_aggregate

import (
	"encoding/json"
	"time"

	workflow_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	errors "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/models"
	"jxt-evidence-system/process-management/shared/common/status"
)

// WorkflowInstance 工作流实例
type WorkflowInstance struct {
	InstanceId   valueobject.InstanceID      `json:"instanceId" gorm:"primaryKey;column:id;type:uuid;comment:主键编码"`
	WorkflowID   valueobject.WorkflowID      `json:"workflowId" gorm:"column:workflow_id;type:uuid;comment:工作流编码"`
	InstanceNo   string                      `json:"instanceNo"`
	WorkflowNo   string                      `json:"workflowNo" gorm:"-"`
	WorkflowName string                      `json:"workflowName" gorm:"-"`
	Status       status.InstanceStatus       `json:"status"`
	Input        json.RawMessage             `gorm:"type:jsonb" json:"input"`
	Output       json.RawMessage             `gorm:"type:jsonb" json:"output"`
	ErrorMessage string                      `json:"errorMessage"`
	StartedAt    time.Time                   `json:"startedAt"`
	CompletedAt  *time.Time                  `json:"completedAt"`
	Workflow     workflow_aggregate.Workflow `json:"-" gorm:"foreignKey:workflow_id;references:id"`

	// 审计字段
	models.ControlBy
	models.ModelTime
}

// NewWorkflowInstance 创建新工作流实例
func NewWorkflowInstance(workflowID valueobject.WorkflowID, input json.RawMessage) *WorkflowInstance {
	now := time.Now()
	return &WorkflowInstance{
		InstanceId: valueobject.NewInstanceID(),
		InstanceNo: "instance-" + now.Format("20060102150405"),
		WorkflowID: workflowID,
		Status:     status.InstanceStatusRunning,
		Input:      input,
		StartedAt:  now,
		ModelTime: models.ModelTime{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
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
