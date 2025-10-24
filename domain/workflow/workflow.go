package workflow

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	StatusDraft     WorkflowStatus = "draft"
	StatusActive    WorkflowStatus = "active"
	StatusFrozen    WorkflowStatus = "frozen"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusCancelled WorkflowStatus = "cancelled"
)

// Workflow 工作流聚合根
type Workflow struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      WorkflowStatus `json:"status"`
	Definition  string         `gorm:"type:jsonb" json:"definition"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// NewWorkflow 创建新工作流
func NewWorkflow(name, description, definition string) *Workflow {
	return &Workflow{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Status:      StatusDraft,
		Definition:  definition,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Activate 激活工作流
func (w *Workflow) Activate() error {
	if w.Status != StatusDraft && w.Status != StatusFrozen {
		return ErrInvalidStatusTransition
	}
	w.Status = StatusActive
	w.UpdatedAt = time.Now()
	return nil
}

// Freeze 冻结工作流
func (w *Workflow) Freeze() error {
	if w.Status != StatusActive {
		return ErrInvalidStatusTransition
	}
	w.Status = StatusFrozen
	w.UpdatedAt = time.Now()
	return nil
}

// Complete 完成工作流
func (w *Workflow) Complete() error {
	if w.Status != StatusActive {
		return ErrInvalidStatusTransition
	}
	w.Status = StatusCompleted
	w.UpdatedAt = time.Now()
	return nil
}

// Fail 工作流失败
func (w *Workflow) Fail() error {
	if w.Status != StatusActive {
		return ErrInvalidStatusTransition
	}
	w.Status = StatusFailed
	w.UpdatedAt = time.Now()
	return nil
}

// Cancel 取消工作流
func (w *Workflow) Cancel() error {
	if w.Status == StatusCompleted || w.Status == StatusFailed {
		return ErrCannotCancelCompletedWorkflow
	}
	w.Status = StatusCancelled
	w.UpdatedAt = time.Now()
	return nil
}

// TableName 指定表名
func (w *Workflow) TableName() string {
	return "workflows"
}
