package persistence

import (
	"context"

	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	"jxt-evidence-system/process-management/shared/common/status"

	"github.com/lib/pq"
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
	err = db.WithContext(ctx).Create(task).Error
	return err
}

// FindByID 根据ID查找任务
func (r *taskRepository) FindByID(ctx context.Context, id string) (*task_aggregate.Task, error) {
	var task task_aggregate.Task
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).First(&task, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// FindByInstanceID 根据实例ID查找任务
func (r *taskRepository) FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*task_aggregate.Task, error) {
	var tasks []*task_aggregate.Task
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Limit(limit).
		Offset(offset).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

// FindTodoByAssignee 查找指定用户的待办任务
func (r *taskRepository) FindTodoByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*task_aggregate.Task, int64, error) {
	var tasks []*task_aggregate.Task
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}
	// 查询待办任务（已认领或待认领）
	query := db.WithContext(ctx).
		Where("(assignee = ? OR ? = ANY(candidate_users))", assignee, assignee).
		Where("status IN (?, ?)", status.TaskStatusPending, status.TaskStatusClaimed)

	// 获取总数
	if err := query.Model(&task_aggregate.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.
		Limit(limit).
		Offset(offset).
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error

	return tasks, total, err
}

// FindDoneByAssignee 查找指定用户的已办任务
func (r *taskRepository) FindDoneByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*task_aggregate.Task, int64, error) {
	var tasks []*task_aggregate.Task
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 查询已办任务（已完成或已驳回）
	query := db.WithContext(ctx).
		Where("assignee = ?", assignee).
		Where("status IN (?, ?)", status.TaskStatusCompleted, status.TaskStatusRejected)

	// 获取总数
	if err := query.Model(&task_aggregate.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.
		Limit(limit).
		Offset(offset).
		Order("completed_at DESC").
		Find(&tasks).Error

	return tasks, total, err
}

// FindClaimable 查找可认领的任务
func (r *taskRepository) FindClaimable(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*task_aggregate.Task, int64, error) {
	var tasks []*task_aggregate.Task
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 查询可认领的任务
	query := db.WithContext(ctx).
		Where("status = ?", status.TaskStatusPending).
		Where("(assignee = '' OR assignee IS NULL OR assignee = ?)"+
			" OR ? = ANY(candidate_users)"+
			" OR candidate_groups && ?",
			userID, userID, pq.Array(userGroups))

	// 获取总数
	if err := query.Model(&task_aggregate.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.
		Limit(limit).
		Offset(offset).
		Order("priority DESC, due_date ASC, created_at ASC").
		Find(&tasks).Error

	return tasks, total, err
}

// FindAll 查找所有任务（支持多条件查询）
func (r *taskRepository) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*task_aggregate.Task, int64, error) {
	var tasks []*task_aggregate.Task
	var total int64
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, 0, err
	}

	query := db.WithContext(ctx)

	// 按任务名称查询
	if taskName, ok := filters["task_name"].(string); ok && taskName != "" {
		query = query.Where("task_name LIKE ?", "%"+taskName+"%")
	}

	// 按流程名称查询（需要JOIN workflow表）
	if workflowName, ok := filters["workflow_name"].(string); ok && workflowName != "" {
		query = query.Where("workflow_id IN (SELECT id FROM workflows WHERE name LIKE ?)", "%"+workflowName+"%")
	}

	// 按任务状态查询
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// 按处理人查询
	if assignee, ok := filters["assignee"].(string); ok && assignee != "" {
		query = query.Where("assignee = ?", assignee)
	}

	// 获取总数
	if err := query.Model(&task_aggregate.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tasks).Error

	return tasks, total, err
}

// Update 更新任务
func (r *taskRepository) Update(ctx context.Context, task *task_aggregate.Task) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Save(task).Error
	return err
}

// Delete 删除任务（软删除）
func (r *taskRepository) Delete(ctx context.Context, id string) error {
	db, err := r.GetOrm(ctx)
	if err != nil {
		return err
	}
	err = db.WithContext(ctx).Delete(&task_aggregate.Task{}, "id = ?", id).Error
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
func (r *taskHistoryRepository) FindByTaskID(ctx context.Context, taskID string, limit, offset int) ([]*task_aggregate.TaskHistory, error) {
	var histories []*task_aggregate.TaskHistory
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}

// FindByInstanceID 根据实例ID查找历史
func (r *taskHistoryRepository) FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*task_aggregate.TaskHistory, error) {
	var histories []*task_aggregate.TaskHistory
	db, err := r.GetOrm(ctx)
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}
