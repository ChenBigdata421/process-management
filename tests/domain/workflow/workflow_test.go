package workflow_test

import (
	"testing"
	"time"

	"github.com/jxt/process-management/domain/workflow"
)

func TestNewWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		wfName      string
		description string
		definition  string
		wantErr     bool
	}{
		{
			name:        "创建有效的工作流",
			wfName:      "订单处理流程",
			description: "处理订单的业务流程",
			definition:  `{"steps": ["validate", "process", "notify"]}`,
			wantErr:     false,
		},
		{
			name:        "创建空名称的工作流",
			wfName:      "",
			description: "描述",
			definition:  `{}`,
			wantErr:     false, // 当前实现允许空名称
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := workflow.NewWorkflow(tt.wfName, tt.description, tt.definition)
			if wf == nil {
				t.Fatal("NewWorkflow returned nil")
			}
			if wf.Name != tt.wfName {
				t.Errorf("Name = %v, want %v", wf.Name, tt.wfName)
			}
			if wf.Description != tt.description {
				t.Errorf("Description = %v, want %v", wf.Description, tt.description)
			}
			if wf.Definition != tt.definition {
				t.Errorf("Definition = %v, want %v", wf.Definition, tt.definition)
			}
			if wf.Status != workflow.StatusDraft {
				t.Errorf("Status = %v, want %v", wf.Status, workflow.StatusDraft)
			}
		})
	}
}

func TestWorkflowActivate(t *testing.T) {
	tests := []struct {
		name    string
		status  workflow.WorkflowStatus
		wantErr bool
		wantMsg string
	}{
		{
			name:    "激活草稿状态的工作流",
			status:  workflow.StatusDraft,
			wantErr: false,
		},
		{
			name:    "激活已激活的工作流",
			status:  workflow.StatusActive,
			wantErr: true,
			wantMsg: "invalid workflow status transition",
		},
		{
			name:    "激活已完成的工作流",
			status:  workflow.StatusCompleted,
			wantErr: true,
			wantMsg: "invalid workflow status transition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := workflow.NewWorkflow("test", "test", "{}")
			wf.Status = tt.status

			err := wf.Activate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Activate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantMsg != "" && err.Error() != tt.wantMsg {
				t.Errorf("Activate() error message = %v, want %v", err.Error(), tt.wantMsg)
			}
			if err == nil && wf.Status != workflow.StatusActive {
				t.Errorf("Status = %v, want %v", wf.Status, workflow.StatusActive)
			}
		})
	}
}

func TestWorkflowComplete(t *testing.T) {
	tests := []struct {
		name    string
		status  workflow.WorkflowStatus
		wantErr bool
	}{
		{
			name:    "完成活跃的工作流",
			status:  workflow.StatusActive,
			wantErr: false,
		},
		{
			name:    "完成草稿状态的工作流",
			status:  workflow.StatusDraft,
			wantErr: true,
		},
		{
			name:    "完成已完成的工作流",
			status:  workflow.StatusCompleted,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := workflow.NewWorkflow("test", "test", "{}")
			wf.Status = tt.status

			err := wf.Complete()
			if (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && wf.Status != workflow.StatusCompleted {
				t.Errorf("Status = %v, want %v", wf.Status, workflow.StatusCompleted)
			}
		})
	}
}

func TestWorkflowFail(t *testing.T) {
	wf := workflow.NewWorkflow("test", "test", "{}")
	wf.Status = workflow.StatusActive

	err := wf.Fail()
	if err != nil {
		t.Errorf("Fail() error = %v", err)
	}
	if wf.Status != workflow.StatusFailed {
		t.Errorf("Status = %v, want %v", wf.Status, workflow.StatusFailed)
	}
}

func TestWorkflowCancel(t *testing.T) {
	tests := []struct {
		name    string
		status  workflow.WorkflowStatus
		wantErr bool
	}{
		{
			name:    "取消活跃的工作流",
			status:  workflow.StatusActive,
			wantErr: false,
		},
		{
			name:    "取消已完成的工作流",
			status:  workflow.StatusCompleted,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := workflow.NewWorkflow("test", "test", "{}")
			wf.Status = tt.status

			err := wf.Cancel()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cancel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && wf.Status != workflow.StatusCancelled {
				t.Errorf("Status = %v, want %v", wf.Status, workflow.StatusCancelled)
			}
		})
	}
}

func TestWorkflowTimestamps(t *testing.T) {
	before := time.Now()
	wf := workflow.NewWorkflow("test", "test", "{}")
	after := time.Now()

	if wf.CreatedAt.Before(before) || wf.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, should be between %v and %v", wf.CreatedAt, before, after)
	}
	if wf.UpdatedAt.Before(before) || wf.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt = %v, should be between %v and %v", wf.UpdatedAt, before, after)
	}
}

