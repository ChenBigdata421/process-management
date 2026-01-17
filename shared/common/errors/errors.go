package errors

import "errors"

var (
	ErrInvalidStatusTransition         = errors.New("invalid workflow status transition")
	ErrCannotCancelCompletedWorkflow   = errors.New("cannot cancel completed or failed workflow")
	ErrInvalidInstanceStatusTransition = errors.New("invalid instance status transition")
	ErrCannotCancelCompletedInstance   = errors.New("cannot cancel completed or failed instance")
	ErrWorkflowNotFound                = errors.New("workflow not found")
	ErrInstanceNotFound                = errors.New("workflow instance not found")
	ErrInvalidWorkflowDefinition       = errors.New("invalid workflow definition")
	// ErrTaskNotFound 任务不存在
	ErrTaskNotFound = errors.New("task not found")

	// ErrTaskNotPending 任务不在待处理状态
	ErrTaskNotPending = errors.New("task is not in pending status")

	// ErrTaskNotClaimable 任务不可认领
	ErrTaskNotClaimable = errors.New("task is not claimable")

	// ErrTaskNotClaimed 任务未被认领
	ErrTaskNotClaimed = errors.New("task is not claimed")

	// ErrUnauthorized 无权限
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidTaskStatus 无效的任务状态
	ErrInvalidTaskStatus = errors.New("invalid task status")

	// ErrTaskAlreadyClaimed 任务已被认领
	ErrTaskAlreadyClaimed = errors.New("task is already claimed")

	// ErrTaskAlreadyCompleted 任务已完成
	ErrTaskAlreadyCompleted = errors.New("task is already completed")
)
