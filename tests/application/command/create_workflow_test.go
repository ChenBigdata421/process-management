package command_test

import (
	"context"
	"testing"

	"github.com/jxt/process-management/application/command"
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

func TestCreateWorkflowHandler(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *command.CreateWorkflowCommand
		wantErr bool
	}{
		{
			name: "创建有效的工作流",
			cmd: &command.CreateWorkflowCommand{
				Name:        "订单处理流程",
				Description: "处理订单的业务流程",
				Definition:  `{"steps": ["validate", "process", "notify"]}`,
			},
			wantErr: false,
		},
		{
			name: "创建空名称的工作流",
			cmd: &command.CreateWorkflowCommand{
				Name:        "",
				Description: "描述",
				Definition:  `{}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockWorkflowRepository()
			handler := command.NewCreateWorkflowHandler(repo)

			id, err := handler.Handle(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if id == "" {
					t.Error("Handle() returned empty ID")
				}
				// 验证工作流已保存
				wf, err := repo.FindByID(context.Background(), id)
				if err != nil {
					t.Errorf("FindByID() error = %v", err)
				}
				if wf.Name != tt.cmd.Name {
					t.Errorf("Name = %v, want %v", wf.Name, tt.cmd.Name)
				}
				if wf.Status != workflow.StatusDraft {
					t.Errorf("Status = %v, want %v", wf.Status, workflow.StatusDraft)
				}
			}
		})
	}
}
