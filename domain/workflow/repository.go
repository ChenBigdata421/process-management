package workflow

import "context"

// WorkflowRepository 工作流仓储接口
type WorkflowRepository interface {
	Save(ctx context.Context, workflow *Workflow) error
	FindByID(ctx context.Context, id string) (*Workflow, error)
	FindAll(ctx context.Context, limit, offset int) ([]*Workflow, error)
	Update(ctx context.Context, workflow *Workflow) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

// WorkflowInstanceRepository 工作流实例仓储接口
type WorkflowInstanceRepository interface {
	Save(ctx context.Context, instance *WorkflowInstance) error
	FindByID(ctx context.Context, id string) (*WorkflowInstance, error)
	FindByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*WorkflowInstance, error)
	FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*WorkflowInstance, int, error)
	Update(ctx context.Context, instance *WorkflowInstance) error
	Delete(ctx context.Context, id string) error
	CountByWorkflowID(ctx context.Context, workflowID string) (int64, error)
}

// TaskRepository 任务仓储接口
type TaskRepository interface {
	Save(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*Task, error)
	FindTodoByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*Task, int64, error)
	FindDoneByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*Task, int64, error)
	FindClaimable(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*Task, int64, error)
	FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*Task, int64, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
}

// TaskHistoryRepository 任务历史仓储接口
type TaskHistoryRepository interface {
	Save(ctx context.Context, history *TaskHistory) error
	FindByTaskID(ctx context.Context, taskID string, limit, offset int) ([]*TaskHistory, error)
	FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*TaskHistory, error)
}

