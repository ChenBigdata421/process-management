package command_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/domain/workflow"
)

func TestActivateWorkflowHandler(t *testing.T) {
	tests := []struct {
		name       string
		workflowID string
		setup      func(*MockWorkflowRepository) string
		wantErr    bool
	}{
		{
			name: "激活草稿状态的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("test", "test", "{}")
				repo.Save(context.Background(), wf)
				return wf.ID
			},
			wantErr: false,
		},
		{
			name: "激活不存在的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				return "non-existent-id"
			},
			wantErr: true,
		},
		{
			name: "激活已激活的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("test", "test", "{}")
				wf.Status = workflow.StatusActive
				repo.Save(context.Background(), wf)
				return wf.ID
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockWorkflowRepository()
			workflowID := tt.setup(repo)

			handler := command.NewActivateWorkflowHandler(repo)
			cmd := &command.ActivateWorkflowCommand{
				ID: workflowID,
			}

			err := handler.Handle(context.Background(), cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				// 验证工作流已激活
				wf, _ := repo.FindByID(context.Background(), workflowID)
				if wf.Status != workflow.StatusActive {
					t.Errorf("Status = %v, want %v", wf.Status, workflow.StatusActive)
				}
			}
		})
	}
}
