package port

import (
	"context"

	"jxt-evidence-system/process-management/internal/application/command"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

type InstanceService interface {
	DeleteInstance(ctx context.Context, cmd *command.DeleteInstanceCommand) error
	CancelInstance(ctx context.Context, cmd *command.CancelInstanceCommand) error
	GetInstanceByID(ctx context.Context, id valueobject.InstanceID) (*instance_aggregate.WorkflowInstance, error)
	GetInstanceDetailByID(ctx context.Context, id valueobject.InstanceID) ([]command.TaskHistoryItem, error)
	GetInstancesByWorkflow(ctx context.Context, query *command.GetInstancesByWorkflowPagedQuery) ([]*instance_aggregate.WorkflowInstance, int, error)
	GetPage(ctx context.Context, query *command.InstancePagedQuery) ([]*instance_aggregate.WorkflowInstance, int, error)
	StartWorkflowInstance(ctx context.Context, cmd *command.StartWorkflowInstanceCommand) (string, error)
	CountInstanceByWorkflow(ctx context.Context, workflowID valueobject.WorkflowID) (int64, error)
}
