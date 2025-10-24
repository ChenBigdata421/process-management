package query_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/query"
	"github.com/jxt/process-management/domain/workflow"
)

func TestWorkflowQueryService_GetWorkflow(t *testing.T) {
	repo := NewMockWorkflowRepository()
	wf := workflow.NewWorkflow("test", "test desc", "{}")
	repo.Save(context.Background(), wf)

	service := query.NewWorkflowQueryService(repo)

	dto, err := service.GetWorkflowByID(context.Background(), wf.ID)
	if err != nil {
		t.Errorf("GetWorkflow() error = %v", err)
	}
	if dto == nil {
		t.Fatal("GetWorkflow() returned nil")
	}
	if dto.ID != wf.ID {
		t.Errorf("ID = %v, want %v", dto.ID, wf.ID)
	}
	if dto.Name != wf.Name {
		t.Errorf("Name = %v, want %v", dto.Name, wf.Name)
	}
	if dto.Status != string(wf.Status) {
		t.Errorf("Status = %v, want %v", dto.Status, wf.Status)
	}
}

func TestWorkflowQueryService_ListWorkflows(t *testing.T) {
	repo := NewMockWorkflowRepository()

	// 创建多个工作流
	for i := 0; i < 3; i++ {
		wf := workflow.NewWorkflow("test", "test", "{}")
		repo.Save(context.Background(), wf)
	}

	service := query.NewWorkflowQueryService(repo)

	dtos, err := service.ListWorkflows(context.Background(), 10, 0)
	if err != nil {
		t.Errorf("ListWorkflows() error = %v", err)
	}
	if len(dtos) != 3 {
		t.Errorf("ListWorkflows() returned %d workflows, want 3", len(dtos))
	}
}

func TestWorkflowQueryService_GetWorkflowNotFound(t *testing.T) {
	repo := NewMockWorkflowRepository()
	service := query.NewWorkflowQueryService(repo)

	_, err := service.GetWorkflowByID(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("GetWorkflowByID() should return error for non-existent workflow")
	}
}
