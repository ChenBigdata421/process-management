package query_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/query"
	"github.com/jxt/process-management/domain/workflow"
)

type MockTaskRepository struct {
	tasks map[string]*workflow.Task
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks: make(map[string]*workflow.Task),
	}
}

func (m *MockTaskRepository) Save(ctx context.Context, task *workflow.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *MockTaskRepository) FindByID(ctx context.Context, id string) (*workflow.Task, error) {
	if task, ok := m.tasks[id]; ok {
		return task, nil
	}
	return nil, workflow.ErrTaskNotFound
}

func (m *MockTaskRepository) FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*workflow.Task, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		if task.InstanceID == instanceID {
			result = append(result, task)
		}
	}
	return result, nil
}

func (m *MockTaskRepository) FindTodoByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*workflow.Task, int64, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		if task.Assignee == assignee && (task.Status == workflow.TaskStatusPending || task.Status == workflow.TaskStatusClaimed) {
			result = append(result, task)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockTaskRepository) FindDoneByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*workflow.Task, int64, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		if task.Assignee == assignee && (task.Status == workflow.TaskStatusCompleted || task.Status == workflow.TaskStatusRejected) {
			result = append(result, task)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockTaskRepository) FindClaimable(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*workflow.Task, int64, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		if task.Status == workflow.TaskStatusPending && (task.Assignee == "" || task.Assignee == userID) {
			result = append(result, task)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockTaskRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*workflow.Task, int64, error) {
	var result []*workflow.Task
	for _, task := range m.tasks {
		result = append(result, task)
	}
	return result, int64(len(result)), nil
}

func (m *MockTaskRepository) Update(ctx context.Context, task *workflow.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	delete(m.tasks, id)
	return nil
}

type MockTaskHistoryRepository struct {
	histories map[string]*workflow.TaskHistory
}

func NewMockTaskHistoryRepository() *MockTaskHistoryRepository {
	return &MockTaskHistoryRepository{
		histories: make(map[string]*workflow.TaskHistory),
	}
}

func (m *MockTaskHistoryRepository) Save(ctx context.Context, history *workflow.TaskHistory) error {
	m.histories[history.ID] = history
	return nil
}

func (m *MockTaskHistoryRepository) FindByTaskID(ctx context.Context, taskID string, limit, offset int) ([]*workflow.TaskHistory, error) {
	var result []*workflow.TaskHistory
	for _, history := range m.histories {
		if history.TaskID == taskID {
			result = append(result, history)
		}
	}
	return result, nil
}

func (m *MockTaskHistoryRepository) FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*workflow.TaskHistory, error) {
	var result []*workflow.TaskHistory
	for _, history := range m.histories {
		if history.InstanceID == instanceID {
			result = append(result, history)
		}
	}
	return result, nil
}

func TestWorkflowQueryService(t *testing.T) {
	t.Run("列出工作流", func(t *testing.T) {
		repo := NewMockWorkflowRepository()

		// 创建工作流
		for i := 0; i < 3; i++ {
			wf := workflow.NewWorkflow("工作流", "描述", `{"steps": ["validate"]}`)
			repo.Save(context.Background(), wf)
		}

		// 查询工作流
		queryService := query.NewWorkflowQueryService(repo)
		workflows, err := queryService.ListWorkflows(context.Background(), 10, 0)
		if err != nil {
			t.Errorf("ListWorkflows() error = %v", err)
		}
		if len(workflows) != 3 {
			t.Errorf("Expected 3 workflows, got %d", len(workflows))
		}
	})

	t.Run("计数工作流", func(t *testing.T) {
		repo := NewMockWorkflowRepository()

		// 创建工作流
		for i := 0; i < 5; i++ {
			wf := workflow.NewWorkflow("工作流", "描述", `{"steps": ["validate"]}`)
			repo.Save(context.Background(), wf)
		}

		// 计数工作流
		queryService := query.NewWorkflowQueryService(repo)
		count, err := queryService.CountWorkflows(context.Background())
		if err != nil {
			t.Errorf("CountWorkflows() error = %v", err)
		}
		if count != 5 {
			t.Errorf("Expected 5 workflows, got %d", count)
		}
	})

	t.Run("获取工作流详情", func(t *testing.T) {
		repo := NewMockWorkflowRepository()

		// 创建工作流
		wf := workflow.NewWorkflow("订单处理", "处理订单的流程", `{"steps": ["validate"]}`)
		repo.Save(context.Background(), wf)

		// 查询工作流
		queryService := query.NewWorkflowQueryService(repo)
		dto, err := queryService.GetWorkflowByID(context.Background(), wf.ID)
		if err != nil {
			t.Errorf("GetWorkflowByID() error = %v", err)
		}
		if dto == nil {
			t.Error("Expected workflow DTO, got nil")
		}
		if dto.Name != "订单处理" {
			t.Errorf("Name = %v, want %v", dto.Name, "订单处理")
		}
	})
}

func TestTaskQueryService(t *testing.T) {
	t.Run("获取任务详情", func(t *testing.T) {
		taskRepo := NewMockTaskRepository()
		historyRepo := NewMockTaskHistoryRepository()
		workflowRepo := NewMockWorkflowRepository()

		// 创建任务
		task := workflow.NewTask("instance-1", "workflow-1", "审核订单", "review_order")
		taskRepo.Save(context.Background(), task)

		// 查询任务
		queryService := query.NewTaskQueryService(taskRepo, historyRepo, workflowRepo)
		dto, err := queryService.GetTaskByID(context.Background(), task.ID)
		if err != nil {
			t.Errorf("GetTaskByID() error = %v", err)
		}
		if dto == nil {
			t.Error("Expected task DTO, got nil")
		}
		if dto.TaskName != "审核订单" {
			t.Errorf("TaskName = %v, want %v", dto.TaskName, "审核订单")
		}
	})

	t.Run("列出所有任务", func(t *testing.T) {
		taskRepo := NewMockTaskRepository()
		historyRepo := NewMockTaskHistoryRepository()
		workflowRepo := NewMockWorkflowRepository()

		// 创建任务
		for i := 0; i < 3; i++ {
			task := workflow.NewTask("instance-1", "workflow-1", "任务", "task_key")
			taskRepo.Save(context.Background(), task)
		}

		// 查询任务
		queryService := query.NewTaskQueryService(taskRepo, historyRepo, workflowRepo)
		filters := make(map[string]interface{})
		tasks, total, err := queryService.ListAllTasks(context.Background(), filters, 10, 0)
		if err != nil {
			t.Errorf("ListAllTasks() error = %v", err)
		}
		if len(tasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(tasks))
		}
		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
	})

	t.Run("获取实例任务", func(t *testing.T) {
		taskRepo := NewMockTaskRepository()
		historyRepo := NewMockTaskHistoryRepository()
		workflowRepo := NewMockWorkflowRepository()

		// 创建任务
		for i := 0; i < 3; i++ {
			task := workflow.NewTask("instance-1", "workflow-1", "任务", "task_key")
			taskRepo.Save(context.Background(), task)
		}

		// 查询实例任务
		queryService := query.NewTaskQueryService(taskRepo, historyRepo, workflowRepo)
		tasks, total, err := queryService.GetInstanceTasks(context.Background(), "instance-1", 10, 0)
		if err != nil {
			t.Errorf("GetInstanceTasks() error = %v", err)
		}
		if len(tasks) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(tasks))
		}
		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
	})
}
