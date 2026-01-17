package port

import (
	"context"

	"jxt-evidence-system/process-management/internal/application/command"
)

type WorkflowService interface {
	ActivateWorkflow(ctx context.Context, cmd *command.ActivateWorkflowCommand) error

	CreateWorkflow(ctx context.Context, cmd *command.CreateWorkflowCommand) (string, error)

	DeleteWorkflow(ctx context.Context, cmd *command.DeleteWorkflowCommand) error

	FreezeWorkflow(ctx context.Context, cmd *command.FreezeWorkflowCommand) error
	GetWorkflowByID(ctx context.Context, id string) (*command.WorkflowDTO, error)
	ListWorkflows(ctx context.Context, limit, offset int) ([]*command.WorkflowDTO, error)
	CountWorkflows(ctx context.Context) (int64, error)
	UpdateWorkflow(ctx context.Context, cmd *command.UpdateWorkflowCommand) error
}
