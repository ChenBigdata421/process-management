package command_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/command"
	"github.com/jxt/process-management/domain/workflow"
)

func TestDeleteWorkflowHandler(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*MockWorkflowRepository) string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "删除草稿状态的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("test", "test", "{}")
				repo.Save(context.Background(), wf)
				return wf.ID
			},
			wantErr: false,
		},
		{
			name: "删除不存在的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				return "non-existent-id"
			},
			wantErr: true,
		},
		{
			name: "删除活跃的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("test", "test", "{}")
				wf.Status = workflow.StatusActive
				repo.Save(context.Background(), wf)
				return wf.ID
			},
			wantErr: true,
		},
		{
			name: "删除已完成的工作流",
			setup: func(repo *MockWorkflowRepository) string {
				wf := workflow.NewWorkflow("test", "test", "{}")
				wf.Status = workflow.StatusCompleted
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

			handler := command.NewDeleteWorkflowHandler(repo)
			cmd := &command.DeleteWorkflowCommand{
				ID: workflowID,
			}

			err := handler.Handle(context.Background(), cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				// 验证工作流已删除
				_, err := repo.FindByID(context.Background(), workflowID)
				if err == nil {
					t.Error("Workflow should be deleted")
				}
			}
		})
	}
}
