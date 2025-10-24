package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待处理
	TaskStatusClaimed   TaskStatus = "claimed"   // 已认领
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
	Status   TaskStatus   `json:"status"`
	Priority TaskPriority `json:"priority"`

	// 任务数据
	TaskData json.RawMessage `gorm:"type:jsonb" json:"task_data"`
	FormData json.RawMessage `gorm:"type:jsonb" json:"form_data"`
	Output   json.RawMessage `gorm:"type:jsonb" json:"output"`

	// 处理信息
	Result  TaskResult `json:"result"`
	Comment string     `json:"comment"`

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
		Status:     TaskStatusPending,
		Priority:   TaskPriorityMedium,
		TaskType:   "user_task",
		CreatedAt:  time.Now(),
	}
}

// Claim 认领任务
func (t *Task) Claim(userID string) error {
	if t.Status != TaskStatusPending {
		return ErrTaskNotPending
	}

	now := time.Now()
	t.Assignee = userID
	t.Status = TaskStatusClaimed
	t.ClaimedAt = &now
	return nil
}

// Complete 完成任务
func (t *Task) Complete(output, comment string, result TaskResult) error {
	if t.Status != TaskStatusClaimed && t.Status != TaskStatusPending {
		return ErrTaskNotClaimable
	}

	now := time.Now()
	t.Output = json.RawMessage(output)
	t.Comment = comment
	t.Result = result
	t.CompletedAt = &now

	if result == TaskResultRejected {
		t.Status = TaskStatusRejected
	} else {
		t.Status = TaskStatusCompleted
	}

	return nil
}

// CanBeClaimed 判断任务是否可以被认领
func (t *Task) CanBeClaimed(userID string, userGroups []string) bool {
	if t.Status != TaskStatusPending {
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
	ID         string          `gorm:"primaryKey" json:"id"`
	TaskID     string          `json:"task_id"`
	InstanceID string          `json:"instance_id"`
	TaskName   string          `json:"task_name"`
	Assignee   string          `json:"assignee"`
	Action     string          `json:"action"` // claim, complete, approve, reject, delegate
	Result     TaskResult      `json:"result"`
	Comment    string          `json:"comment"`
	Output     json.RawMessage `gorm:"type:jsonb" json:"output"`
	CreatedAt  time.Time       `json:"created_at"`
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
