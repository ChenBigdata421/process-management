package workflow

import "errors"

var (
	ErrInvalidStatusTransition         = errors.New("invalid workflow status transition")
	ErrCannotCancelCompletedWorkflow   = errors.New("cannot cancel completed or failed workflow")
	ErrInvalidInstanceStatusTransition = errors.New("invalid instance status transition")
	ErrCannotCancelCompletedInstance   = errors.New("cannot cancel completed or failed instance")
	ErrWorkflowNotFound                = errors.New("workflow not found")
	ErrInstanceNotFound                = errors.New("workflow instance not found")
	ErrInvalidWorkflowDefinition       = errors.New("invalid workflow definition")
)

