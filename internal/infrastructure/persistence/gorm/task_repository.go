package persistence

import (
	"context"
	"errors"

	"jxt-evidence-system/process-management/internal/application/command"
	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	errors_ "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/global"
	cQuery "jxt-evidence-system/process-management/shared/common/query"
	"jxt-evidence-system/process-management/shared/common/status"

	"gorm.io/gorm"
)

// taskRepository 任务仓储实现
type taskRepository struct {
	GormRepository
}

// Save 保存任务
func (r *taskRepository) Save(ctx context.Context, task *task_aggregate.Task) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Omit("Workflow", "Instance").Create(task).Error
	return err
}

// FindByID 根据ID查找任务
func (r *taskRepository) FindByID(ctx context.Context, id valueobject.TaskID) (*task_aggregate.Task, error) {
	var task task_aggregate.Task
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).Debug().
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name")
		}).
		Preload("Instance", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "instance_no")
		}).
		Where("id = ?", id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors_.ErrTaskNotFound
		}
		return nil, err
	}
	task.WorklowName = task.Workflow.Name
	task.InstaceNo = task.Instance.InstanceNo
	task.WorkflowNo = task.Workflow.WorkflowNo
	return &task, nil
}

// FindByInstanceID 根据实例ID查找任务
func (r *taskRepository) FindByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) ([]*task_aggregate.Task, error) {
	var tasks []*task_aggregate.Task
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).Debug().
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name")
		}).
		Preload("Instance", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "instance_no")
		}).
		Where("instance_id = ?", instanceID).
		Find(&tasks).Error
	for i := range tasks {
		tasks[i].WorklowName = tasks[i].Workflow.Name
		tasks[i].InstaceNo = tasks[i].Instance.InstanceNo
		tasks[i].WorkflowNo = tasks[i].Workflow.WorkflowNo
	}
	return tasks, err
}

func (r *taskRepository) CountByInstanceID(ctx context.Context, id valueobject.InstanceID) (int, error) {
	var count int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return 0, err
	}
	err = db.WithContext(ctx).Debug().Model(task_aggregate.Task{}).Count(&count).Error
	return int(count), err
}

// FindRecentByInstanceID 根据实例ID查找最近的一条任务
func (r *taskRepository) FindRecentByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (*task_aggregate.Task, error) {
	var task task_aggregate.Task
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name")
		}).
		Preload("Instance", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "instance_no")
		}).
		Where("instance_id = ?", instanceID).
		Order("created_at DESC").
		First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors_.ErrTaskNotFound
		}
		return nil, err
	}
	task.WorklowName = task.Workflow.Name
	task.InstaceNo = task.Instance.InstanceNo
	task.WorkflowNo = task.Workflow.WorkflowNo
	return &task, nil
}

// CountTasksByInstanceID 统计实例的任务数量
func (r *taskRepository) CountTasksByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) (int, error) {
	var count int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return 0, err
	}

	err = db.WithContext(ctx).
		Model(&task_aggregate.Task{}).
		Where("instance_id = ?", instanceID).
		Count(&count).Error

	return int(count), err
}

// FindTodoByAssignee 查找指定用户的待办任务
func (r *taskRepository) FindTodoByAssignee(ctx context.Context, assignee int, query *command.TodoTaskPagedQuery) ([]*task_aggregate.Task, int, error) {
	var tasks []*task_aggregate.Task
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 基础查询条件：待办任务（已认领或待认领）
	baseQuery := db.WithContext(ctx).Debug().
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name")
		}).
		Preload("Instance", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "instance_no")
		}).
		Joins("LEFT JOIN workflows ON workflow_tasks.workflow_id = workflows.id").
		Select("workflow_tasks.*", "workflows.name as worflow_name").
		Where("(workflow_tasks.assignee = ?)", assignee).
		Where("workflow_tasks.status IN (?)", status.TaskStatusPending)

	// 按任务名称查询
	if query.TaskName != "" {
		baseQuery = baseQuery.Where("workflow_tasks.task_name LIKE ?", "%"+query.TaskName+"%")
	}

	// 按工作流名称查询
	if query.WorkflowName != "" {
		baseQuery = baseQuery.Where("workflows.name LIKE ?", "%"+query.WorkflowName+"%")
	}

	// 获取总数
	if err := baseQuery.Model(&task_aggregate.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := baseQuery.
		Limit(query.GetPageSize()).
		Offset((query.GetPageIndex() - 1) * query.GetPageSize()).
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	for i := range tasks {
		tasks[i].WorklowName = tasks[i].Workflow.Name
		tasks[i].InstaceNo = tasks[i].Instance.InstanceNo
		tasks[i].WorkflowNo = tasks[i].Workflow.WorkflowNo
	}
	return tasks, int(total), nil
}

// FindDoneByAssignee 查找指定用户的已办任务
func (r *taskRepository) FindDoneByAssignee(ctx context.Context, assignee int, query *command.DoneTaskPagedQuery) ([]*task_aggregate.Task, int, error) {
	var tasks []*task_aggregate.Task
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 基础查询条件：已办任务（已完成或已驳回）
	baseQuery := db.WithContext(ctx).Debug().
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name")
		}).
		Preload("Instance", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "instance_no")
		}).
		Joins("LEFT JOIN workflows ON workflow_tasks.workflow_id = workflows.id").
		Select("workflow_tasks.*", "workflows.name as worflow_name").
		Where("workflow_tasks.assignee = ?", assignee).
		Where("workflow_tasks.status IN (?, ?)", status.TaskStatusCompleted, status.TaskStatusRejected)

	// 按任务名称查询
	if query.TaskName != "" {
		baseQuery = baseQuery.Where("workflow_tasks.task_name LIKE ?", "%"+query.TaskName+"%")
	}

	// 按工作流名称查询
	if query.WorkflowName != "" {
		baseQuery = baseQuery.Where("workflows.name LIKE ?", "%"+query.WorkflowName+"%")
	}

	// 获取总数
	if err := baseQuery.Model(&task_aggregate.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	if err := baseQuery.
		Scopes(
			cQuery.Paginate(query.GetPageSize(), query.GetPageIndex()), // 分页
		).
		Order("completed_at DESC").
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	// 填充工作流名称
	for i := range tasks {
		tasks[i].WorklowName = tasks[i].Workflow.Name
		tasks[i].InstaceNo = tasks[i].Instance.InstanceNo
		tasks[i].WorkflowNo = tasks[i].Workflow.WorkflowNo
	}

	return tasks, int(total), nil
}

// GetPage 查找所有任务（支持筛选）
func (r *taskRepository) GetPage(ctx context.Context, query *command.TaskPagedQuery) ([]*task_aggregate.Task, int, error) {
	var tasks []*task_aggregate.Task
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	err = db.WithContext(ctx).Model(&task_aggregate.Task{}).
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "workflow_no", "name")
		}).
		Preload("Instance", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "instance_no")
		}).
		Scopes(
			cQuery.MakeCondition(query.GetNeedSearch(), global.ProcessDriver), // 使用通用查询条件
			cQuery.Paginate(query.GetPageSize(), query.GetPageIndex()),        // 分页
		).
		Find(&tasks).Limit(-1).Offset(-1).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	for i := range tasks {
		tasks[i].WorklowName = tasks[i].Workflow.Name
		tasks[i].InstaceNo = tasks[i].Instance.InstanceNo
		tasks[i].WorkflowNo = tasks[i].Workflow.WorkflowNo
	}

	return tasks, int(total), err
}

// Update 更新任务
func (r *taskRepository) Update(ctx context.Context, task *task_aggregate.Task) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	// 使用 Omit 排除关联对象，避免自动保存 Workflow 导致 JSON 格式错误
	err = db.WithContext(ctx).Omit("Workflow", "Instance").Save(task).Error
	return err
}

// Delete 删除任务（软删除）
func (r *taskRepository) Delete(ctx context.Context, id valueobject.TaskID) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Where("id = ?", id).Delete(&task_aggregate.Task{}).Error
	return err
}

// taskHistoryRepository 任务历史仓储实现
type taskHistoryRepository struct {
	GormRepository
}

// Save 保存任务历史
func (r *taskHistoryRepository) Save(ctx context.Context, history *task_aggregate.TaskHistory) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Create(history).Error
	return err
}

// FindByTaskID 根据任务ID查找历史
func (r *taskHistoryRepository) FindByTaskID(ctx context.Context, id valueobject.TaskID) ([]*task_aggregate.TaskHistory, error) {
	var histories []*task_aggregate.TaskHistory
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Where("task_id = ?", id).
		Find(&histories).Error
	return histories, err
}

// FindByInstanceID 根据实例ID查找历史
func (r *taskHistoryRepository) FindByInstanceID(ctx context.Context, instanceID valueobject.InstanceID) ([]*task_aggregate.TaskHistory, error) {
	var histories []*task_aggregate.TaskHistory
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Find(&histories).Error

	return histories, err
}
