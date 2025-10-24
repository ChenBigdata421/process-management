package command

import (
	"context"
	"log"

	"github.com/jxt/process-management/domain/workflow"
)

// ClaimTaskCommand 认领任务命令
type ClaimTaskCommand struct {
	TaskID string
	UserID string
}

// ClaimTaskHandler 认领任务处理器
type ClaimTaskHandler struct {
	taskRepo    workflow.TaskRepository
	historyRepo workflow.TaskHistoryRepository
}

// NewClaimTaskHandler 创建认领任务处理器
func NewClaimTaskHandler(taskRepo workflow.TaskRepository, historyRepo workflow.TaskHistoryRepository) *ClaimTaskHandler {
	return &ClaimTaskHandler{
		taskRepo:    taskRepo,
		historyRepo: historyRepo,
	}
}

// Handle 处理认领任务命令
func (h *ClaimTaskHandler) Handle(ctx context.Context, cmd *ClaimTaskCommand) error {
	task, err := h.taskRepo.FindByID(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if task == nil {
		return workflow.ErrTaskNotFound
	}

	if err := task.Claim(cmd.UserID); err != nil {
		return err
	}

	if err := h.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// 记录历史
	history := workflow.NewTaskHistory(task.ID, task.InstanceID, task.TaskName, cmd.UserID, "claim")
	return h.historyRepo.Save(ctx, history)
}

// CompleteTaskCommand 完成任务命令
type CompleteTaskCommand struct {
	TaskID  string
	UserID  string
	Output  string
	Comment string
	Result  workflow.TaskResult
}

// CompleteTaskHandler 完成任务处理器
type CompleteTaskHandler struct {
	taskRepo      workflow.TaskRepository
	historyRepo   workflow.TaskHistoryRepository
	engineService *workflow.WorkflowEngineService
}

// NewCompleteTaskHandler 创建完成任务处理器
func NewCompleteTaskHandler(
	taskRepo workflow.TaskRepository,
	historyRepo workflow.TaskHistoryRepository,
	engineService *workflow.WorkflowEngineService,
) *CompleteTaskHandler {
	return &CompleteTaskHandler{
		taskRepo:      taskRepo,
		historyRepo:   historyRepo,
		engineService: engineService,
	}
}

// Handle 处理完成任务命令
func (h *CompleteTaskHandler) Handle(ctx context.Context, cmd *CompleteTaskCommand) error {
	task, err := h.taskRepo.FindByID(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if task == nil {
		return workflow.ErrTaskNotFound
	}

	if task.Assignee != cmd.UserID {
		return workflow.ErrUnauthorized
	}

	if err := task.Complete(cmd.Output, cmd.Comment, cmd.Result); err != nil {
		return err
	}

	if err := h.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// 记录历史
	history := workflow.NewTaskHistory(task.ID, task.InstanceID, task.TaskName, cmd.UserID, "complete")
	history.Result = cmd.Result
	history.Comment = cmd.Comment
	history.Output = task.Output
	if err := h.historyRepo.Save(ctx, history); err != nil {
		return err
	}

	// 🆕 任务完成后，根据结果自动推进流程或回退
	if h.engineService != nil {
		if cmd.Result == workflow.TaskResultRejected {
			// 驳回：回退到上一个步骤
			log.Printf("[CompleteTaskHandler] Task rejected, calling RejectAndGoBack for task: %s", task.ID)
			if err := h.engineService.RejectAndGoBack(ctx, task); err != nil {
				// 记录错误但不影响任务完成
				log.Printf("[CompleteTaskHandler] RejectAndGoBack failed: %v", err)
			} else {
				log.Printf("[CompleteTaskHandler] RejectAndGoBack succeeded")
			}
		} else if cmd.Result == workflow.TaskResultApproved || cmd.Result == workflow.TaskResultCompleted {
			// 通过/完成：继续下一步
			if err := h.engineService.ContinueAfterTask(ctx, task); err != nil {
				// 记录错误但不影响任务完成
				log.Printf("[CompleteTaskHandler] ContinueAfterTask failed: %v", err)
			}
		}
	}

	return nil
}

// DelegateTaskCommand 转办任务命令
type DelegateTaskCommand struct {
	TaskID   string
	UserID   string
	TargetID string
	Comment  string
}

// DeleteTaskCommand 删除任务命令
type DeleteTaskCommand struct {
	TaskID string
}

// DeleteTaskHandler 删除任务处理器
type DeleteTaskHandler struct {
	taskRepo workflow.TaskRepository
}

// NewDeleteTaskHandler 创建删除任务处理器
func NewDeleteTaskHandler(taskRepo workflow.TaskRepository) *DeleteTaskHandler {
	return &DeleteTaskHandler{
		taskRepo: taskRepo,
	}
}

// Handle 处理删除任务命令
func (h *DeleteTaskHandler) Handle(ctx context.Context, cmd *DeleteTaskCommand) error {
	task, err := h.taskRepo.FindByID(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if task == nil {
		return workflow.ErrTaskNotFound
	}

	// 只有待处理状态的任务才能删除
	if task.Status != workflow.TaskStatusPending {
		return workflow.ErrTaskNotPending
	}

	// 删除任务
	return h.taskRepo.Delete(ctx, cmd.TaskID)
}

// DelegateTaskHandler 转办任务处理器
type DelegateTaskHandler struct {
	taskRepo    workflow.TaskRepository
	historyRepo workflow.TaskHistoryRepository
}

// NewDelegateTaskHandler 创建转办任务处理器
func NewDelegateTaskHandler(taskRepo workflow.TaskRepository, historyRepo workflow.TaskHistoryRepository) *DelegateTaskHandler {
	return &DelegateTaskHandler{
		taskRepo:    taskRepo,
		historyRepo: historyRepo,
	}
}

// Handle 处理转办任务命令
func (h *DelegateTaskHandler) Handle(ctx context.Context, cmd *DelegateTaskCommand) error {
	task, err := h.taskRepo.FindByID(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if task == nil {
		return workflow.ErrTaskNotFound
	}

	if task.Assignee != cmd.UserID {
		return workflow.ErrUnauthorized
	}

	// 记录转办前的历史
	history := workflow.NewTaskHistory(task.ID, task.InstanceID, task.TaskName, cmd.UserID, "delegate")
	history.Comment = cmd.Comment
	if err := h.historyRepo.Save(ctx, history); err != nil {
		return err
	}

	// 更新任务处理人
	task.Assignee = cmd.TargetID
	return h.taskRepo.Update(ctx, task)
}

// CreateTaskCommand 创建任务命令
type CreateTaskCommand struct {
	InstanceID      string
	WorkflowID      string
	TaskName        string
	TaskKey         string
	Description     string
	Assignee        string
	CandidateUsers  []string
	CandidateGroups []string
	Priority        string
}

// CreateTaskHandler 创建任务处理器
type CreateTaskHandler struct {
	taskRepo workflow.TaskRepository
}

// NewCreateTaskHandler 创建任务处理器
func NewCreateTaskHandler(taskRepo workflow.TaskRepository) *CreateTaskHandler {
	return &CreateTaskHandler{
		taskRepo: taskRepo,
	}
}

// Handle 处理创建任务命令
func (h *CreateTaskHandler) Handle(ctx context.Context, cmd *CreateTaskCommand) (string, error) {
	// 创建新任务
	task := workflow.NewTask(cmd.InstanceID, cmd.WorkflowID, cmd.TaskName, cmd.TaskKey)
	task.Description = cmd.Description
	task.Assignee = cmd.Assignee
	task.CandidateUsers = cmd.CandidateUsers
	task.CandidateGroups = cmd.CandidateGroups

	// 设置优先级
	if cmd.Priority == "high" {
		task.Priority = workflow.TaskPriorityHigh
	} else if cmd.Priority == "low" {
		task.Priority = workflow.TaskPriorityLow
	} else {
		task.Priority = workflow.TaskPriorityMedium
	}

	// 保存任务
	if err := h.taskRepo.Save(ctx, task); err != nil {
		return "", err
	}

	return task.ID, nil
}
