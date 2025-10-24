package persistence

import (
	"context"

	"github.com/jxt/process-management/domain/workflow"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// TaskRepositoryImpl 任务仓储实现
type TaskRepositoryImpl struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓储
func NewTaskRepository(db *gorm.DB) workflow.TaskRepository {
	return &TaskRepositoryImpl{db: db}
}

// Save 保存任务
func (r *TaskRepositoryImpl) Save(ctx context.Context, task *workflow.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// FindByID 根据ID查找任务
func (r *TaskRepositoryImpl) FindByID(ctx context.Context, id string) (*workflow.Task, error) {
	var task workflow.Task
	err := r.db.WithContext(ctx).First(&task, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// FindByInstanceID 根据实例ID查找任务
func (r *TaskRepositoryImpl) FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*workflow.Task, error) {
	var tasks []*workflow.Task
	err := r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// FindTodoByAssignee 查找指定用户的待办任务
func (r *TaskRepositoryImpl) FindTodoByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*workflow.Task, int64, error) {
	var tasks []*workflow.Task
	var total int64

	// 查询待办任务（已认领或待认领）
	query := r.db.WithContext(ctx).
		Where("(assignee = ? OR ? = ANY(candidate_users))", assignee, assignee).
		Where("status IN (?, ?)", workflow.TaskStatusPending, workflow.TaskStatusClaimed)

	// 获取总数
	if err := query.Model(&workflow.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.
		Limit(limit).
		Offset(offset).
		Order("priority DESC, created_at ASC").
		Find(&tasks).Error

	return tasks, total, err
}

// FindDoneByAssignee 查找指定用户的已办任务
func (r *TaskRepositoryImpl) FindDoneByAssignee(ctx context.Context, assignee string, limit, offset int) ([]*workflow.Task, int64, error) {
	var tasks []*workflow.Task
	var total int64

	// 查询已办任务（已完成或已驳回）
	query := r.db.WithContext(ctx).
		Where("assignee = ?", assignee).
		Where("status IN (?, ?)", workflow.TaskStatusCompleted, workflow.TaskStatusRejected)

	// 获取总数
	if err := query.Model(&workflow.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.
		Limit(limit).
		Offset(offset).
		Order("completed_at DESC").
		Find(&tasks).Error

	return tasks, total, err
}

// FindClaimable 查找可认领的任务
func (r *TaskRepositoryImpl) FindClaimable(ctx context.Context, userID string, userGroups []string, limit, offset int) ([]*workflow.Task, int64, error) {
	var tasks []*workflow.Task
	var total int64

	// 查询可认领的任务
	query := r.db.WithContext(ctx).
		Where("status = ?", workflow.TaskStatusPending).
		Where("(assignee = '' OR assignee IS NULL OR assignee = ?)"+
			" OR ? = ANY(candidate_users)"+
			" OR candidate_groups && ?",
			userID, userID, pq.Array(userGroups))

	// 获取总数
	if err := query.Model(&workflow.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.
		Limit(limit).
		Offset(offset).
		Order("priority DESC, due_date ASC, created_at ASC").
		Find(&tasks).Error

	return tasks, total, err
}

// FindAll 查找所有任务（支持多条件查询）
func (r *TaskRepositoryImpl) FindAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*workflow.Task, int64, error) {
	var tasks []*workflow.Task
	var total int64

	query := r.db.WithContext(ctx)

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
	if err := query.Model(&workflow.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tasks).Error

	return tasks, total, err
}

// Update 更新任务
func (r *TaskRepositoryImpl) Update(ctx context.Context, task *workflow.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Delete 删除任务（软删除）
func (r *TaskRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&workflow.Task{}, "id = ?", id).Error
}

// TaskHistoryRepositoryImpl 任务历史仓储实现
type TaskHistoryRepositoryImpl struct {
	db *gorm.DB
}

// NewTaskHistoryRepository 创建任务历史仓储
func NewTaskHistoryRepository(db *gorm.DB) workflow.TaskHistoryRepository {
	return &TaskHistoryRepositoryImpl{db: db}
}

// Save 保存任务历史
func (r *TaskHistoryRepositoryImpl) Save(ctx context.Context, history *workflow.TaskHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// FindByTaskID 根据任务ID查找历史
func (r *TaskHistoryRepositoryImpl) FindByTaskID(ctx context.Context, taskID string, limit, offset int) ([]*workflow.TaskHistory, error) {
	var histories []*workflow.TaskHistory
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}

// FindByInstanceID 根据实例ID查找历史
func (r *TaskHistoryRepositoryImpl) FindByInstanceID(ctx context.Context, instanceID string, limit, offset int) ([]*workflow.TaskHistory, error) {
	var histories []*workflow.TaskHistory
	err := r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&histories).Error
	return histories, err
}
