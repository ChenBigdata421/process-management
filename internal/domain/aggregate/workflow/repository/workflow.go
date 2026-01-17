package repository

import (
	"context"
	workflow "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
)

// WorkflowRepository 工作流仓储接口"context"

// WorkflowRepository 工作流仓储接口
type WorkflowRepository interface {
	Save(ctx context.Context, workflow *workflow.Workflow) error
	FindByID(ctx context.Context, id string) (*workflow.Workflow, error)
	FindAll(ctx context.Context, limit, offset int) ([]*workflow.Workflow, error)
	Update(ctx context.Context, workflow *workflow.Workflow) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}
