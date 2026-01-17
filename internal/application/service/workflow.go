package service

import (
	"context"

	"jxt-evidence-system/process-management/internal/application/command"
	workflow_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
	workflow_repository "jxt-evidence-system/process-management/internal/domain/aggregate/workflow/repository"
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

	if wf == nil {
		return errors.ErrWorkflowNotFound
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
	if cmd.Name == "" {
		return "", errors.ErrInvalidWorkflowDefinition
	}

	// 创建工作流
	wf := workflow_aggregate.NewWorkflow(cmd.Name, cmd.Description, cmd.Definition)

	// 保存到仓储
	if err := h.repo.Save(ctx, wf); err != nil {
		return "", err
	}

	return wf.ID, nil
}

func (h *workflowService) DeleteWorkflow(ctx context.Context, cmd *command.DeleteWorkflowCommand) error {
	// 查找工作流
	wf, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if wf == nil {
		return errors.ErrWorkflowNotFound
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

	if wf == nil {
		return errors.ErrWorkflowNotFound
	}

	// 冻结工作流
	if err := wf.Freeze(); err != nil {
		return err
	}

	// 保存更新
	return h.repo.Update(ctx, wf)
}

// GetWorkflowByID 根据ID获取工作流
func (h *workflowService) GetWorkflowByID(ctx context.Context, id string) (*command.WorkflowDTO, error) {
	wf, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if wf == nil {
		return nil, errors.ErrWorkflowNotFound
	}

	return &command.WorkflowDTO{
		ID:          wf.ID,
		Name:        wf.Name,
		Description: wf.Description,
		Status:      string(wf.Status),
		Definition:  wf.Definition,
		CreatedAt:   wf.CreatedAt.String(),
		UpdatedAt:   wf.UpdatedAt.String(),
	}, nil
}

// ListWorkflows 列出所有工作流
func (h *workflowService) ListWorkflows(ctx context.Context, limit, offset int) ([]*command.WorkflowDTO, error) {
	workflows, err := h.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*command.WorkflowDTO, len(workflows))
	for i, wf := range workflows {
		dtos[i] = &command.WorkflowDTO{
			ID:          wf.ID,
			Name:        wf.Name,
			Description: wf.Description,
			Status:      string(wf.Status),
			Definition:  wf.Definition,
			CreatedAt:   wf.CreatedAt.String(),
			UpdatedAt:   wf.UpdatedAt.String(),
		}
	}

	return dtos, nil
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

	if wf == nil {
		return errors.ErrWorkflowNotFound
	}

	// 业务规则验证：只能更新草稿或冻结状态的工作流
	if wf.Status != status.StatusDraft && wf.Status != status.StatusFrozen {
		return errors.ErrInvalidStatusTransition
	}

	// 更新字段
	wf.Name = cmd.Name
	wf.Description = cmd.Description
	wf.Definition = cmd.Definition

	// 保存更新
	return h.repo.Update(ctx, wf)
}
