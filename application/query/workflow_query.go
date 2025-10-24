package query

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
)

// WorkflowDTO 工作流数据传输对象
type WorkflowDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Definition  string `json:"definition"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// WorkflowQueryService 工作流查询服务
type WorkflowQueryService struct {
	repo workflow.WorkflowRepository
}

// NewWorkflowQueryService 创建查询服务
func NewWorkflowQueryService(repo workflow.WorkflowRepository) *WorkflowQueryService {
	return &WorkflowQueryService{repo: repo}
}

// GetWorkflowByID 根据ID获取工作流
func (qs *WorkflowQueryService) GetWorkflowByID(ctx context.Context, id string) (*WorkflowDTO, error) {
	wf, err := qs.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if wf == nil {
		return nil, workflow.ErrWorkflowNotFound
	}

	return &WorkflowDTO{
		ID:          wf.ID,
		Name:        wf.Name,
		Description: wf.Description,
		Status:      string(wf.Status),
		Definition:  wf.Definition,
		CreatedAt:   wf.CreatedAt.String(),
		UpdatedAt:   wf.UpdatedAt.String(),
	}, nil
}

// ListWorkflows 列出所有工作流
func (qs *WorkflowQueryService) ListWorkflows(ctx context.Context, limit, offset int) ([]*WorkflowDTO, error) {
	workflows, err := qs.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*WorkflowDTO, len(workflows))
	for i, wf := range workflows {
		dtos[i] = &WorkflowDTO{
			ID:          wf.ID,
			Name:        wf.Name,
			Description: wf.Description,
			Status:      string(wf.Status),
			Definition:  wf.Definition,
			CreatedAt:   wf.CreatedAt.String(),
			UpdatedAt:   wf.UpdatedAt.String(),
		}
	}

	return dtos, nil
}

// CountWorkflows 统计工作流数量
func (qs *WorkflowQueryService) CountWorkflows(ctx context.Context) (int64, error) {
	return qs.repo.Count(ctx)
}

