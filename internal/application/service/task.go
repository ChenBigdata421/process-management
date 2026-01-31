package service

import (
	"context"
	"fmt"
	"log"

	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	task_repository "jxt-evidence-system/process-management/internal/domain/aggregate/task/repository"
	workflow_repository "jxt-evidence-system/process-management/internal/domain/aggregate/workflow/repository"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/status"

	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"
)

// ClaimTaskHandler 认领任务处理器
type taskService struct {
	taskRepo      task_repository.TaskRepository
	workflowRepo  workflow_repository.WorkflowRepository
	historyRepo   task_repository.TaskHistoryRepository
	engineService port.WorkflowEngineService
}

// Handle 处理完成任务命令
func (h taskService) CompleteTask(ctx context.Context, cmd *command.CompleteTaskCommand) error {
	task, err := h.taskRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if task.Assignee != cmd.UserID {
		return errors.ErrUnauthorized
	}

	if err := task.Complete(cmd); err != nil {
		return err
	}

	if err := h.taskRepo.Update(ctx, task); err != nil {
		return err
	}
	ctx = context.WithValue(ctx, "next_task_approver", cmd.NextTaskApprover)
	// 记录历史
	history := task_aggregate.NewTaskHistory(task.TaskID, task.InstanceID, task.TaskName, fmt.Sprintf("%d", cmd.UserID), "complete")
	history.Result = cmd.Result
	history.Comment = cmd.Comment
	history.Output = task.Output
	if err := h.historyRepo.Save(ctx, history); err != nil {
		return err
	}

	// 🆕 任务完成后，根据结果自动推进流程或回退
	if h.engineService != nil {
		if cmd.Result == status.TaskResultRejected {
			// 驳回：回退到上一个步骤
			if err := h.engineService.RejectAndGoBack(ctx, task); err != nil {
				// 记录错误但不影响任务完成
				log.Printf("[CompleteTaskHandler] RejectAndGoBack failed: %v", err)
			} else {
				log.Printf("[CompleteTaskHandler] RejectAndGoBack succeeded")
			}
		} else if cmd.Result == status.TaskResultApproved || cmd.Result == status.TaskResultCompleted {
			// 通过/完成：继续下一步
			if err := h.engineService.ContinueAfterTask(ctx, task); err != nil {
				// 记录错误但不影响任务完成
				log.Printf("[CompleteTaskHandler] ContinueAfterTask failed: %v", err)
			}
		}
	}

	return nil
}

// Handle 处理删除任务命令
func (h *taskService) DeleteTask(ctx context.Context, cmd *command.DeleteTaskCommand) error {

	// 查找任务
	task, err := h.taskRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	// 只有待处理状态的任务才能删除
	if task.Status != status.TaskStatusPending {
		return errors.ErrTaskNotPending
	}

	// 执行删除
	return h.taskRepo.Delete(ctx, cmd.ID)
}

func (h *taskService) CancelTask(ctx context.Context, cmd *command.CancelTaskCommand) error {

	// 查找任务
	task, err := h.taskRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if task.Assignee != cmd.UserID {
		return errors.ErrUnauthorized
	}

	// 只有待处理状态的任务才能撤销
	if task.Status != status.TaskStatusPending {
		return errors.ErrTaskNotPending
	}

	task.Status = status.TaskStatusCancelled

	// 执行删除
	return h.taskRepo.Update(ctx, task)
}

// 处理转办任务命令
func (h *taskService) DelegateTask(ctx context.Context, cmd *command.DelegateTaskCommand) error {

	task, err := h.taskRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return err
	}

	if task.Assignee != cmd.UserID {
		return errors.ErrUnauthorized
	}

	// 记录转办前的历史
	history := task_aggregate.NewTaskHistory(task.TaskID, task.InstanceID, task.TaskName, fmt.Sprintf("%d", cmd.UserID), "delegate")
	history.Comment = cmd.Comment
	if err := h.historyRepo.Save(ctx, history); err != nil {
		return err
	}

	// 更新任务处理人
	task.Assignee = cmd.TargetID
	return h.taskRepo.Update(ctx, task)
}

// Handle 处理创建任务命令
func (h *taskService) CreateTask(ctx context.Context, cmd *command.CreateTaskCommand) (valueobject.TaskID, error) {

	// 创建新任务
	task := task_aggregate.NewTask(cmd.InstanceID, cmd.WorkflowID)
	task.TaskName = cmd.TaskName
	task.TaskKey = cmd.TaskKey
	task.Description = cmd.Description
	task.Assignee = cmd.Assignee

	// 设置优先级
	if cmd.Priority == "high" {
		task.Priority = status.TaskPriorityHigh
	} else if cmd.Priority == "low" {
		task.Priority = status.TaskPriorityLow
	} else {
		task.Priority = status.TaskPriorityMedium
	}

	// 保存任务
	if err := h.taskRepo.Save(ctx, task); err != nil {
		return valueobject.TaskID{}, err
	}

	return task.TaskID, nil
}

// GetTaskByID 根据ID获取任务
func (h *taskService) GetTaskByID(ctx context.Context, taskID valueobject.TaskID) (*task_aggregate.Task, error) {
	task, err := h.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetRecentTask 根据实例ID获取最近的一条任务
func (h *taskService) GetRecentTask(ctx context.Context, instanceID valueobject.InstanceID) (*task_aggregate.Task, error) {
	task, err := h.taskRepo.FindRecentByInstanceID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetTodoTasks 查询待办任务
func (h *taskService) GetTodoTasks(ctx context.Context, userID int, query *command.TodoTaskPagedQuery) ([]*task_aggregate.Task, int, error) {
	return h.taskRepo.FindTodoByAssignee(ctx, userID, query)
}

// GetDoneTasks 查询已办任务
func (h *taskService) GetDoneTasks(ctx context.Context, userID int, query *command.DoneTaskPagedQuery) ([]*task_aggregate.Task, int, error) {
	return h.taskRepo.FindDoneByAssignee(ctx, userID, query)
}

// GetPage 查询所有任务（支持筛选）
func (h *taskService) GetPage(ctx context.Context, query *command.TaskPagedQuery) ([]*task_aggregate.Task, int, error) {
	return h.taskRepo.GetPage(ctx, query)
}

// GetTaskHistory 获取任务历史
func (h *taskService) GetTaskHistory(ctx context.Context, taskID valueobject.TaskID) ([]*task_aggregate.TaskHistory, error) {
	return h.historyRepo.FindByTaskID(ctx, taskID)
}

// GetInstanceTaskHistory 获取实例的任务历史
func (h *taskService) GetInstanceTaskHistory(ctx context.Context, instanceID valueobject.InstanceID) ([]*task_aggregate.TaskHistory, error) {
	return h.historyRepo.FindByInstanceID(ctx, instanceID)
}

// GetTasksByInstanceID 获取实例的所有任务（包含当前状态）
func (h *taskService) GetTasksByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) ([]*task_aggregate.Task, error) {
	return h.taskRepo.FindByInstanceID(ctx, instanceID)
}

func (h *taskService) CountTasksByInstanceID(ctx context.Context, instanceId valueobject.InstanceID) (int, error) {
	// 获取总数
	return h.taskRepo.CountByInstanceID(ctx, instanceId)
}
