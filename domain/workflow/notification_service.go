package workflow

import (
	"context"
	"log"
)

// NotificationService 通知服务接口
type NotificationService interface {
	// NotifyTaskCreated 通知任务已创建
	NotifyTaskCreated(ctx context.Context, task *Task)

	// NotifyTaskAssigned 通知任务已分配
	NotifyTaskAssigned(ctx context.Context, task *Task, assignee string)

	// NotifyTaskCompleted 通知任务已完成
	NotifyTaskCompleted(ctx context.Context, task *Task)

	// NotifyWorkflowCompleted 通知工作流已完成
	NotifyWorkflowCompleted(ctx context.Context, instance *WorkflowInstance)
}

// WebSocketNotifier WebSocket通知器
type WebSocketNotifier interface {
	SendToUser(userID string, msgType string, data map[string]interface{})
	SendToUsers(userIDs []string, msgType string, data map[string]interface{})
}

// DefaultNotificationService 默认通知服务实现
type DefaultNotificationService struct {
	wsNotifier WebSocketNotifier
}

// NewNotificationService 创建通知服务
func NewNotificationService(wsNotifier WebSocketNotifier) NotificationService {
	return &DefaultNotificationService{
		wsNotifier: wsNotifier,
	}
}

// NotifyTaskCreated 通知任务已创建
func (s *DefaultNotificationService) NotifyTaskCreated(ctx context.Context, task *Task) {
	if s.wsNotifier == nil {
		return
	}

	log.Printf("[NotificationService] Notifying task created: %s", task.ID)

	data := map[string]interface{}{
		"task_id":     task.ID,
		"task_name":   task.TaskName,
		"task_key":    task.TaskKey,
		"instance_id": task.InstanceID,
		"workflow_id": task.WorkflowID,
		"assignee":    task.Assignee,
		"priority":    task.Priority,
		"status":      task.Status,
		"created_at":  task.CreatedAt,
		"description": task.Description,
	}

	// 通知受让人
	if task.Assignee != "" {
		s.wsNotifier.SendToUser(task.Assignee, "task_created", data)
	}

	// 通知候选用户
	if len(task.CandidateUsers) > 0 {
		s.wsNotifier.SendToUsers(task.CandidateUsers, "task_created", data)
	}

	// 通知候选组（这里简化处理，实际应该查询组成员）
	if len(task.CandidateGroups) > 0 {
		for _, group := range task.CandidateGroups {
			log.Printf("[NotificationService] Task created for group: %s", group)
			// TODO: 查询组成员并通知
		}
	}
}

// NotifyTaskAssigned 通知任务已分配
func (s *DefaultNotificationService) NotifyTaskAssigned(ctx context.Context, task *Task, assignee string) {
	if s.wsNotifier == nil {
		return
	}

	log.Printf("[NotificationService] Notifying task assigned: %s to %s", task.ID, assignee)

	data := map[string]interface{}{
		"task_id":     task.ID,
		"task_name":   task.TaskName,
		"instance_id": task.InstanceID,
		"workflow_id": task.WorkflowID,
		"assignee":    assignee,
		"priority":    task.Priority,
		"status":      task.Status,
	}

	s.wsNotifier.SendToUser(assignee, "task_assigned", data)
}

// NotifyTaskCompleted 通知任务已完成
func (s *DefaultNotificationService) NotifyTaskCompleted(ctx context.Context, task *Task) {
	if s.wsNotifier == nil {
		return
	}

	log.Printf("[NotificationService] Notifying task completed: %s", task.ID)

	data := map[string]interface{}{
		"task_id":      task.ID,
		"task_name":    task.TaskName,
		"instance_id":  task.InstanceID,
		"workflow_id":  task.WorkflowID,
		"assignee":     task.Assignee,
		"status":       task.Status,
		"completed_at": task.CompletedAt,
		"result":       task.Result,
	}

	// 通知任务创建者（如果有）
	if task.Assignee != "" {
		s.wsNotifier.SendToUser(task.Assignee, "task_completed", data)
	}

	// TODO: 通知流程发起人
}

// NotifyWorkflowCompleted 通知工作流已完成
func (s *DefaultNotificationService) NotifyWorkflowCompleted(ctx context.Context, instance *WorkflowInstance) {
	if s.wsNotifier == nil {
		return
	}

	log.Printf("[NotificationService] Notifying workflow completed: %s", instance.ID)

	_ = map[string]interface{}{
		"instance_id":  instance.ID,
		"workflow_id":  instance.WorkflowID,
		"status":       instance.Status,
		"completed_at": instance.CompletedAt,
	}

	// TODO: 通知流程发起人和相关人员
	// 这里需要从实例输入中获取发起人信息
	log.Printf("[NotificationService] Workflow completed notification sent")
}

// NoOpNotificationService 空操作通知服务（用于测试或禁用通知）
type NoOpNotificationService struct{}

// NewNoOpNotificationService 创建空操作通知服务
func NewNoOpNotificationService() NotificationService {
	return &NoOpNotificationService{}
}

func (s *NoOpNotificationService) NotifyTaskCreated(ctx context.Context, task *Task) {}

func (s *NoOpNotificationService) NotifyTaskAssigned(ctx context.Context, task *Task, assignee string) {
}

func (s *NoOpNotificationService) NotifyTaskCompleted(ctx context.Context, task *Task) {}

func (s *NoOpNotificationService) NotifyWorkflowCompleted(ctx context.Context, instance *WorkflowInstance) {
}
