package command

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// DeleteWorkflowCommand 删除工作流命令
type DeleteWorkflowCommand struct {
	ID string
}

// DeleteWorkflowHandler 删除工作流处理器
type DeleteWorkflowHandler struct {
	repo workflow.WorkflowRepository
}

// NewDeleteWorkflowHandler 创建处理器
func NewDeleteWorkflowHandler(repo workflow.WorkflowRepository) *DeleteWorkflowHandler {
	return &DeleteWorkflowHandler{repo: repo}
}

// Handle 处理命令
func (h *DeleteWorkflowHandler) Handle(ctx context.Context, cmd *DeleteWorkflowCommand) error {
	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if wf == nil {
		return workflow.ErrWorkflowNotFound
	}

	// 业务规则验证：只能删除草稿或已取消的工作流
	if wf.Status != workflow.StatusDraft && wf.Status != workflow.StatusCancelled {
		return workflow.ErrInvalidStatusTransition
	}

	// 执行删除
	return h.repo.Delete(ctx, cmd.ID)
}

