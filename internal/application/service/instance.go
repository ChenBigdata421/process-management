package service

import (
	"context"

	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	instance_repository "jxt-evidence-system/process-management/internal/domain/aggregate/instance/repository"
	workflow_repository "jxt-evidence-system/process-management/internal/domain/aggregate/workflow/repository"
	"jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/status"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"
)

// DeleteInstanceHandler 删除工作流实例处理器
type instanceService struct {
	workflowRepo  workflow_repository.WorkflowRepository
	instanceRepo  instance_repository.WorkflowInstanceRepository
	engineService port.WorkflowEngineService
}

// Handle 处理命令
func (h *instanceService) DeleteInstance(ctx context.Context, cmd *command.DeleteInstanceCommand) error {
	// 查找工作流实例
	instance, err := h.instanceRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if instance == nil {
		return errors.ErrInstanceNotFound
	}

	// 业务规则验证：只能删除已完成、失败或已取消的实例
	if instance.Status == status.InstanceStatusRunning {
		//return workflow.ErrInvalidInstanceStatusTransition
	}

	// 执行删除
	return h.instanceRepo.Delete(ctx, cmd.ID)
}

// GetInstanceByID 根据ID获取实例
func (h *instanceService) GetInstanceByID(ctx context.Context, id string) (*command.WorkflowInstanceDTO, error) {
	instance, err := h.instanceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if instance == nil {
		return nil, errors.ErrInstanceNotFound
	}

	completedAt := ""
	if instance.CompletedAt != nil {
		completedAt = instance.CompletedAt.String()
	}

	return &command.WorkflowInstanceDTO{
		ID:           instance.ID,
		WorkflowID:   instance.WorkflowID,
		Status:       string(instance.Status),
		Input:        string(instance.Input),
		Output:       string(instance.Output),
		ErrorMessage: instance.ErrorMessage,
		StartedAt:    instance.StartedAt.String(),
		CompletedAt:  completedAt,
		CreatedAt:    instance.CreatedAt.String(),
		UpdatedAt:    instance.UpdatedAt.String(),
	}, nil
}

// ListInstancesByWorkflowID 列出工作流的所有实例
func (h *instanceService) ListInstancesByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*command.WorkflowInstanceDTO, error) {
	instances, err := h.instanceRepo.FindByWorkflowID(ctx, workflowID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*command.WorkflowInstanceDTO, len(instances))
	for i, instance := range instances {
		completedAt := ""
		if instance.CompletedAt != nil {
			completedAt = instance.CompletedAt.String()
		}

		dtos[i] = &command.WorkflowInstanceDTO{
			ID:           instance.ID,
			WorkflowID:   instance.WorkflowID,
			Status:       string(instance.Status),
			Input:        string(instance.Input),
			Output:       string(instance.Output),
			ErrorMessage: instance.ErrorMessage,
			StartedAt:    instance.StartedAt.String(),
			CompletedAt:  completedAt,
			CreatedAt:    instance.CreatedAt.String(),
			UpdatedAt:    instance.UpdatedAt.String(),
		}
	}

	return dtos, nil
}

// ListAllInstances 列出所有实例（支持筛选）
func (h *instanceService) ListAllInstances(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*command.WorkflowInstanceDTO, int, error) {
	instances, total, err := h.instanceRepo.FindAll(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*command.WorkflowInstanceDTO, len(instances))
	for i, instance := range instances {
		completedAt := ""
		if instance.CompletedAt != nil {
			completedAt = instance.CompletedAt.String()
		}

		dtos[i] = &command.WorkflowInstanceDTO{
			ID:           instance.ID,
			WorkflowID:   instance.WorkflowID,
			Status:       string(instance.Status),
			Input:        string(instance.Input),
			Output:       string(instance.Output),
			ErrorMessage: instance.ErrorMessage,
			StartedAt:    instance.StartedAt.String(),
			CompletedAt:  completedAt,
			CreatedAt:    instance.CreatedAt.String(),
			UpdatedAt:    instance.UpdatedAt.String(),
		}
	}

	return dtos, total, nil
}

func (h *instanceService) StartWorkflowInstance(ctx context.Context, cmd *command.StartWorkflowInstanceCommand) (string, error) {
	// 验证工作流存在且处于活跃状态
	wf, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return "", err
	}

	if wf == nil {
		return "", errors.ErrWorkflowNotFound
	}

	if wf.Status != status.StatusActive {
		return "", errors.ErrInvalidStatusTransition
	}

	// 创建工作流实例
	instance := instance_aggregate.NewWorkflowInstance(cmd.WorkflowID, cmd.Input)

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
