package port

import (
	"context"

	"jxt-evidence-system/process-management/internal/application/command"
	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// ClaimTaskHandler 认领任务处理器
type TaskService interface {

	// Handle 处理完成任务命令
	CompleteTask(ctx context.Context, cmd *command.CompleteTaskCommand) error

	// Handle 处理删除任务命令
	DeleteTask(ctx context.Context, cmd *command.DeleteTaskCommand) error

	// 处理转办任务命令
	DelegateTask(ctx context.Context, cmd *command.DelegateTaskCommand) error
	// Handle 处理创建任务命令
	CreateTask(ctx context.Context, cmd *command.CreateTaskCommand) (string, error)

	// GetTaskByID 根据ID获取任务
	GetTaskByID(ctx context.Context, id valueobject.TaskID) (*task_aggregate.Task, error)

	// GetRecentTask 根据实例ID获取最近的一条任务
	GetRecentTask(ctx context.Context, instanceID valueobject.InstanceID) (*task_aggregate.Task, error)

	// GetTodoTasks 查询待办任务
	GetTodoTasks(ctx context.Context, userID int, query *command.TodoTaskPagedQuery) ([]*task_aggregate.Task, int, error)

	// GetDoneTasks 查询已办任务
	GetDoneTasks(ctx context.Context, userID int, query *command.DoneTaskPagedQuery) ([]*task_aggregate.Task, int, error)

	// GetPage 查询所有任务（支持筛选）
	GetPage(ctx context.Context, query *command.TaskPagedQuery) ([]*task_aggregate.Task, int, error)

	// GetTaskHistory 获取任务历史
	GetTaskHistory(ctx context.Context, taskID valueobject.TaskID) ([]*task_aggregate.TaskHistory, error)

	// GetInstanceTaskHistory 获取实例的任务历史
	GetInstanceTaskHistory(ctx context.Context, instanceID valueobject.InstanceID) ([]*task_aggregate.TaskHistory, error)
	// GetInstanceTasks 获取实例的所有任务（包含当前状态）
	GetTasksByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) ([]*task_aggregate.Task, error)
	CountTasksByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (int, error)
}
