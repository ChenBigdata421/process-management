package repository

import (
	"context"
	instance "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
)

// WorkflowInstanceRepository 工作流实例仓储接口
type WorkflowInstanceRepository interface {
	Save(ctx context.Context, instance *instance.WorkflowInstance) error
	FindByID(ctx context.Context, id string) (*instance.WorkflowInstance, error)
	FindByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*instance.WorkflowInstance, error)
	FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*instance.WorkflowInstance, int, error)
	Update(ctx context.Context, instance *instance.WorkflowInstance) error
	Delete(ctx context.Context, id string) error
	CountByWorkflowID(ctx context.Context, workflowID string) (int64, error)
}
