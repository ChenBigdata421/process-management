package workflow_aggregate

import (
	"time"

	"jxt-evidence-system/process-management/internal/domain/valueobject"
	errors "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/models"
	"jxt-evidence-system/process-management/shared/common/status"
)

// Workflow 工作流聚合根
type Workflow struct {
	WorkflowID  valueobject.WorkflowID `json:"workflowId" gorm:"primaryKey;column:id;type:uuid;comment:主键编码"`
	WorkflowNo  string                 `json:"workflowNo"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      status.WorkflowStatus  `json:"status"`
	Definition  string                 `gorm:"type:jsonb" json:"definition"`
	// 审计字段
	models.ControlBy
	models.ModelTime
}

// NewWorkflow 创建新工作流
func NewWorkflow(name, description, definition string, createBy int) *Workflow {
	return &Workflow{
		WorkflowID:  valueobject.NewWorkflowID(),
		WorkflowNo:  "workflow-" + time.Now().Format("20060102150405"),
		Name:        name,
		Description: description,
		Status:      status.StatusDraft,
		Definition:  definition,
		ControlBy: models.ControlBy{
			CreateBy: createBy,
			UpdateBy: createBy,
		},
		ModelTime: models.ModelTime{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// Activate 激活工作流
func (w *Workflow) Activate() error {
	if w.Status != status.StatusDraft && w.Status != status.StatusFrozen {
		return errors.ErrInvalidStatusTransition
	}
	w.Status = status.StatusActive
	w.UpdatedAt = time.Now()
	return nil
}

// Freeze 冻结工作流
func (w *Workflow) Freeze() error {
	if w.Status != status.StatusActive {
		return errors.ErrInvalidStatusTransition
	}
	w.Status = status.StatusFrozen
	w.UpdatedAt = time.Now()
	return nil
}

// Complete 完成工作流
func (w *Workflow) Complete() error {
	if w.Status != status.StatusActive {
		return errors.ErrInvalidStatusTransition
	}
	w.Status = status.StatusCompleted
	w.UpdatedAt = time.Now()
	return nil
}

// Fail 工作流失败
func (w *Workflow) Fail() error {
	if w.Status != status.StatusActive {
		return errors.ErrInvalidStatusTransition
	}
	w.Status = status.StatusFailed
	w.UpdatedAt = time.Now()
	return nil
}

// Cancel 取消工作流
func (w *Workflow) Cancel() error {
	if w.Status == status.StatusCompleted || w.Status == status.StatusFailed {
		return errors.ErrCannotCancelCompletedWorkflow
	}
	w.Status = status.StatusCancelled
	w.UpdatedAt = time.Now()
	return nil
}

// TableName 指定表名
func (w *Workflow) TableName() string {
	return "workflows"
}
