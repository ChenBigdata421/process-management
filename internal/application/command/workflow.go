package command

import (
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	common "jxt-evidence-system/process-management/shared/common/models"
	"jxt-evidence-system/process-management/shared/common/query"
	"jxt-evidence-system/process-management/shared/common/status"
)

// ActivateWorkflowCommand 激活工作流命令
type ActivateWorkflowCommand struct {
	ID valueobject.WorkflowID `uri:"id" binding:"required"`
}

// CreateWorkflowCommand 创建工作流命令
type CreateWorkflowCommand struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Definition  string `json:"definition" binding:"required"`
	common.ControlBy
}

// DeleteWorkflowCommand 删除工作流命令
type DeleteWorkflowCommand struct {
	ID valueobject.WorkflowID `uri:"id" binding:"required"`
}

// FreezeWorkflowCommand 冻结工作流命令
type FreezeWorkflowCommand struct {
	ID valueobject.WorkflowID `uri:"id" binding:"required"`
}

// CheckCanFreezeCommand 检查工作流是否可以冻结命令
type CheckCanFreezeCommand struct {
	ID valueobject.WorkflowID `uri:"id" binding:"required"`
}

// UpdateWorkflowCommand 更新工作流命令
type UpdateWorkflowCommand struct {
	ID          valueobject.WorkflowID `uri:"id" binding:"required"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Definition  string                 `json:"definition"`
	common.ControlBy
}

// GetWorkflowByIDCommand 获取工作流命令
type GetWorkflowByIDCommand struct {
	ID valueobject.WorkflowID `uri:"id" binding:"required"`
}

type GetWorkflowByNameCommand struct {
	Name string `uri:"name" binding:"required"`
}

type WorkflowPagedQuery struct {
	query.Pagination `search:"-"`
	Name             string                `form:"name" search:"type:contains;column:name;table:workflows"`
	Status           status.WorkflowStatus `form:"status" search:"type:exact;column:status;table:workflows"`
}

func (q *WorkflowPagedQuery) GetNeedSearch() interface{} {
	return *q
}
