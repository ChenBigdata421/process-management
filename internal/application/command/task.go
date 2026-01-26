package command

import (
	"encoding/json"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/query"
	"jxt-evidence-system/process-management/shared/common/status"
)

// CompleteTaskCommand 完成任务命令
type CompleteTaskCommand struct {
	ID               valueobject.TaskID `uri:"id" binding:"required"`
	UserID           int
	Output           json.RawMessage   `json:"output"`
	Comment          string            `json:"comment"`
	NextTaskApprover int               `json:"nextTaskApprover"`
	Result           status.TaskResult `json:"result"`
}

// UnmarshalJSON 自定义 JSON 解组，处理字符串化的 output
func (c *CompleteTaskCommand) UnmarshalJSON(data []byte) error {
	type Alias CompleteTaskCommand
	aux := &struct {
		Output interface{} `json:"output"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 处理 output 字段
	if aux.Output != nil {
		switch v := aux.Output.(type) {
		case string:
			// 如果是字符串，直接转换为 RawMessage
			c.Output = json.RawMessage(v)
		case json.RawMessage:
			c.Output = v
		default:
			// 如果是对象，转换为 JSON 字节
			b, err := json.Marshal(v)
			if err != nil {
				return err
			}
			c.Output = json.RawMessage(b)
		}
	}
	return nil
}

// DelegateTaskCommand 转办任务命令
type DelegateTaskCommand struct {
	ID       valueobject.TaskID `uri:"id" binding:"required"`
	UserID   int                `json:"userId"` //uri有required，json如有equired，会致报错，所以前端做约束即可
	TargetID int                `json:"targetId"`
	Comment  string             `json:"comment"`
}

// DeleteTaskCommand 删除任务命令
type DeleteTaskCommand struct {
	ID valueobject.TaskID `uri:"id" binding:"required"`
}

// GetTasksByInstanceIDPagedQuery 获取实例任务分页查询命令
type GetTasksByInstanceID struct {
	ID valueobject.InstanceID `uri:"instanceId" binding:"required"`
}

// GetTaskByIDCommand 获取任务命令
type GetTaskByIDCommand struct {
	ID valueobject.TaskID `uri:"id" binding:"required"`
}

type GetRecentTaskCommand struct {
	ID valueobject.InstanceID `uri:"instanceId" binding:"required"`
}

// CreateTaskCommand 创建任务命令
type CreateTaskCommand struct {
	InstanceID      valueobject.InstanceID `json:"instanceId" binding:"required"`
	WorkflowID      valueobject.WorkflowID `json:"workflowId" binding:"required"`
	TaskName        string                 `json:"taskName" binding:"required"`
	TaskKey         string                 `json:"taskKey" binding:"required"`
	Description     string                 `json:"description"`
	Assignee        int                    `json:"assignee"`
	CandidateUsers  []int64                `json:"candidateUsers"`
	CandidateGroups []int64                `json:"candidateGroups"`
	Priority        string                 `json:"priority"`
}

// TaskPagedQuery 任务分页查询
type TaskPagedQuery struct {
	query.Pagination `search:"-"`
	TaskName         string                 `form:"taskName" search:"type:contains;column:task_name;table:workflow_tasks"`
	WorkflowId       valueobject.WorkflowID `form:"workflowId" search:"type:exact;column:workflow_id;table:workflow_tasks"`
	InstanceId       valueobject.InstanceID `form:"instanceId" search:"type:exact;column:instance_id;table:workflow_tasks"`
	Status           status.TaskStatus      `form:"status" search:"type:exact;column:status;table:workflow_tasks"`
	Assignee         int                    `form:"assignee" search:"type:exact;column:assignee;table:workflow_tasks"`
}

func (q *TaskPagedQuery) GetNeedSearch() interface{} {
	return *q
}

// TodoTaskPagedQuery 待办任务分页查询
type TodoTaskPagedQuery struct {
	query.Pagination `search:"-"`
	TaskName         string `form:"taskName" search:"type:contains;column:task_name;table:tasks"`
	WorkflowName     string `form:"workflowName"`
}

type ClaimableTaskPagedQuery struct {
	query.Pagination `search:"-"`
	TaskName         string `form:"taskName" search:"type:contains;column:task_name;table:tasks"`
	WorkflowName     string `form:"workflowName"`
}

func (q *TodoTaskPagedQuery) GetNeedSearch() interface{} {
	return *q
}

// TodoTaskPagedQuery 待办任务分页查询
type DoneTaskPagedQuery struct {
	query.Pagination `search:"-"`
	TaskName         string `form:"taskName" search:"type:contains;column:task_name;table:tasks"`
	WorkflowName     string `form:"workflowName"`
}

func (q *DoneTaskPagedQuery) GetNeedSearch() interface{} {
	return *q
}

type TaskHistory struct {
	ID valueobject.TaskID `uri:"id" binding:"required"`
}

func (q *TaskHistory) GetNeedSearch() interface{} {
	return *q
}

// TaskHistoryItem 任务历史记录项
type TaskHistoryItem struct {
	TaskName    string                 `json:"taskName"`
	TaskKey     string                 `json:"taskKey"`
	Assignee    int                    `json:"assignee"`
	Status      string                 `json:"status"`
	Result      string                 `json:"result"`
	Comment     string                 `json:"comment"`
	Output      map[string]interface{} `json:"output"`
	CompletedAt string                 `json:"completedAt"`
	CreatedAt   string                 `json:"createdAt"`
}
