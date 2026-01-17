package command

// DeleteInstanceCommand 删除工作流实例命令
type DeleteInstanceCommand struct {
	ID string
}

// StartWorkflowInstanceCommand 启动工作流实例命令
type StartWorkflowInstanceCommand struct {
	WorkflowID string
	Input      string
}

// WorkflowInstanceDTO 工作流实例数据传输对象
type WorkflowInstanceDTO struct {
	ID           string `json:"id"`
	WorkflowID   string `json:"workflow_id"`
	Status       string `json:"status"`
	Input        string `json:"input"`
	Output       string `json:"output"`
	ErrorMessage string `json:"error_message"`
	StartedAt    string `json:"started_at"`
	CompletedAt  string `json:"completed_at"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
