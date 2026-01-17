package port

import (
	"context"

	"jxt-evidence-system/process-management/internal/application/command"
)

type InstanceService interface {
	DeleteInstance(ctx context.Context, cmd *command.DeleteInstanceCommand) error
	GetInstanceByID(ctx context.Context, id string) (*command.WorkflowInstanceDTO, error)
	ListInstancesByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*command.WorkflowInstanceDTO, error)
	ListAllInstances(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*command.WorkflowInstanceDTO, int, error)
	StartWorkflowInstance(ctx context.Context, cmd *command.StartWorkflowInstanceCommand) (string, error)
}
