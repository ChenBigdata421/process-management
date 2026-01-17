package port

import (
	"context"

	"jxt-evidence-system/process-management/internal/application/command"
)

// ClaimTaskHandler 认领任务处理器
type TaskService interface {

	// 处理认领任务命令
	ClaimTask(ctx context.Context, cmd *command.ClaimTaskCommand) error

	// Handle 处理完成任务命令
	CompleteTask(ctx context.Context, cmd *command.CompleteTaskCommand) error

	// Handle 处理删除任务命令
	DeleteTask(ctx context.Context, cmd *command.DeleteTaskCommand) error

	// 处理转办任务命令
	DelegateTask(ctx context.Context, cmd *command.DelegateTaskCommand) error
	// Handle 处理创建任务命令
	CreateTask(ctx context.Context, cmd *command.CreateTaskCommand) (string, error)

	// GetTaskByID 根据ID获取任务
	GetTaskByID(ctx context.Context, id string) (*command.TaskDTO, error)

	// ListTodoTasks 查询待办任务
	ListTodoTasks(ctx context.Context, userID string, limit, offset int) ([]*command.TaskDTO, int64, error)

	// ListDoneTasks 查询已办任务
	ListDoneTasks(ctx context.Context, userID string, limit, offset int) ([]*command.TaskDTO, int64, error)

	// ListClaimableTasks 查询可认领的任务
	ListClaimableTasks(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*command.TaskDTO, int64, error)

	// ListTasksByInstanceID 查询实例的所有任务
	ListTasksByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*command.TaskDTO, error)

	// ListAllTasks 查询所有任务（支持多条件查询）
	ListAllTasks(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*command.TaskDTO, int64, error)

	// GetTaskHistory 获取任务历史
	GetTaskHistory(ctx context.Context, taskID string, limit, offset int) ([]*command.TaskHistoryDTO, error)

	// GetInstanceTaskHistory 获取实例的任务历史
	GetInstanceTaskHistory(ctx context.Context, instanceID string, limit, offset int) ([]*command.TaskHistoryDTO, error)
	// GetInstanceTasks 获取实例的所有任务（包含当前状态）
	GetInstanceTasks(ctx context.Context, instanceID string, limit, offset int) ([]*command.TaskDTO, int64, error)
}
