package command

import (
	"jxt-evidence-system/process-management/shared/common/status"
)

// ClaimTaskCommand 认领任务命令
type ClaimTaskCommand struct {
	TaskID string
	UserID string
}

// CompleteTaskCommand 完成任务命令
type CompleteTaskCommand struct {
	TaskID  string
	UserID  string
	Output  string
	Comment string
	Result  status.TaskResult
}

// DelegateTaskCommand 转办任务命令
type DelegateTaskCommand struct {
	TaskID   string
	UserID   string
	TargetID string
	Comment  string
}

// DeleteTaskCommand 删除任务命令
type DeleteTaskCommand struct {
	TaskID string
}

// CreateTaskCommand 创建任务命令
type CreateTaskCommand struct {
	InstanceID      string   `json:"instance_id" binding:"required"`
	WorkflowID      string   `json:"workflow_id" binding:"required"`
	TaskName        string   `json:"task_name" binding:"required"`
	TaskKey         string   `json:"task_key" binding:"required"`
	Description     string   `json:"description"`
	Assignee        string   `json:"assignee"`
	CandidateUsers  []string `json:"candidate_users"`
	CandidateGroups []string `json:"candidate_groups"`
	Priority        string   `json:"priority"`
}

// TaskDTO 任务数据传输对象
type TaskDTO struct {
	ID              string   `json:"id"`
	InstanceID      string   `json:"instance_id"`
	WorkflowID      string   `json:"workflow_id"`
	WorkflowName    string   `json:"workflow_name"`
	TaskName        string   `json:"task_name"`
	TaskKey         string   `json:"task_key"`
	Description     string   `json:"description"`
	TaskType        string   `json:"task_type"`
	Assignee        string   `json:"assignee"`
	CandidateUsers  []string `json:"candidate_users"`
	CandidateGroups []string `json:"candidate_groups"`
	Status          string   `json:"status"`
	Priority        string   `json:"priority"`
	TaskData        string   `json:"task_data"`
	FormData        string   `json:"form_data"`
	Output          string   `json:"output"`
	Result          string   `json:"result"`
	Comment         string   `json:"comment"`
	CreatedAt       string   `json:"created_at"`
	ClaimedAt       string   `json:"claimed_at"`
	CompletedAt     string   `json:"completed_at"`
	DueDate         string   `json:"due_date"`
}

// TaskHistoryDTO 任务历史数据传输对象
type TaskHistoryDTO struct {
	ID         string `json:"id"`
	TaskID     string `json:"task_id"`
	InstanceID string `json:"instance_id"`
	TaskName   string `json:"task_name"`
	Assignee   string `json:"assignee"`
	Action     string `json:"action"`
	Result     string `json:"result"`
	Comment    string `json:"comment"`
	Output     string `json:"output"`
	CreatedAt  string `json:"created_at"`
}
