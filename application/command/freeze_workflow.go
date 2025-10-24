package command

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// FreezeWorkflowCommand 冻结工作流命令
type FreezeWorkflowCommand struct {
	ID string
}

// FreezeWorkflowHandler 冻结工作流处理器
type FreezeWorkflowHandler struct {
	repo workflow.WorkflowRepository
}

// NewFreezeWorkflowHandler 创建处理器
func NewFreezeWorkflowHandler(repo workflow.WorkflowRepository) *FreezeWorkflowHandler {
	return &FreezeWorkflowHandler{repo: repo}
}

// Handle 处理命令
func (h *FreezeWorkflowHandler) Handle(ctx context.Context, cmd *FreezeWorkflowCommand) error {
	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if wf == nil {
		return workflow.ErrWorkflowNotFound
	}

	// 冻结工作流
	if err := wf.Freeze(); err != nil {
		return err
	}

	// 保存更新
	return h.repo.Update(ctx, wf)
}

