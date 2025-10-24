package query

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

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

// WorkflowInstanceQueryService 工作流实例查询服务
type WorkflowInstanceQueryService struct {
	repo workflow.WorkflowInstanceRepository
}

// NewWorkflowInstanceQueryService 创建查询服务
func NewWorkflowInstanceQueryService(repo workflow.WorkflowInstanceRepository) *WorkflowInstanceQueryService {
	return &WorkflowInstanceQueryService{repo: repo}
}

// GetInstanceByID 根据ID获取实例
func (qs *WorkflowInstanceQueryService) GetInstanceByID(ctx context.Context, id string) (*WorkflowInstanceDTO, error) {
	instance, err := qs.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if instance == nil {
		return nil, workflow.ErrInstanceNotFound
	}

	completedAt := ""
	if instance.CompletedAt != nil {
		completedAt = instance.CompletedAt.String()
	}

	return &WorkflowInstanceDTO{
		ID:           instance.ID,
		WorkflowID:   instance.WorkflowID,
		Status:       string(instance.Status),
		Input:        string(instance.Input),
		Output:       string(instance.Output),
		ErrorMessage: instance.ErrorMessage,
		StartedAt:    instance.StartedAt.String(),
		CompletedAt:  completedAt,
		CreatedAt:    instance.CreatedAt.String(),
		UpdatedAt:    instance.UpdatedAt.String(),
	}, nil
}

// ListInstancesByWorkflowID 列出工作流的所有实例
func (qs *WorkflowInstanceQueryService) ListInstancesByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*WorkflowInstanceDTO, error) {
	instances, err := qs.repo.FindByWorkflowID(ctx, workflowID, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*WorkflowInstanceDTO, len(instances))
	for i, instance := range instances {
		completedAt := ""
		if instance.CompletedAt != nil {
			completedAt = instance.CompletedAt.String()
		}

		dtos[i] = &WorkflowInstanceDTO{
			ID:           instance.ID,
			WorkflowID:   instance.WorkflowID,
			Status:       string(instance.Status),
			Input:        string(instance.Input),
			Output:       string(instance.Output),
			ErrorMessage: instance.ErrorMessage,
			StartedAt:    instance.StartedAt.String(),
			CompletedAt:  completedAt,
			CreatedAt:    instance.CreatedAt.String(),
			UpdatedAt:    instance.UpdatedAt.String(),
		}
	}

	return dtos, nil
}

// ListAllInstances 列出所有实例（支持筛选）
func (qs *WorkflowInstanceQueryService) ListAllInstances(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*WorkflowInstanceDTO, int, error) {
        instances, total, err := qs.repo.FindAll(ctx, filters, limit, offset)
        if err != nil {
                return nil, 0, err
        }

        dtos := make([]*WorkflowInstanceDTO, len(instances))
        for i, instance := range instances {
                completedAt := ""
                if instance.CompletedAt != nil {
                        completedAt = instance.CompletedAt.String()
                }

                dtos[i] = &WorkflowInstanceDTO{
                        ID:           instance.ID,
                        WorkflowID:   instance.WorkflowID,
                        Status:       string(instance.Status),
                        Input:        string(instance.Input),
                        Output:       string(instance.Output),
                        ErrorMessage: instance.ErrorMessage,
                        StartedAt:    instance.StartedAt.String(),
                        CompletedAt:  completedAt,
                        CreatedAt:    instance.CreatedAt.String(),
                        UpdatedAt:    instance.UpdatedAt.String(),
                }
        }

        return dtos, total, nil
}
