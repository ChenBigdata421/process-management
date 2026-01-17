package task_aggregate

import (
	"encoding/json"
	"time"

	errors "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/status"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Task 任务领域模型
type Task struct {
	ID          string `gorm:"primaryKey" json:"id"`
	InstanceID  string `json:"instance_id"`
	WorkflowID  string `json:"workflow_id"`
	TaskName    string `json:"task_name"`
	TaskKey     string `json:"task_key"`
	Description string `json:"description"`
	TaskType    string `json:"task_type"`

	// 任务分配
	Assignee        string         `json:"assignee"`
	CandidateUsers  pq.StringArray `gorm:"type:text[]" json:"candidate_users"`
	CandidateGroups pq.StringArray `gorm:"type:text[]" json:"candidate_groups"`

	// 任务状态
	Status   status.TaskStatus   `json:"status"`
	Priority status.TaskPriority `json:"priority"`

	// 任务数据
	TaskData json.RawMessage `gorm:"type:jsonb" json:"task_data"`
	FormData json.RawMessage `gorm:"type:jsonb" json:"form_data"`
	Output   json.RawMessage `gorm:"type:jsonb" json:"output"`

	// 处理信息
	Result  status.TaskResult `json:"result"`
	Comment string            `json:"comment"`

	// 时间信息
	CreatedAt   time.Time      `json:"created_at"`
	ClaimedAt   *time.Time     `json:"claimed_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	DueDate     *time.Time     `json:"due_date"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "workflow_tasks"
}

// NewTask 创建新任务
func NewTask(instanceID, workflowID, taskName, taskKey string) *Task {
	return &Task{
		ID:         uuid.New().String(),
		InstanceID: instanceID,
		WorkflowID: workflowID,
		TaskName:   taskName,
		TaskKey:    taskKey,
		Status:     status.TaskStatusPending,
		Priority:   status.TaskPriorityMedium,
		TaskType:   "user_task",
		CreatedAt:  time.Now(),
	}
}

// Claim 认领任务
func (t *Task) Claim(userID string) error {
	if t.Status != status.TaskStatusPending {
		return errors.ErrTaskNotPending
	}

	now := time.Now()
	t.Assignee = userID
	t.Status = status.TaskStatusClaimed
	t.ClaimedAt = &now
	return nil
}

// TaskStatusPending 表示任务尚未被任何人领取，系统视其为“开放待办”，等待某人认领。
// TaskStatusClaimed 表示任务已被认领，等待处理。
// TaskStatusCompleted 表示任务已完成。
// TaskStatusRejected 表示任务被驳回。
// 前端“认领”按钮其实是把状态从 Pending 转为 Claimed，之后才进入用户视角的“待办列表”。
// 在一些场景里，任务可能由系统自动指派并直接处理，这时它仍处于 Pending 状态但可以被完成（比如自动化流程或默认处理人）。
// 如果你希望强制“必须先认领再完成”，那就把条件改成只允许 TaskStatusClaimed，或在调用前加入校验。
// Complete 完成任务
func (t *Task) Complete(output, comment string, result status.TaskResult) error {
	if t.Status != status.TaskStatusClaimed && t.Status != status.TaskStatusPending {
		return errors.ErrTaskNotClaimable
	}

	now := time.Now()
	t.Output = json.RawMessage(output)
	t.Comment = comment
	t.Result = result
	t.CompletedAt = &now

	if result == status.TaskResultRejected {
		t.Status = status.TaskStatusRejected
	} else {
		t.Status = status.TaskStatusCompleted
	}

	return nil
}

// CanBeClaimed 判断任务是否可以被认领
func (t *Task) CanBeClaimed(userID string, userGroups []string) bool {
	if t.Status != status.TaskStatusPending {
		return false
	}

	// 如果指定了处理人，只有该处理人可以认领
	if t.Assignee != "" {
		return t.Assignee == userID
	}

	// 检查候选用户
	for _, user := range t.CandidateUsers {
		if user == userID {
			return true
		}
	}

	// 检查候选组
	for _, group := range t.CandidateGroups {
		for _, userGroup := range userGroups {
			if group == userGroup {
				return true
			}
		}
	}

	return false
}

// TaskHistory 任务历史记录
type TaskHistory struct {
	ID         string            `gorm:"primaryKey" json:"id"`
	TaskID     string            `json:"task_id"`
	InstanceID string            `json:"instance_id"`
	TaskName   string            `json:"task_name"`
	Assignee   string            `json:"assignee"`
	Action     string            `json:"action"` // claim, complete, approve, reject, delegate
	Result     status.TaskResult `json:"result"`
	Comment    string            `json:"comment"`
	Output     json.RawMessage   `gorm:"type:jsonb" json:"output"`
	CreatedAt  time.Time         `json:"created_at"`
}

// TableName 指定表名
func (TaskHistory) TableName() string {
	return "workflow_task_history"
}

// NewTaskHistory 创建任务历史记录
func NewTaskHistory(taskID, instanceID, taskName, assignee, action string) *TaskHistory {
	return &TaskHistory{
		ID:         uuid.New().String(),
		TaskID:     taskID,
		InstanceID: instanceID,
		TaskName:   taskName,
		Assignee:   assignee,
		Action:     action,
		CreatedAt:  time.Now(),
	}
}
