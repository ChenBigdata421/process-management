package service

import (
	"context"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	instance_repository "jxt-evidence-system/process-management/internal/domain/aggregate/instance/repository"
	domain_service "jxt-evidence-system/process-management/internal/domain/service"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/errors"
	errors_ "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/status"
)

// DeleteInstanceHandler 删除工作流实例处理器
type instanceService struct {
	workflowService port.WorkflowService
	instanceRepo    instance_repository.WorkflowInstanceRepository
	taskService     port.TaskService
	engineService   port.WorkflowEngineService
	domainService   domain_service.WorkflowDomainService
}

// CancelInstance 取消运行中的实例（仅标记状态，不删除记录）
func (h *instanceService) CancelInstance(ctx context.Context, cmd *command.CancelInstanceCommand) error {
	instance, err := h.instanceRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if err := instance.Cancel(); err != nil {
		return err
	}
	return h.instanceRepo.Update(ctx, instance)
}

// Handle 处理命令
func (h *instanceService) DeleteInstance(ctx context.Context, cmd *command.DeleteInstanceCommand) error {

	// 查找工作流实例
	instance, err := h.instanceRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}
	count, err := h.taskService.CountTasksByInstanceID(ctx, cmd.ID)
	if err != nil {
		return err

	}
	// 业务规则验证：只能删除已完成、失败或已取消的实例,或者没有任务的实例
	if instance.Status == status.InstanceStatusRunning && count != 0 {
		return errors_.ErrInvalidInstanceStatusTransition
	}

	// 执行删除
	return h.instanceRepo.Delete(ctx, cmd.ID)
}

// GetInstanceByID 根据ID获取实例
func (h *instanceService) GetInstanceByID(ctx context.Context, id valueobject.InstanceID) (*instance_aggregate.WorkflowInstance, error) {

	return h.instanceRepo.FindByID(ctx, id)
}

func (h *instanceService) GetInstanceDetailByID(ctx context.Context, id valueobject.InstanceID) ([]command.TaskHistoryItem, error) {
	tasks, err := h.taskService.GetTasksByInstanceID(ctx, id)
	if err != nil {
		return nil, err
	}
	return h.domainService.BuildInstanceDetail(tasks), nil
}

// ListInstancesByWorkflowID 列出工作流的所有实例
func (h *instanceService) GetInstancesByWorkflow(ctx context.Context, query *command.GetInstancesByWorkflowPagedQuery) ([]*instance_aggregate.WorkflowInstance, int, error) {

	return h.instanceRepo.FindByWorkflowID(ctx, query)

}

// GetPage 列出所有实例（支持筛选）
func (h *instanceService) GetPage(ctx context.Context, query *command.InstancePagedQuery) ([]*instance_aggregate.WorkflowInstance, int, error) {
	return h.instanceRepo.GetPage(ctx, query)
}

// CountInstanceByWorkflow 统计工作流的实例数量
func (h *instanceService) CountInstanceByWorkflow(ctx context.Context, workflowID valueobject.WorkflowID) (int64, error) {
	return h.instanceRepo.CountByWorkflowID(ctx, workflowID)
}

func (h *instanceService) StartWorkflowInstance(ctx context.Context, cmd *command.StartWorkflowInstanceCommand) (string, error) {
	// 验证工作流存在且处于活跃状态
	wf, err := h.workflowService.GetWorkflowByID(ctx, cmd.ID)
	if err != nil {
		return "", err
	}
	if wf.Status != status.StatusActive {
		return "", errors.ErrInvalidStatusTransition
	}

	// 创建工作流实例
	instance := instance_aggregate.NewWorkflowInstance(cmd.ID, cmd.Input)

	// 保存实例
	if err := h.instanceRepo.Save(ctx, instance); err != nil {
		return "", err
	}

	// 🆕 启动工作流引擎，自动执行第一步
	if h.engineService != nil {
		if err := h.engineService.StartInstance(ctx, instance.InstanceId); err != nil {
			return "", err
		}
	}

	return instance.InstanceId.String(), nil
}
