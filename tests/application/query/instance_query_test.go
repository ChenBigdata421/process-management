package query_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/query"
	"github.com/jxt/process-management/domain/workflow"
)

// MockWorkflowRepository 是一个模拟的工作流仓储
type MockWorkflowRepository struct {
	workflows map[string]*workflow.Workflow
}

func NewMockWorkflowRepository() *MockWorkflowRepository {
	return &MockWorkflowRepository{
		workflows: make(map[string]*workflow.Workflow),
	}
}

func (m *MockWorkflowRepository) Save(ctx context.Context, wf *workflow.Workflow) error {
	m.workflows[wf.ID] = wf
	return nil
}

func (m *MockWorkflowRepository) FindByID(ctx context.Context, id string) (*workflow.Workflow, error) {
	if wf, ok := m.workflows[id]; ok {
		return wf, nil
	}
	return nil, workflow.ErrWorkflowNotFound
}

func (m *MockWorkflowRepository) FindAll(ctx context.Context, limit, offset int) ([]*workflow.Workflow, error) {
	var result []*workflow.Workflow
	for _, wf := range m.workflows {
		result = append(result, wf)
	}
	return result, nil
}

func (m *MockWorkflowRepository) Update(ctx context.Context, wf *workflow.Workflow) error {
	m.workflows[wf.ID] = wf
	return nil
}

func (m *MockWorkflowRepository) Delete(ctx context.Context, id string) error {
	delete(m.workflows, id)
	return nil
}

func (m *MockWorkflowRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.workflows)), nil
}

// MockWorkflowInstanceRepository 是一个模拟的工作流实例仓储
type MockWorkflowInstanceRepository struct {
	instances map[string]*workflow.WorkflowInstance
}

func NewMockWorkflowInstanceRepository() *MockWorkflowInstanceRepository {
	return &MockWorkflowInstanceRepository{
		instances: make(map[string]*workflow.WorkflowInstance),
	}
}

func (m *MockWorkflowInstanceRepository) Save(ctx context.Context, instance *workflow.WorkflowInstance) error {
	m.instances[instance.ID] = instance
	return nil
}

func (m *MockWorkflowInstanceRepository) FindByID(ctx context.Context, id string) (*workflow.WorkflowInstance, error) {
	if instance, ok := m.instances[id]; ok {
		return instance, nil
	}
	return nil, workflow.ErrInstanceNotFound
}

func (m *MockWorkflowInstanceRepository) FindByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*workflow.WorkflowInstance, error) {
	var result []*workflow.WorkflowInstance
	for _, instance := range m.instances {
		if instance.WorkflowID == workflowID {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (m *MockWorkflowInstanceRepository) Update(ctx context.Context, instance *workflow.WorkflowInstance) error {
	m.instances[instance.ID] = instance
	return nil
}

func (m *MockWorkflowInstanceRepository) Delete(ctx context.Context, id string) error {
	delete(m.instances, id)
	return nil
}

func (m *MockWorkflowInstanceRepository) CountByWorkflowID(ctx context.Context, workflowID string) (int64, error) {
	count := int64(0)
	for _, instance := range m.instances {
		if instance.WorkflowID == workflowID {
			count++
		}
	}
	return count, nil
}

func (m *MockWorkflowInstanceRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*workflow.WorkflowInstance, int, error) {
	var result []*workflow.WorkflowInstance
	for _, instance := range m.instances {
		result = append(result, instance)
	}
	return result, len(result), nil
}

func TestWorkflowInstanceQueryService_GetInstance(t *testing.T) {
	repo := NewMockWorkflowInstanceRepository()
	instance := workflow.NewWorkflowInstance("wf-001", `{"order_id": "ORD-001"}`)
	repo.Save(context.Background(), instance)

	service := query.NewWorkflowInstanceQueryService(repo)

	dto, err := service.GetInstanceByID(context.Background(), instance.ID)
	if err != nil {
		t.Errorf("GetInstance() error = %v", err)
	}
	if dto == nil {
		t.Fatal("GetInstance() returned nil")
	}
	if dto.ID != instance.ID {
		t.Errorf("ID = %v, want %v", dto.ID, instance.ID)
	}
	if dto.WorkflowID != instance.WorkflowID {
		t.Errorf("WorkflowID = %v, want %v", dto.WorkflowID, instance.WorkflowID)
	}
	if dto.Status != string(instance.Status) {
		t.Errorf("Status = %v, want %v", dto.Status, instance.Status)
	}
}

func TestWorkflowInstanceQueryService_ListInstances(t *testing.T) {
	repo := NewMockWorkflowInstanceRepository()
	workflowID := "wf-001"

	// 创建多个实例
	for i := 0; i < 3; i++ {
		instance := workflow.NewWorkflowInstance(workflowID, "{}")
		repo.Save(context.Background(), instance)
	}

	service := query.NewWorkflowInstanceQueryService(repo)

	dtos, err := service.ListInstancesByWorkflowID(context.Background(), workflowID, 10, 0)
	if err != nil {
		t.Errorf("ListInstances() error = %v", err)
	}
	if len(dtos) != 3 {
		t.Errorf("ListInstances() returned %d instances, want 3", len(dtos))
	}
}

func TestWorkflowInstanceQueryService_GetInstanceNotFound(t *testing.T) {
	repo := NewMockWorkflowInstanceRepository()
	service := query.NewWorkflowInstanceQueryService(repo)

	_, err := service.GetInstanceByID(context.Background(), "non-existent-id")
	if err == nil {
		t.Error("GetInstanceByID() should return error for non-existent instance")
	}
}
