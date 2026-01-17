package command

// ActivateWorkflowCommand 激活工作流命令
type ActivateWorkflowCommand struct {
	ID string
}

// CreateWorkflowCommand 创建工作流命令
type CreateWorkflowCommand struct {
	Name        string
	Description string
	Definition  string
}

// DeleteWorkflowCommand 删除工作流命令
type DeleteWorkflowCommand struct {
	ID string
}

// FreezeWorkflowCommand 冻结工作流命令
type FreezeWorkflowCommand struct {
	ID string
}

// UpdateWorkflowCommand 更新工作流命令
type UpdateWorkflowCommand struct {
	ID          string
	Name        string
	Description string
	Definition  string
}

// WorkflowDTO 工作流数据传输对象
type WorkflowDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Definition  string `json:"definition"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
