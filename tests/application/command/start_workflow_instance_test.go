package command_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/domain/workflow"
)

func TestStartWorkflowInstanceHandler(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockWorkflowRepository) string
		cmd     *command.StartWorkflowInstanceCommand
		wantErr bool
	}{
		{
			name: "启动活跃工作流的实例",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("test", "test", "{}")
				wf.Status = workflow.StatusActive
				repo.Save(context.Background(), wf)
				return wf.ID
			},
			cmd: &command.StartWorkflowInstanceCommand{
				Input: `{"order_id": "ORD-001"}`,
			},
			wantErr: false,
		},
		{
			name: "启动不存在的工作流的实例",
			setup: func(repo *MockWorkflowRepository) string {
				return "non-existent-id"
			},
			cmd: &command.StartWorkflowInstanceCommand{
				Input: `{}`,
			},
			wantErr: true,
		},
		{
			name: "启动草稿工作流的实例",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("test", "test", "{}")
				repo.Save(context.Background(), wf)
				return wf.ID
			},
			cmd: &command.StartWorkflowInstanceCommand{
				Input: `{}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflowRepo := NewMockWorkflowRepository()
			instanceRepo := NewMockWorkflowInstanceRepository()
			workflowID := tt.setup(workflowRepo)

			handler := command.NewStartWorkflowInstanceHandler(workflowRepo, instanceRepo, nil)
			tt.cmd.WorkflowID = workflowID

			id, err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if id == "" {
					t.Error("Handle() returned empty ID")
				}
				// 验证实例已保存
				instance, err := instanceRepo.FindByID(context.Background(), id)
				if err != nil {
					t.Errorf("FindByID() error = %v", err)
				}
				if instance.WorkflowID != workflowID {
					t.Errorf("WorkflowID = %v, want %v", instance.WorkflowID, workflowID)
				}
				if instance.Status != workflow.InstanceStatusRunning {
					t.Errorf("Status = %v, want %v", instance.Status, workflow.InstanceStatusRunning)
				}
			}
		})
	}
}
