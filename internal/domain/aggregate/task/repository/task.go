package repository

import (
	"context"
	task "jxt-evidence-system/process-management/internal/domain/aggregate/task"
)

// TaskRepository 任务仓储接口
type TaskRepository interface {
	Save(ctx context.Context, task *task.Task) error
	FindByID(ctx context.Context, id string) (*task.Task, error)
	FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*task.Task, error)
	FindTodoByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*task.Task, int64, error)
	FindDoneByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*task.Task, int64, error)
	FindClaimable(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*task.Task, int64, error)
	FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*task.Task, int64, error)
	Update(ctx context.Context, task *task.Task) error
	Delete(ctx context.Context, id string) error
}

// TaskHistoryRepository 任务历史仓储接口
type TaskHistoryRepository interface {
	Save(ctx context.Context, history *task.TaskHistory) error
	FindByTaskID(ctx context.Context, taskID string, limit, offset int) ([]*task.TaskHistory, error)
	FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*task.TaskHistory, error)
}
