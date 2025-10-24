package command

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// ActivateWorkflowCommand 激活工作流命令
type ActivateWorkflowCommand struct {
	ID string
}

// ActivateWorkflowHandler 激活工作流处理器
type ActivateWorkflowHandler struct {
	repo workflow.WorkflowRepository
}

// NewActivateWorkflowHandler 创建处理器
func NewActivateWorkflowHandler(repo workflow.WorkflowRepository) *ActivateWorkflowHandler {
	return &ActivateWorkflowHandler{repo: repo}
}

// Handle 处理命令
func (h *ActivateWorkflowHandler) Handle(ctx context.Context, cmd *ActivateWorkflowCommand) error {
	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if wf == nil {
		return workflow.ErrWorkflowNotFound
	}

	// 激活工作流
	if err := wf.Activate(); err != nil {
		return err
	}

	// 保存更新
	return h.repo.Update(ctx, wf)
}
