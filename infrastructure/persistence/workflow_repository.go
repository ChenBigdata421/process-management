package persistence

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
	"gorm.io/gorm"
)

// WorkflowRepositoryImpl 工作流仓储实现
type WorkflowRepositoryImpl struct {
	db *gorm.DB
}

// NewWorkflowRepository 创建工作流仓储
func NewWorkflowRepository(db *gorm.DB) workflow.WorkflowRepository {
	return &WorkflowRepositoryImpl{db: db}
}

// Save 保存工作流
func (r *WorkflowRepositoryImpl) Save(ctx context.Context, wf *workflow.Workflow) error {
	return r.db.WithContext(ctx).Create(wf).Error
}

// FindByID 根据ID查找工作流
func (r *WorkflowRepositoryImpl) FindByID(ctx context.Context, id string) (*workflow.Workflow, error) {
	var wf workflow.Workflow
	err := r.db.WithContext(ctx).First(&wf, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &wf, nil
}

// FindAll 查找所有工作流
func (r *WorkflowRepositoryImpl) FindAll(ctx context.Context, limit, offset int) ([]*workflow.Workflow, error) {
	var workflows []*workflow.Workflow
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&workflows).Error
	return workflows, err
}

// Update 更新工作流
func (r *WorkflowRepositoryImpl) Update(ctx context.Context, wf *workflow.Workflow) error {
	return r.db.WithContext(ctx).Save(wf).Error
}

// Delete 删除工作流（软删除）
func (r *WorkflowRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&workflow.Workflow{}, "id = ?", id).Error
}

// Count 统计工作流数量
func (r *WorkflowRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&workflow.Workflow{}).Count(&count).Error
	return count, err
}

