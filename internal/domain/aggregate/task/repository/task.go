package repository

import (
	"context"
	"jxt-evidence-system/process-management/internal/application/command"
	task "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// TaskRepository 任务仓储接口
type TaskRepository interface {
	Save(ctx context.Context, task *task.Task) error
	FindByID(ctx context.Context, id valueobject.TaskID) (*task.Task, error)
	FindByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) ([]*task.Task, error)
	CountByInstanceID(ctx context.Context, id valueobject.InstanceID) (int, error)
	FindRecentByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (*task.Task, error)
	CountTasksByInstanceID(ctx context.Context, id valueobject.InstanceID) (int, error)
	FindTodoByAssignee(ctx context.Context, assignee int, query *command.TodoTaskPagedQuery) ([]*task.Task, int, error)
	FindDoneByAssignee(ctx context.Context, assignee int, query *command.DoneTaskPagedQuery) ([]*task.Task, int, error)
	GetPage(ctx context.Context, query *command.TaskPagedQuery) ([]*task.Task, int, error)
	Update(ctx context.Context, task *task.Task) error
	Delete(ctx context.Context, id valueobject.TaskID) error
}

// TaskHistoryRepository 任务历史仓储接口
type TaskHistoryRepository interface {
	Save(ctx context.Context, history *task.TaskHistory) error
	FindByTaskID(ctx context.Context, taskID valueobject.TaskID) ([]*task.TaskHistory, error)
	FindByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) ([]*task.TaskHistory, error)
}
