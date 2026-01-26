package repository

import (
	"context"
	"jxt-evidence-system/process-management/internal/application/command"
	workflow "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// WorkflowRepository 工作流仓储接口
type WorkflowRepository interface {
	Save(ctx context.Context, workflow *workflow.Workflow) error
	FindByID(ctx context.Context, id valueobject.WorkflowID) (*workflow.Workflow, error)
	FindByName(ctx context.Context, name string) (*workflow.Workflow, error)
	GetPage(ctx context.Context, query *command.WorkflowPagedQuery) ([]*workflow.Workflow, int, error)
	GetAllWorkflow(ctx context.Context) ([]*workflow.Workflow, error)
	Update(ctx context.Context, workflow *workflow.Workflow) error
	Delete(ctx context.Context, id valueobject.WorkflowID) error
	Count(ctx context.Context) (int64, error)
}
