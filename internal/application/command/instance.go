package command

import (
	"encoding/json"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/query"
	"time"
)

// CancelInstanceCommand 删除工作流实例命令
type CancelInstanceCommand struct {
	ID valueobject.InstanceID `uri:"id" binding:"required"`
}

// DeleteInstanceCommand 删除工作流实例命令
type DeleteInstanceCommand struct {
	ID valueobject.InstanceID `uri:"id" binding:"required"`
}

func (s *DeleteInstanceCommand) GetId() valueobject.InstanceID {
	return s.ID
}

// StartWorkflowInstanceCommand 启动工作流实例命令
type StartWorkflowInstanceCommand struct {
	ID    valueobject.WorkflowID `json:"id" binding:"required"`
	Input json.RawMessage        `json:"input"`
}

// InstancePagedQuery 工作流实例分页查询命令
type InstancePagedQuery struct {
	query.Pagination `search:"-"`
	WorkflowID       valueobject.WorkflowID `form:"workflowId" search:"type:exact;column:workflow_id;table:workflow_instances"`
	Status           string                 `form:"status" search:"type:exact;column:status;table:workflow_instances"`
	StartedAtStart   *time.Time             `form:"startedAtStart" search:"type:gte;column:started_at;table:workflow_instances"`
	StartedAtEnd     *time.Time             `form:"startedAtEnd" search:"type:lte;column:started_at;table:workflow_instances"`
	CompletedAtStart *time.Time             `form:"completedAtStart" search:"type:gte;column:completed_at;table:workflow_instances"`
	CompletedAtEnd   *time.Time             `form:"completedAtEnd" search:"type:lte;column:completed_at;table:workflow_instances"`
}

func (q *InstancePagedQuery) GetNeedSearch() interface{} {
	return *q
}

type GetInstancesByWorkflowPagedQuery struct {
	query.Pagination `search:"-"`
	ID               valueobject.WorkflowID `form:"id" search:"type:exact;column:workflow_id;table:workflow_instances"`
}

func (q *GetInstancesByWorkflowPagedQuery) GetNeedSearch() interface{} {
	return *q
}

func (q *GetInstancesByWorkflowPagedQuery) GetId() valueobject.WorkflowID {
	return q.ID
}

type GetInstanceCommand struct {
	ID valueobject.InstanceID `uri:"id" binding:"required"`
}

func (s *GetInstanceCommand) GetId() valueobject.InstanceID {
	return s.ID
}
