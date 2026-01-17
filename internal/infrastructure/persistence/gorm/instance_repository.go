package persistence

import (
	"context"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"

	"gorm.io/gorm"
)

// workflowInstanceRepository 工作流实例仓储实现
type workflowInstanceRepository struct {
	GormRepository
}

// Save 保存工作流实例
func (r *workflowInstanceRepository) Save(ctx context.Context, instance *instance_aggregate.WorkflowInstance) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Create(instance).Error
	return err
}

// FindByID 根据ID查找实例
func (r *workflowInstanceRepository) FindByID(ctx context.Context, id string) (*instance_aggregate.WorkflowInstance, error) {
	var instance instance_aggregate.WorkflowInstance
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).First(&instance, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// FindByWorkflowID 根据工作流ID查找实例
func (r *workflowInstanceRepository) FindByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*instance_aggregate.WorkflowInstance, error) {
	var instances []*instance_aggregate.WorkflowInstance
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Limit(limit).
		Offset(offset).
		Find(&instances).Error
	return instances, err
}

// Update 更新实例
func (r *workflowInstanceRepository) Update(ctx context.Context, instance *instance_aggregate.WorkflowInstance) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Save(instance).Error
	return err
}

// Delete 删除实例（软删除）
func (r *workflowInstanceRepository) Delete(ctx context.Context, id string) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Delete(&instance_aggregate.WorkflowInstance{}, "id = ?", id).Error
	return err
}

// CountByWorkflowID 统计工作流的实例数量
func (r *workflowInstanceRepository) CountByWorkflowID(ctx context.Context, workflowID string) (int64, error) {
	var count int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return 0, err
	}
	err = db.WithContext(ctx).
		Model(&instance_aggregate.WorkflowInstance{}).
		Where("workflow_id = ?", workflowID).
		Count(&count).Error
	return count, err
}

// FindAll 查找所有实例（支持筛选）
func (r *workflowInstanceRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*instance_aggregate.WorkflowInstance, int, error) {
	var instances []*instance_aggregate.WorkflowInstance
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	query := db.WithContext(ctx).Model(&instance_aggregate.WorkflowInstance{})

	// 应用筛选条件
	if workflowID, ok := filters["workflow_id"].(string); ok && workflowID != "" {
		query = query.Where("workflow_id = ?", workflowID)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	err = query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&instances).Error

	return instances, int(total), err
}
