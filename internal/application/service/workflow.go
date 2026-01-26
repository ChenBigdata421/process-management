package service

import (
	"context"
	"time"

	"jxt-evidence-system/process-management/internal/application/command"
	workflow_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
	workflow_repository "jxt-evidence-system/process-management/internal/domain/aggregate/workflow/repository"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/status"
)

// ActivateWorkflowHandler 激活工作流处理器
type workflowService struct {
	repo workflow_repository.WorkflowRepository
}

func (h *workflowService) ActivateWorkflow(ctx context.Context, cmd *command.ActivateWorkflowCommand) error {

	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	// 激活工作流
	if err := wf.Activate(); err != nil {
		return err
	}

	// 保存更新
	return h.repo.Update(ctx, wf)
}

func (h *workflowService) CreateWorkflow(ctx context.Context, cmd *command.CreateWorkflowCommand) (string, error) {
	// 业务规则验证
	if cmd.Name == "" || cmd.Definition == "" {
		return "", errors.ErrInvalidWorkflowDefinition
	}

	// 创建工作流
	wf := workflow_aggregate.NewWorkflow(cmd.Name, cmd.Description, cmd.Definition, cmd.CreateBy)

	// 保存到仓储
	if err := h.repo.Save(ctx, wf); err != nil {
		return "", err
	}

	return wf.WorkflowID.String(), nil
}

func (h *workflowService) DeleteWorkflow(ctx context.Context, cmd *command.DeleteWorkflowCommand) error {

	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	// 业务规则验证：只能删除草稿或已取消的工作流
	if wf.Status != status.StatusDraft && wf.Status != status.StatusCancelled {
		return errors.ErrInvalidStatusTransition
	}

	// 执行删除
	return h.repo.Delete(ctx, cmd.ID)
}

func (h *workflowService) FreezeWorkflow(ctx context.Context, cmd *command.FreezeWorkflowCommand) error {

	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	// 冻结工作流
	if err := wf.Freeze(); err != nil {
		return err
	}

	// 保存更新
	return h.repo.Update(ctx, wf)
}

// GetWorkflowByID 根据ID获取工作流
func (h *workflowService) GetWorkflowByID(ctx context.Context, workflowID valueobject.WorkflowID) (*workflow_aggregate.Workflow, error) {
	wf, err := h.repo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	if wf == nil {
		return nil, errors.ErrWorkflowNotFound
	}

	return wf, nil
}

// GetWorkflowByID 根据ID获取工作流
func (h *workflowService) GetWorkflowByName(ctx context.Context, name string) (*workflow_aggregate.Workflow, error) {
	wf, err := h.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if wf == nil {
		return nil, errors.ErrWorkflowNotFound
	}

	return wf, nil
}

// GetPage 列出所有工作流（支持筛选）
func (h *workflowService) GetPage(ctx context.Context, query *command.WorkflowPagedQuery) ([]*workflow_aggregate.Workflow, int, error) {
	return h.repo.GetPage(ctx, query)
}

func (h *workflowService) GetAllWorkflow(ctx context.Context) ([]*workflow_aggregate.Workflow, error) {
	return h.repo.GetAllWorkflow(ctx)
}

// CountWorkflows 统计工作流数量
func (h *workflowService) CountWorkflows(ctx context.Context) (int64, error) {
	return h.repo.Count(ctx)
}

// Handle 处理命令
func (h *workflowService) UpdateWorkflow(ctx context.Context, cmd *command.UpdateWorkflowCommand) error {

	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	// 业务规则验证：只能更新草稿或冻结状态的工作流
	if wf.Status != status.StatusDraft && wf.Status != status.StatusFrozen {
		return errors.ErrInvalidStatusTransition
	}

	// 更新字段
	wf.Name = cmd.Name
	wf.Description = cmd.Description
	wf.Definition = cmd.Definition
	wf.UpdateBy = cmd.UpdateBy
	wf.UpdatedAt = time.Now()

	// 保存更新
	return h.repo.Update(ctx, wf)
}
