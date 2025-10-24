package command

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// StartWorkflowInstanceCommand 启动工作流实例命令
type StartWorkflowInstanceCommand struct {
	WorkflowID string
	Input      string
}

// StartWorkflowInstanceHandler 启动工作流实例处理器
type StartWorkflowInstanceHandler struct {
	workflowRepo  workflow.WorkflowRepository
	instanceRepo  workflow.WorkflowInstanceRepository
	engineService *workflow.WorkflowEngineService
}

// NewStartWorkflowInstanceHandler 创建处理器
func NewStartWorkflowInstanceHandler(
	workflowRepo workflow.WorkflowRepository,
	instanceRepo workflow.WorkflowInstanceRepository,
	engineService *workflow.WorkflowEngineService,
) *StartWorkflowInstanceHandler {
	return &StartWorkflowInstanceHandler{
		workflowRepo:  workflowRepo,
		instanceRepo:  instanceRepo,
		engineService: engineService,
	}
}

// Handle 处理命令
func (h *StartWorkflowInstanceHandler) Handle(ctx context.Context, cmd *StartWorkflowInstanceCommand) (string, error) {
	// 验证工作流存在且处于活跃状态
	wf, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return "", err
	}

	if wf == nil {
		return "", workflow.ErrWorkflowNotFound
	}

	if wf.Status != workflow.StatusActive {
		return "", workflow.ErrInvalidStatusTransition
	}

	// 创建工作流实例
	instance := workflow.NewWorkflowInstance(cmd.WorkflowID, cmd.Input)

	// 保存实例
	if err := h.instanceRepo.Save(ctx, instance); err != nil {
		return "", err
	}

	// 🆕 启动工作流引擎，自动执行第一步
	if h.engineService != nil {
		if err := h.engineService.StartInstance(ctx, instance.ID); err != nil {
			// 记录错误但不影响实例创建
			// 可以后续手动触发
			// TODO: 添加日志记录
			_ = err
		}
	}

	return instance.ID, nil
}
