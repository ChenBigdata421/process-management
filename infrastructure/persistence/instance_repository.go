package persistence

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
	"gorm.io/gorm"
)

// WorkflowInstanceRepositoryImpl 工作流实例仓储实现
type WorkflowInstanceRepositoryImpl struct {
	db *gorm.DB
}

// NewWorkflowInstanceRepository 创建工作流实例仓储
func NewWorkflowInstanceRepository(db *gorm.DB) workflow.WorkflowInstanceRepository {
	return &WorkflowInstanceRepositoryImpl{db: db}
}

// Save 保存工作流实例
func (r *WorkflowInstanceRepositoryImpl) Save(ctx context.Context, instance *workflow.WorkflowInstance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

// FindByID 根据ID查找实例
func (r *WorkflowInstanceRepositoryImpl) FindByID(ctx context.Context, id string) (*workflow.WorkflowInstance, error) {
	var instance workflow.WorkflowInstance
	err := r.db.WithContext(ctx).First(&instance, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// FindByWorkflowID 根据工作流ID查找实例
func (r *WorkflowInstanceRepositoryImpl) FindByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*workflow.WorkflowInstance, error) {
	var instances []*workflow.WorkflowInstance
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Limit(limit).
		Offset(offset).
		Find(&instances).Error
	return instances, err
}

// Update 更新实例
func (r *WorkflowInstanceRepositoryImpl) Update(ctx context.Context, instance *workflow.WorkflowInstance) error {
	return r.db.WithContext(ctx).Save(instance).Error
}

// Delete 删除实例（软删除）
func (r *WorkflowInstanceRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&workflow.WorkflowInstance{}, "id = ?", id).Error
}

// CountByWorkflowID 统计工作流的实例数量
func (r *WorkflowInstanceRepositoryImpl) CountByWorkflowID(ctx context.Context, workflowID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&workflow.WorkflowInstance{}).
		Where("workflow_id = ?", workflowID).
		Count(&count).Error
	return count, err
}


// FindAll 查找所有实例（支持筛选）
func (r *WorkflowInstanceRepositoryImpl) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*workflow.WorkflowInstance, int, error) {
        var instances []*workflow.WorkflowInstance
        var total int64

        query := r.db.WithContext(ctx).Model(&workflow.WorkflowInstance{})

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
        err := query.
                Order("created_at DESC").
                Limit(limit).
                Offset(offset).
                Find(&instances).Error

        return instances, int(total), err
}
