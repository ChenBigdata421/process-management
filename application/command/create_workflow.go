package command

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// CreateWorkflowCommand 创建工作流命令
type CreateWorkflowCommand struct {
	Name        string
	Description string
	Definition  string
}

// CreateWorkflowHandler 创建工作流处理器
type CreateWorkflowHandler struct {
	repo workflow.WorkflowRepository
}

// NewCreateWorkflowHandler 创建处理器
func NewCreateWorkflowHandler(repo workflow.WorkflowRepository) *CreateWorkflowHandler {
	return &CreateWorkflowHandler{repo: repo}
}

// Handle 处理命令
func (h *CreateWorkflowHandler) Handle(ctx context.Context, cmd *CreateWorkflowCommand) (string, error) {
	// 业务规则验证
	if cmd.Name == "" {
		return "", workflow.ErrInvalidWorkflowDefinition
	}

	// 创建工作流
	wf := workflow.NewWorkflow(cmd.Name, cmd.Description, cmd.Definition)

	// 保存到仓储
	if err := h.repo.Save(ctx, wf); err != nil {
		return "", err
	}

	return wf.ID, nil
}

