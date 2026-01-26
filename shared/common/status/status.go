package status

// InstanceStatus 工作流实例状态
type InstanceStatus string

const (
	InstanceStatusRunning   InstanceStatus = "running"
	InstanceStatusCompleted InstanceStatus = "completed"
	InstanceStatusFailed    InstanceStatus = "failed"
	InstanceStatusCancelled InstanceStatus = "cancelled"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待处理
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusRejected  TaskStatus = "rejected"  // 已驳回
)

// TaskPriority 任务优先级
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// TaskResult 任务处理结果
type TaskResult string

const (
	TaskResultApproved  TaskResult = "approved"  // 通过
	TaskResultRejected  TaskResult = "rejected"  // 驳回
	TaskResultCompleted TaskResult = "completed" // 完成
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	StatusDraft     WorkflowStatus = "draft"
	StatusActive    WorkflowStatus = "active"
	StatusFrozen    WorkflowStatus = "frozen"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusCancelled WorkflowStatus = "cancelled"
)
