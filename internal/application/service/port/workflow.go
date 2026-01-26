package port

import (
	"context"

	"jxt-evidence-system/process-management/internal/application/command"
	workflow_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

type WorkflowService interface {
	ActivateWorkflow(ctx context.Context, cmd *command.ActivateWorkflowCommand) error

	CreateWorkflow(ctx context.Context, cmd *command.CreateWorkflowCommand) (string, error)

	DeleteWorkflow(ctx context.Context, cmd *command.DeleteWorkflowCommand) error

	FreezeWorkflow(ctx context.Context, cmd *command.FreezeWorkflowCommand) error
	GetWorkflowByID(ctx context.Context, id valueobject.WorkflowID) (*workflow_aggregate.Workflow, error)
	GetWorkflowByName(ctx context.Context, Name string) (*workflow_aggregate.Workflow, error)
	GetPage(ctx context.Context, query *command.WorkflowPagedQuery) ([]*workflow_aggregate.Workflow, int, error)
	GetAllWorkflow(ctx context.Context) ([]*workflow_aggregate.Workflow, error)
	CountWorkflows(ctx context.Context) (int64, error)
	UpdateWorkflow(ctx context.Context, cmd *command.UpdateWorkflowCommand) error
}
