package command

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// DeleteInstanceCommand 删除工作流实例命令
type DeleteInstanceCommand struct {
	ID string
}

// DeleteInstanceHandler 删除工作流实例处理器
type DeleteInstanceHandler struct {
	repo workflow.WorkflowInstanceRepository
}

// NewDeleteInstanceHandler 创建处理器
func NewDeleteInstanceHandler(repo workflow.WorkflowInstanceRepository) *DeleteInstanceHandler {
	return &DeleteInstanceHandler{repo: repo}
}

// Handle 处理命令
func (h *DeleteInstanceHandler) Handle(ctx context.Context, cmd *DeleteInstanceCommand) error {
	// 查找工作流实例
	instance, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if instance == nil {
		return workflow.ErrInstanceNotFound
	}

	// 业务规则验证：只能删除已完成、失败或已取消的实例
	if instance.Status == workflow.InstanceStatusRunning {
		//return workflow.ErrInvalidInstanceStatusTransition
	}

	// 执行删除
	return h.repo.Delete(ctx, cmd.ID)
}
