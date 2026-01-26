package task_aggregate

import (
	"encoding/json"
	"time"

	"jxt-evidence-system/process-management/internal/application/command"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	workflow_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	errors "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/models"
	"jxt-evidence-system/process-management/shared/common/status"
)

// Task 任务领域模型
type Task struct {
	TaskID      valueobject.TaskID                  `json:"taskId" gorm:"primaryKey;column:id;type:uuid;comment:主键编码"`
	InstanceID  valueobject.InstanceID              `json:"instanceId" gorm:"column:instance_id;type:uuid;index;comment:实例编码"`
	WorkflowID  valueobject.WorkflowID              `json:"workflowId" gorm:"column:workflow_id;type:uuid;comment:工作流编码"`
	TaskNo      string                              `json:"taskNo"`
	InstaceNo   string                              `json:"instanceNo" gorm:"-"`
	WorkflowNo  string                              `json:"workflowNo" gorm:"-"`
	WorklowName string                              `json:"workflowName" gorm:"-"`
	TaskName    string                              `json:"taskName"`
	TaskKey     string                              `json:"taskKey"`
	Description string                              `json:"description"`
	TaskType    string                              `json:"taskType"`
	Workflow    workflow_aggregate.Workflow         `json:"-"`
	Instance    instance_aggregate.WorkflowInstance `json:"-"`
	// 任务分配
	Assignee int `json:"assignee"`

	// 任务状态
	Status   status.TaskStatus   `json:"status"`
	Priority status.TaskPriority `json:"priority"`

	// 任务数据
	TaskData json.RawMessage `gorm:"type:jsonb" json:"taskData"`
	FormData json.RawMessage `gorm:"type:jsonb" json:"formData"`
	Output   json.RawMessage `gorm:"type:jsonb" json:"output"`

	// 处理信息
	Result  status.TaskResult `json:"result"`
	Comment string            `json:"comment"`

	// 时间信息
	ClaimedAt   *time.Time `json:"claimedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	DueDate     *time.Time `json:"dueDate"`

	// 审计字段
	models.ControlBy
	models.ModelTime
}

// TableName 指定表名
func (Task) TableName() string {
	return "workflow_tasks"
}

// NewTask 创建新任务
func NewTask(instanceID valueobject.InstanceID, workflowID valueobject.WorkflowID) *Task {
	return &Task{
		TaskID:     valueobject.NewTaskID(),
		TaskNo:     "task-" + time.Now().Format("20060102150405"),
		InstanceID: instanceID,
		WorkflowID: workflowID,
		Status:     status.TaskStatusPending,
		ModelTime: models.ModelTime{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// TaskStatusPending 表示任务尚未被任何人领取，系统视其为“开放待办”，等待某人待办
// TaskStatusCompleted 表示任务已完成。
// TaskStatusRejected 表示任务被驳回。
// Complete 完成任务
func (t *Task) Complete(cmd *command.CompleteTaskCommand) error {
	if t.Status != status.TaskStatusPending {
		return errors.ErrTaskNotPending
	}

	now := time.Now()
	t.Output = cmd.Output
	t.Comment = cmd.Comment
	t.CompletedAt = &now
	t.Result = cmd.Result

	if cmd.Result == status.TaskResultRejected {
		t.Status = status.TaskStatusRejected
	} else {
		t.Status = status.TaskStatusCompleted
	}

	return nil
}

// CanBeClaimed 判断任务是否可以被认领
func (t *Task) CanBeClaimed(userID int, userGroups []int) bool {
	if t.Status != status.TaskStatusPending {
		return false
	}

	// 如果指定了处理人，只有该处理人可以认领
	if t.Assignee != 0 {
		return t.Assignee == userID
	}

	return false
}

// TaskHistory 任务历史记录
type TaskHistory struct {
	ID          valueobject.TaskHistoryID `json:"id" gorm:"column:id;type:uuid;comment:主键编码"`
	TaskID      valueobject.TaskID        `json:"taskId" gorm:"column:task_id;type:uuid"`
	InstanceID  valueobject.InstanceID    `json:"instanceId" gorm:"column:instance_id;type:uuid"`
	TaskName    string                    `json:"taskName"`
	Assignee    string                    `json:"assignee"`
	Action      string                    `json:"action"` // claim, complete, approve, reject, delegate
	Result      status.TaskResult         `json:"result"`
	Comment     string                    `json:"comment"`
	Output      json.RawMessage           `gorm:"type:jsonb" json:"output"`
	CreatedAt   time.Time                 `json:"createdAt"`
	CompletedAt time.Time                 `json:"completedAt"`
}

// TableName 指定表名
func (TaskHistory) TableName() string {
	return "workflow_task_history"
}

// NewTaskHistory 创建任务历史记录
func NewTaskHistory(taskID valueobject.TaskID, instanceID valueobject.InstanceID, taskName, assignee, action string) *TaskHistory {
	return &TaskHistory{
		ID:         valueobject.NewTaskHistoryID(),
		TaskID:     taskID,
		InstanceID: instanceID,
		TaskName:   taskName,
		Assignee:   assignee,
		Action:     action,
		CreatedAt:  time.Now(),
	}
}
