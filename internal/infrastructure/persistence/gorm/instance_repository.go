package persistence

import (
	"context"
	"errors"
	"jxt-evidence-system/process-management/internal/application/command"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	errors_ "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/global"
	cQuery "jxt-evidence-system/process-management/shared/common/query"

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
	err = db.WithContext(ctx).Omit("Workflow").Create(instance).Error
	return err
}

// FindByID 根据ID查找实例
func (r *workflowInstanceRepository) FindByID(ctx context.Context, id valueobject.InstanceID) (*instance_aggregate.WorkflowInstance, error) {
	var instance instance_aggregate.WorkflowInstance
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name")
		}).
		Where("id = ?", id).First(&instance).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors_.ErrInstanceNotFound
		}
		return nil, err
	}
	instance.WorkflowNo = instance.Workflow.WorkflowNo
	instance.WorkflowName = instance.Workflow.Name
	return &instance, nil
}

// FindByWorkflowID 根据工作流ID查找实例
func (r *workflowInstanceRepository) FindByWorkflowID(ctx context.Context, query *command.GetInstancesByWorkflowPagedQuery) ([]*instance_aggregate.WorkflowInstance, int, error) {
	var instances []*instance_aggregate.WorkflowInstance
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}
	err = db.WithContext(ctx).
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name") //必须有关联字段id，否则出现Panic错误
		}).
		Scopes(
			cQuery.MakeCondition(query.GetNeedSearch(), global.ProcessDriver), // 使用通用查询条件
			cQuery.Paginate(query.GetPageSize(), query.GetPageIndex()),        // 分页
		).
		Find(&instances).Limit(-1).Offset(-1).
		Count(&total).Error
	for i := range instances {
		instances[i].WorkflowNo = instances[i].Workflow.WorkflowNo
		instances[i].WorkflowName = instances[i].Workflow.Name
	}
	return instances, int(total), err
}

// Update 更新实例
func (r *workflowInstanceRepository) Update(ctx context.Context, instance *instance_aggregate.WorkflowInstance) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Omit("Workflow").Save(instance).Error
	return err
}

// Delete 删除实例（软删除）
func (r *workflowInstanceRepository) Delete(ctx context.Context, id valueobject.InstanceID) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Where("id = ?", id).Delete(&instance_aggregate.WorkflowInstance{}).Error
	return err
}

// CountByWorkflowID 统计工作流的实例数量
func (r *workflowInstanceRepository) CountByWorkflowID(ctx context.Context, workflowID valueobject.WorkflowID) (int64, error) {
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

// GetPage 查找所有实例（支持筛选）
func (r *workflowInstanceRepository) GetPage(ctx context.Context, query *command.InstancePagedQuery) ([]*instance_aggregate.WorkflowInstance, int, error) {
	var instances []*instance_aggregate.WorkflowInstance
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = db.WithContext(ctx).Model(&instance_aggregate.WorkflowInstance{}).
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name")
		}).
		Scopes(
			cQuery.MakeCondition(query.GetNeedSearch(), global.ProcessDriver), // 使用通用查询条件
			cQuery.Paginate(query.GetPageSize(), query.GetPageIndex()),        // 分页
		).
		Find(&instances).Limit(-1).Offset(-1).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	for i := range instances {
		instances[i].WorkflowNo = instances[i].Workflow.WorkflowNo
		instances[i].WorkflowName = instances[i].Workflow.Name
	}

	return instances, int(total), err
}
