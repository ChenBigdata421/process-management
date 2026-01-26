package repository

import (
	"context"
	"jxt-evidence-system/process-management/internal/application/command"
	instance "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// WorkflowInstanceRepository 工作流实例仓储接口
type WorkflowInstanceRepository interface {
	Save(ctx context.Context, instance *instance.WorkflowInstance) error
	FindByID(ctx context.Context, id valueobject.InstanceID) (*instance.WorkflowInstance, error)
	FindByWorkflowID(ctx context.Context, query *command.GetInstancesByWorkflowPagedQuery) ([]*instance.WorkflowInstance, int, error)
	GetPage(ctx context.Context, query *command.InstancePagedQuery) ([]*instance.WorkflowInstance, int, error)
	Update(ctx context.Context, instance *instance.WorkflowInstance) error
	Delete(ctx context.Context, id valueobject.InstanceID) error
	CountByWorkflowID(ctx context.Context, workflowID valueobject.WorkflowID) (int64, error)
}
