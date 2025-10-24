package command_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/domain/workflow"
)

// MockTaskRepository 模拟任务仓储
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

// MockTaskHistoryRepository 模拟任务历史仓储
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

func TestCreateTaskHandler(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *command.CreateTaskCommand
		wantErr bool
	}{
		{
			name: "创建有效的任务",
			cmd: &command.CreateTaskCommand{
				InstanceID:     "instance-1",
				WorkflowID:     "workflow-1",
				TaskName:       "审核订单",
				TaskKey:        "review_order",
				Description:    "审核订单信息",
				Assignee:       "user1",
				CandidateUsers: []string{"user1", "user2"},
				Priority:       "high",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockTaskRepository()
			handler := command.NewCreateTaskHandler(repo)

			id, err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && id == "" {
				t.Error("Handle() returned empty ID")
			}
		})
	}
}

func TestClaimTaskHandler(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockTaskRepository) string
		cmd     *command.ClaimTaskCommand
		wantErr bool
	}{
		{
			name: "认领待处理任务",
			setup: func(repo *MockTaskRepository) string {
				task := workflow.NewTask("instance-1", "workflow-1", "审核订单", "review_order")
				task.Status = workflow.TaskStatusPending
				repo.Save(context.Background(), task)
				return task.ID
			},
			cmd: &command.ClaimTaskCommand{
				UserID: "user1",
			},
			wantErr: false,
		},
		{
			name: "认领已认领的任务",
			setup: func(repo *MockTaskRepository) string {
				task := workflow.NewTask("instance-1", "workflow-1", "审核订单", "review_order")
				task.Status = workflow.TaskStatusClaimed
				task.Assignee = "user2"
				repo.Save(context.Background(), task)
				return task.ID
			},
			cmd: &command.ClaimTaskCommand{
				UserID: "user1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskRepo := NewMockTaskRepository()
			historyRepo := NewMockTaskHistoryRepository()
			taskID := tt.setup(taskRepo)

			handler := command.NewClaimTaskHandler(taskRepo, historyRepo)
			tt.cmd.TaskID = taskID

			err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompleteTaskHandler(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockTaskRepository) string
		cmd     *command.CompleteTaskCommand
		wantErr bool
	}{
		{
			name: "完成已认领的任务",
			setup: func(repo *MockTaskRepository) string {
				task := workflow.NewTask("instance-1", "workflow-1", "审核订单", "review_order")
				task.Status = workflow.TaskStatusClaimed
				task.Assignee = "user1"
				repo.Save(context.Background(), task)
				return task.ID
			},
			cmd: &command.CompleteTaskCommand{
				UserID:  "user1",
				Output:  `{"approved": true}`,
				Comment: "已审核",
				Result:  workflow.TaskResultApproved,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskRepo := NewMockTaskRepository()
			historyRepo := NewMockTaskHistoryRepository()
			taskID := tt.setup(taskRepo)

			handler := command.NewCompleteTaskHandler(taskRepo, historyRepo, nil)
			tt.cmd.TaskID = taskID

			err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDelegateTaskHandler(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockTaskRepository) string
		cmd     *command.DelegateTaskCommand
		wantErr bool
	}{
		{
			name: "转办已认领的任务",
			setup: func(repo *MockTaskRepository) string {
				task := workflow.NewTask("instance-1", "workflow-1", "审核订单", "review_order")
				task.Status = workflow.TaskStatusClaimed
				task.Assignee = "user1"
				repo.Save(context.Background(), task)
				return task.ID
			},
			cmd: &command.DelegateTaskCommand{
				UserID:   "user1",
				TargetID: "user2",
				Comment:  "需要专家审核",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskRepo := NewMockTaskRepository()
			historyRepo := NewMockTaskHistoryRepository()
			taskID := tt.setup(taskRepo)

			handler := command.NewDelegateTaskHandler(taskRepo, historyRepo)
			tt.cmd.TaskID = taskID

			err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteTaskHandler(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockTaskRepository) string
		wantErr bool
	}{
		{
			name: "删除待处理任务",
			setup: func(repo *MockTaskRepository) string {
				task := workflow.NewTask("instance-1", "workflow-1", "审核订单", "review_order")
				task.Status = workflow.TaskStatusPending
				repo.Save(context.Background(), task)
				return task.ID
			},
			wantErr: false,
		},
		{
			name: "删除已认领的任务",
			setup: func(repo *MockTaskRepository) string {
				task := workflow.NewTask("instance-1", "workflow-1", "审核订单", "review_order")
				task.Status = workflow.TaskStatusClaimed
				repo.Save(context.Background(), task)
				return task.ID
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockTaskRepository()
			taskID := tt.setup(repo)

			handler := command.NewDeleteTaskHandler(repo)
			cmd := &command.DeleteTaskCommand{TaskID: taskID}
			err := handler.Handle(context.Background(), cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
