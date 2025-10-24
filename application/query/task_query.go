package query

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// TaskDTO 任务数据传输对象
type TaskDTO struct {
	ID              string   `json:"id"`
	InstanceID      string   `json:"instance_id"`
	WorkflowID      string   `json:"workflow_id"`
	WorkflowName    string   `json:"workflow_name"`
	TaskName        string   `json:"task_name"`
	TaskKey         string   `json:"task_key"`
	Description     string   `json:"description"`
	TaskType        string   `json:"task_type"`
	Assignee        string   `json:"assignee"`
	CandidateUsers  []string `json:"candidate_users"`
	CandidateGroups []string `json:"candidate_groups"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority"`
	TaskData        string   `json:"task_data"`
	FormData        string   `json:"form_data"`
	Output          string   `json:"output"`
	Result          string   `json:"result"`
	Comment         string   `json:"comment"`
	CreatedAt       string   `json:"created_at"`
	ClaimedAt       string   `json:"claimed_at"`
	CompletedAt     string   `json:"completed_at"`
	DueDate         string   `json:"due_date"`
}

// TaskHistoryDTO 任务历史数据传输对象
type TaskHistoryDTO struct {
	ID         string `json:"id"`
	TaskID     string `json:"task_id"`
	InstanceID string `json:"instance_id"`
	TaskName   string `json:"task_name"`
	Assignee   string `json:"assignee"`
	Action     string `json:"action"`
	Result     string `json:"result"`
	Comment    string `json:"comment"`
	Output     string `json:"output"`
	CreatedAt  string `json:"created_at"`
}

// TaskQueryService 任务查询服务
type TaskQueryService struct {
	taskRepo     workflow.TaskRepository
	historyRepo  workflow.TaskHistoryRepository
	workflowRepo workflow.WorkflowRepository
}

// NewTaskQueryService 创建任务查询服务
func NewTaskQueryService(taskRepo workflow.TaskRepository, historyRepo workflow.TaskHistoryRepository, workflowRepo workflow.WorkflowRepository) *TaskQueryService {
	return &TaskQueryService{
		taskRepo:     taskRepo,
		historyRepo:  historyRepo,
		workflowRepo: workflowRepo,
	}
}

// GetTaskByID 根据ID获取任务
func (qs *TaskQueryService) GetTaskByID(ctx context.Context, id string) (*TaskDTO, error) {
	task, err := qs.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, workflow.ErrTaskNotFound
	}

	return qs.taskToDTO(ctx, task), nil
}

// ListTodoTasks 查询待办任务
func (qs *TaskQueryService) ListTodoTasks(ctx context.Context, userID string, limit, offset int) ([]*TaskDTO, int64, error) {
	tasks, total, err := qs.taskRepo.FindTodoByAssignee(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = qs.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// ListDoneTasks 查询已办任务
func (qs *TaskQueryService) ListDoneTasks(ctx context.Context, userID string, limit, offset int) ([]*TaskDTO, int64, error) {
	tasks, total, err := qs.taskRepo.FindDoneByAssignee(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = qs.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// ListClaimableTasks 查询可认领的任务
func (qs *TaskQueryService) ListClaimableTasks(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*TaskDTO, int64, error) {
	tasks, total, err := qs.taskRepo.FindClaimable(ctx, userID, userGroups, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = qs.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// ListTasksByInstanceID 查询实例的所有任务
func (qs *TaskQueryService) ListTasksByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*TaskDTO, error) {
	tasks, err := qs.taskRepo.FindByInstanceID(ctx, instanceID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = qs.taskToDTO(ctx, task)
	}

	return dtos, nil
}

// ListAllTasks 查询所有任务（支持多条件查询）
func (qs *TaskQueryService) ListAllTasks(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*TaskDTO, int64, error) {
	tasks, total, err := qs.taskRepo.FindAll(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = qs.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// GetTaskHistory 获取任务历史
func (qs *TaskQueryService) GetTaskHistory(ctx context.Context, taskID string, limit, offset int) ([]*TaskHistoryDTO, error) {
	histories, err := qs.historyRepo.FindByTaskID(ctx, taskID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*TaskHistoryDTO, len(histories))
	for i, history := range histories {
		dtos[i] = historyToDTO(history)
	}

	return dtos, nil
}

// GetInstanceTaskHistory 获取实例的任务历史
func (qs *TaskQueryService) GetInstanceTaskHistory(ctx context.Context, instanceID string, limit, offset int) ([]*TaskHistoryDTO, error) {
	histories, err := qs.historyRepo.FindByInstanceID(ctx, instanceID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*TaskHistoryDTO, len(histories))
	for i, history := range histories {
		dtos[i] = historyToDTO(history)
	}

	return dtos, nil
}

// GetInstanceTasks 获取实例的所有任务（包含当前状态）
func (qs *TaskQueryService) GetInstanceTasks(ctx context.Context, instanceID string, limit, offset int) ([]*TaskDTO, int64, error) {
	tasks, err := qs.taskRepo.FindByInstanceID(ctx, instanceID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数（不分页）
	allTasks, err := qs.taskRepo.FindByInstanceID(ctx, instanceID, 999999, 0)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(allTasks))

	dtos := make([]*TaskDTO, len(tasks))
	for i, task := range tasks {
		dtos[i] = qs.taskToDTO(ctx, task)
	}

	return dtos, total, nil
}

// 辅助函数

func (qs *TaskQueryService) taskToDTO(ctx context.Context, task *workflow.Task) *TaskDTO {
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
		wf, err := qs.workflowRepo.FindByID(ctx, task.WorkflowID)
		if err == nil && wf != nil {
			workflowName = wf.Name
		}
	}

	return &TaskDTO{
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

func historyToDTO(history *workflow.TaskHistory) *TaskHistoryDTO {
	return &TaskHistoryDTO{
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
