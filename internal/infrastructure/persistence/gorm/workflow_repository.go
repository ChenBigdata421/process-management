package persistence

import (
	"context"

	workflow_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"

	"gorm.io/gorm"
)

// workflowRepository 工作流仓储实现
type workflowRepository struct {
	GormRepository
}

// Save 保存工作流
func (r *workflowRepository) Save(ctx context.Context, wf *workflow_aggregate.Workflow) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Create(wf).Error
	return err
}

// FindByID 根据ID查找工作流
func (r *workflowRepository) FindByID(ctx context.Context, id string) (*workflow_aggregate.Workflow, error) {
	var wf workflow_aggregate.Workflow
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).First(&wf, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &wf, nil
}

// FindAll 查找所有工作流
func (r *workflowRepository) FindAll(ctx context.Context, limit, offset int) ([]*workflow_aggregate.Workflow, error) {
	var workflows []*workflow_aggregate.Workflow
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&workflows).Error
	return workflows, err
}

// Update 更新工作流
func (r *workflowRepository) Update(ctx context.Context, wf *workflow_aggregate.Workflow) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Save(wf).Error
	return err
}

// Delete 删除工作流（软删除）
func (r *workflowRepository) Delete(ctx context.Context, id string) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Delete(&workflow_aggregate.Workflow{}, "id = ?", id).Error
	return err
}

// Count 统计工作流数量
func (r *workflowRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return 0, err
	}

	err = db.WithContext(ctx).Model(&workflow_aggregate.Workflow{}).Count(&count).Error
	return count, err
}
