package command

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// UpdateWorkflowCommand 更新工作流命令
type UpdateWorkflowCommand struct {
	ID          string
	Name        string
	Description string
	Definition  string
}

// UpdateWorkflowHandler 更新工作流处理器
type UpdateWorkflowHandler struct {
	repo workflow.WorkflowRepository
}

// NewUpdateWorkflowHandler 创建处理器
func NewUpdateWorkflowHandler(repo workflow.WorkflowRepository) *UpdateWorkflowHandler {
	return &UpdateWorkflowHandler{repo: repo}
}

// Handle 处理命令
func (h *UpdateWorkflowHandler) Handle(ctx context.Context, cmd *UpdateWorkflowCommand) error {
	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if wf == nil {
		return workflow.ErrWorkflowNotFound
	}

	// 业务规则验证：只能更新草稿或冻结状态的工作流
	if wf.Status != workflow.StatusDraft && wf.Status != workflow.StatusFrozen {
		return workflow.ErrInvalidStatusTransition
	}

	// 更新字段
	wf.Name = cmd.Name
	wf.Description = cmd.Description
	wf.Definition = cmd.Definition

	// 保存更新
	return h.repo.Update(ctx, wf)
}
