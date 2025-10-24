package workflow_test

import (
	"testing"
	"time"

	"github.com/jxt/process-management/domain/workflow"
)

func TestNewWorkflowInstance(t *testing.T) {
	tests := []struct {
		name       string
		workflowID string
		input      string
		wantErr    bool
	}{
		{
			name:       "创建有效的工作流实例",
			workflowID: "wf-001",
			input:      `{"order_id": "ORD-001"}`,
			wantErr:    false,
		},
		{
			name:       "创建空工作流ID的实例",
			workflowID: "",
			input:      `{}`,
			wantErr:    false, // 当前实现允许空ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := workflow.NewWorkflowInstance(tt.workflowID, tt.input)
			if instance == nil {
				t.Fatal("NewWorkflowInstance returned nil")
			}
			if instance.WorkflowID != tt.workflowID {
				t.Errorf("WorkflowID = %v, want %v", instance.WorkflowID, tt.workflowID)
			}
			if string(instance.Input) != tt.input {
				t.Errorf("Input = %v, want %v", string(instance.Input), tt.input)
			}
			if instance.Status != workflow.InstanceStatusRunning {
				t.Errorf("Status = %v, want %v", instance.Status, workflow.InstanceStatusRunning)
			}
		})
	}
}

func TestInstanceComplete(t *testing.T) {
	tests := []struct {
		name    string
		status  workflow.InstanceStatus
		wantErr bool
	}{
		{
			name:    "完成运行中的实例",
			status:  workflow.InstanceStatusRunning,
			wantErr: false,
		},
		{
			name:    "完成已完成的实例",
			status:  workflow.InstanceStatusCompleted,
			wantErr: true,
		},
		{
			name:    "完成已失败的实例",
			status:  workflow.InstanceStatusFailed,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := workflow.NewWorkflowInstance("wf-001", "{}")
			instance.Status = tt.status

			err := instance.Complete(`{"result": "success"}`)
			if (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if instance.Status != workflow.InstanceStatusCompleted {
					t.Errorf("Status = %v, want %v", instance.Status, workflow.InstanceStatusCompleted)
				}
				if string(instance.Output) != `{"result": "success"}` {
					t.Errorf("Output = %v, want %v", string(instance.Output), `{"result": "success"}`)
				}
				if instance.CompletedAt == nil {
					t.Error("CompletedAt should not be nil")
				}
			}
		})
	}
}

func TestInstanceFail(t *testing.T) {
	instance := workflow.NewWorkflowInstance("wf-001", "{}")
	instance.Status = workflow.InstanceStatusRunning

	err := instance.Fail("处理失败")
	if err != nil {
		t.Errorf("Fail() error = %v", err)
	}
	if instance.Status != workflow.InstanceStatusFailed {
		t.Errorf("Status = %v, want %v", instance.Status, workflow.InstanceStatusFailed)
	}
	if instance.ErrorMessage != "处理失败" {
		t.Errorf("ErrorMessage = %v, want %v", instance.ErrorMessage, "处理失败")
	}
	if instance.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestInstanceCancel(t *testing.T) {
	tests := []struct {
		name    string
		status  workflow.InstanceStatus
		wantErr bool
	}{
		{
			name:    "取消运行中的实例",
			status:  workflow.InstanceStatusRunning,
			wantErr: false,
		},
		{
			name:    "取消已完成的实例",
			status:  workflow.InstanceStatusCompleted,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := workflow.NewWorkflowInstance("wf-001", "{}")
			instance.Status = tt.status

			err := instance.Cancel()
			if (err != nil) != tt.wantErr {
				t.Errorf("Cancel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && instance.Status != workflow.InstanceStatusCancelled {
				t.Errorf("Status = %v, want %v", instance.Status, workflow.InstanceStatusCancelled)
			}
		})
	}
}

func TestInstanceTimestamps(t *testing.T) {
	before := time.Now()
	instance := workflow.NewWorkflowInstance("wf-001", "{}")
	after := time.Now()

	if instance.StartedAt.Before(before) || instance.StartedAt.After(after) {
		t.Errorf("StartedAt = %v, should be between %v and %v", instance.StartedAt, before, after)
	}
	if instance.CompletedAt != nil {
		t.Error("CompletedAt should be nil for running instance")
	}
}
