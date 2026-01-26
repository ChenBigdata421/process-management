package port

import (
	"context"

	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
)

// ClaimTaskHandler 认领任务处理器
type WorkflowEngineService interface {
	StartInstance(ctx context.Context, instanceID valueobject.InstanceID) error
	ContinueAfterTask(ctx context.Context, task *task_aggregate.Task) error
	RejectAndGoBack(ctx context.Context, task *task_aggregate.Task) error
}
