package persistence

import (
	"context"
	"errors"

	"jxt-evidence-system/process-management/internal/application/command"
	workflow_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/workflow"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	errors_ "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/global"
	cQuery "jxt-evidence-system/process-management/shared/common/query"

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
func (r *workflowRepository) FindByID(ctx context.Context, id valueobject.WorkflowID) (*workflow_aggregate.Workflow, error) {
	var wf workflow_aggregate.Workflow
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).Where("id = ?", id).First(&wf).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors_.ErrWorkflowNotFound
		}
		return nil, err
	}
	return &wf, nil
}

// FindByName 根据Name查找工作流
func (r *workflowRepository) FindByName(ctx context.Context, name string) (*workflow_aggregate.Workflow, error) {
	var wf workflow_aggregate.Workflow
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).Where("name = ?", name).First(&wf).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors_.ErrWorkflowNotFound
		}
		return nil, err
	}
	return &wf, nil
}

// GetPage 查找所有工作流（支持筛选）
func (r *workflowRepository) GetPage(ctx context.Context, query *command.WorkflowPagedQuery) ([]*workflow_aggregate.Workflow, int, error) {
	var workflows []*workflow_aggregate.Workflow
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = db.WithContext(ctx).Model(&workflow_aggregate.Workflow{}).
		Scopes(
			cQuery.MakeCondition(query.GetNeedSearch(), global.ProcessDriver), // 使用通用查询条件
			cQuery.Paginate(query.GetPageSize(), query.GetPageIndex()),        // 分页
		).
		Find(&workflows).Limit(-1).Offset(-1).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return workflows, int(total), err
}

func (r *workflowRepository) GetAllWorkflow(ctx context.Context) ([]*workflow_aggregate.Workflow, error) {
	var workflows []*workflow_aggregate.Workflow
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}

	err = db.WithContext(ctx).Model(&workflow_aggregate.Workflow{}).
		Find(&workflows).Error
	if err != nil {
		return nil, err
	}

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
func (r *workflowRepository) Delete(ctx context.Context, id valueobject.WorkflowID) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Where("id = ?", id).Delete(&workflow_aggregate.Workflow{}).Error
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
