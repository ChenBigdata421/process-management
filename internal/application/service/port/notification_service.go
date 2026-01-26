package port

import (
	"context"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
)

// NotificationService 通知服务接口
type NotificationService interface {
	// NotifyTaskCreated 通知任务已创建
	NotifyTaskCreated(ctx context.Context, task *task_aggregate.Task)

	// NotifyTaskAssigned 通知任务已分配
	NotifyTaskAssigned(ctx context.Context, task *task_aggregate.Task, assignee int)

	// NotifyTaskCompleted 通知任务已完成
	NotifyTaskCompleted(ctx context.Context, task *task_aggregate.Task)

	// NotifyWorkflowCompleted 通知工作流已完成
	NotifyWorkflowCompleted(ctx context.Context, instance *instance_aggregate.WorkflowInstance)
}
