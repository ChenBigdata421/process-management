package command_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/domain/workflow"
)

func TestUpdateWorkflowHandler(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockWorkflowRepository) string
		cmd     *command.UpdateWorkflowCommand
		wantErr bool
	}{
		{
			name: "更新草稿状态的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("old name", "old desc", "{}")
				repo.Save(context.Background(), wf)
				return wf.ID
			},
			cmd: &command.UpdateWorkflowCommand{
				Name:        "new name",
				Description: "new desc",
				Definition:  `{"new": "definition"}`,
			},
			wantErr: false,
		},
		{
			name: "更新不存在的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				return "non-existent-id"
			},
			cmd: &command.UpdateWorkflowCommand{
				Name: "new name",
			},
			wantErr: true,
		},
		{
			name: "更新活跃的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("test", "test", "{}")
				wf.Status = workflow.StatusActive
				repo.Save(context.Background(), wf)
				return wf.ID
			},
			cmd: &command.UpdateWorkflowCommand{
				Name: "new name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockWorkflowRepository()
			workflowID := tt.setup(repo)

			handler := command.NewUpdateWorkflowHandler(repo)
			tt.cmd.ID = workflowID

			err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				// 验证工作流已更新
				wf, _ := repo.FindByID(context.Background(), workflowID)
				if wf.Name != tt.cmd.Name {
					t.Errorf("Name = %v, want %v", wf.Name, tt.cmd.Name)
				}
				if wf.Description != tt.cmd.Description {
					t.Errorf("Description = %v, want %v", wf.Description, tt.cmd.Description)
				}
			}
		})
	}
}
