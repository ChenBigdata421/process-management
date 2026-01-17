package port

import (
	"context"

	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
)

// ClaimTaskHandler 认领任务处理器
type WorkflowEngineService interface {
	StartInstance(ctx context.Context, instanceID string) error
	ContinueAfterTask(ctx context.Context, task *task_aggregate.Task) error
	RejectAndGoBack(ctx context.Context, task *task_aggregate.Task) error
}
