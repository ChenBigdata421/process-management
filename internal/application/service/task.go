package service

import (
	"context"
	"log"

	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	task_repository "jxt-evidence-system/process-management/internal/domain/aggregate/task/repository"
	workflow_repository "jxt-evidence-system/process-management/internal/domain/aggregate/workflow/repository"
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

// 处理认领任务命令
func (h *taskService) ClaimTask(ctx context.Context, cmd *command.ClaimTaskCommand) error {
	task, err := h.taskRepo.FindByID(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if task == nil {
		return errors.ErrTaskNotFound
	}

	if err := task.Claim(cmd.UserID); err != nil {
		return err
	}

	if err := h.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// 记录历史
	history := task_aggregate.NewTaskHistory(task.ID, task.InstanceID, task.TaskName, cmd.UserID, "claim")
	return h.historyRepo.Save(ctx, history)
}

// Handle 处理完成任务命令
func (h taskService) CompleteTask(ctx context.Context, cmd *command.CompleteTaskCommand) error {
	task, err := h.taskRepo.FindByID(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if task == nil {
		return errors.ErrTaskNotFound
	}

	if task.Assignee != cmd.UserID {
		return errors.ErrUnauthorized
	}

	if err := task.Complete(cmd.Output, cmd.Comment, cmd.Result); err != nil {
		return err
	}

	if err := h.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// 记录历史
	history := task_aggregate.NewTaskHistory(task.ID, task.InstanceID, task.TaskName, cmd.UserID, "complete")
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
			log.Printf("[CompleteTaskHandler] Task rejected, calling RejectAndGoBack for task: %s", task.ID)
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
	task, err := h.taskRepo.FindByID(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if task == nil {
		return errors.ErrTaskNotFound
	}

	// 只有待处理状态的任务才能删除
	if task.Status != status.TaskStatusPending {
		return errors.ErrTaskNotPending
	}

	// 删除任务
	return h.taskRepo.Delete(ctx, cmd.TaskID)
}

// 处理转办任务命令
func (h *taskService) DelegateTask(ctx context.Context, cmd *command.DelegateTaskCommand) error {
	task, err := h.taskRepo.FindByID(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if task == nil {
		return errors.ErrTaskNotFound
	}

	if task.Assignee != cmd.UserID {
		return errors.ErrUnauthorized
	}

	// 记录转办前的历史
	history := task_aggregate.NewTaskHistory(task.ID, task.InstanceID, task.TaskName, cmd.UserID, "delegate")
	history.Comment = cmd.Comment
	if err := h.historyRepo.Save(ctx, history); err != nil {
		return err
	}

	// 更新任务处理人
	task.Assignee = cmd.TargetID
	return h.taskRepo.Update(ctx, task)
}

// Handle 处理创建任务命令
func (h *taskService) CreateTask(ctx context.Context, cmd *command.CreateTaskCommand) (string, error) {
	// 创建新任务
	task := task_aggregate.NewTask(cmd.InstanceID, cmd.WorkflowID, cmd.TaskName, cmd.TaskKey)
	task.Description = cmd.Description
	task.Assignee = cmd.Assignee
	task.CandidateUsers = cmd.CandidateUsers
	task.CandidateGroups = cmd.CandidateGroups

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
		return "", err
	}

	return task.ID, nil
}

// GetTaskByID 根据ID获取任务
func (h *taskService) GetTaskByID(ctx context.Context, id string) (*command.TaskDTO, error) {
	task, err := h.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, errors.ErrTaskNotFound
	}

	return h.taskToDTO(ctx, task), nil
}

// ListTodoTasks 查询待办任务
func (h *taskService) ListTodoTasks(ctx context.Context, userID string, limit, offset int) ([]*command.TaskDTO, int64, error) {
	tasks, total, err := h.taskRepo.FindTodoByAssignee(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*command.TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = h.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// ListDoneTasks 查询已办任务
func (h *taskService) ListDoneTasks(ctx context.Context, userID string, limit, offset int) ([]*command.TaskDTO, int64, error) {
	tasks, total, err := h.taskRepo.FindDoneByAssignee(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*command.TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = h.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// ListClaimableTasks 查询可认领的任务
func (h *taskService) ListClaimableTasks(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*command.TaskDTO, int64, error) {
	tasks, total, err := h.taskRepo.FindClaimable(ctx, userID, userGroups, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*command.TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = h.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// ListTasksByInstanceID 查询实例的所有任务
func (h *taskService) ListTasksByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*command.TaskDTO, error) {
	tasks, err := h.taskRepo.FindByInstanceID(ctx, instanceID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*command.TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = h.taskToDTO(ctx, task)
	}

	return dtos, nil
}

// ListAllTasks 查询所有任务（支持多条件查询）
func (h *taskService) ListAllTasks(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*command.TaskDTO, int64, error) {
	tasks, total, err := h.taskRepo.FindAll(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*command.TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = h.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// GetTaskHistory 获取任务历史
func (h *taskService) GetTaskHistory(ctx context.Context, taskID string, limit, offset int) ([]*command.TaskHistoryDTO, error) {
	histories, err := h.historyRepo.FindByTaskID(ctx, taskID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*command.TaskHistoryDTO, len(histories))
	for i, history := range histories {
		dtos[i] = historyToDTO(history)
	}

	return dtos, nil
}

// GetInstanceTaskHistory 获取实例的任务历史
func (h *taskService) GetInstanceTaskHistory(ctx context.Context, instanceID string, limit, offset int) ([]*command.TaskHistoryDTO, error) {
	histories, err := h.historyRepo.FindByInstanceID(ctx, instanceID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*command.TaskHistoryDTO, len(histories))
	for i, history := range histories {
		dtos[i] = historyToDTO(history)
	}

	return dtos, nil
}

// GetInstanceTasks 获取实例的所有任务（包含当前状态）
func (h *taskService) GetInstanceTasks(ctx context.Context, instanceID string, limit, offset int) ([]*command.TaskDTO, int64, error) {
	tasks, err := h.taskRepo.FindByInstanceID(ctx, instanceID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数（不分页）
	allTasks, err := h.taskRepo.FindByInstanceID(ctx, instanceID, 999999, 0)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(allTasks))

	dtos := make([]*command.TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = h.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// 辅助函数

func (h *taskService) taskToDTO(ctx context.Context, task *task_aggregate.Task) *command.TaskDTO {
	claimedAt := ""
	if task.ClaimedAt != nil {
		claimedAt = task.ClaimedAt.String()
	}

	completedAt := ""
	if task.CompletedAt != nil {
		completedAt = task.CompletedAt.String()
	}

	dueDate := ""
	if task.DueDate != nil {
		dueDate = task.DueDate.String()
	}

	// 查询工作流名称
	workflowName := ""
	if task.WorkflowID != "" {
		wf, err := h.workflowRepo.FindByID(ctx, task.WorkflowID)
		if err == nil && wf != nil {
			workflowName = wf.Name
		}
	}

	return &command.TaskDTO{
		ID:              task.ID,
		InstanceID:      task.InstanceID,
		WorkflowID:      task.WorkflowID,
		WorkflowName:    workflowName,
		TaskName:        task.TaskName,
		TaskKey:         task.TaskKey,
		Description:     task.Description,
		TaskType:        task.TaskType,
		Assignee:        task.Assignee,
		CandidateUsers:  task.CandidateUsers,
		CandidateGroups: task.CandidateGroups,
		Status:          string(task.Status),
		Priority:        string(task.Priority),
		TaskData:        string(task.TaskData),
		FormData:        string(task.FormData),
		Output:          string(task.Output),
		Result:          string(task.Result),
		Comment:         task.Comment,
		CreatedAt:       task.CreatedAt.String(),
		ClaimedAt:       claimedAt,
		CompletedAt:     completedAt,
		DueDate:         dueDate,
	}
}

func historyToDTO(history *task_aggregate.TaskHistory) *command.TaskHistoryDTO {
	return &command.TaskHistoryDTO{
		ID:         history.ID,
		TaskID:     history.TaskID,
		InstanceID: history.InstanceID,
		TaskName:   history.TaskName,
		Assignee:   history.Assignee,
		Action:     history.Action,
		Result:     string(history.Result),
		Comment:    history.Comment,
		Output:     string(history.Output),
		CreatedAt:  history.CreatedAt.String(),
	}
}
