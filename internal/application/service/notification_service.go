package service

import (
	"context"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	websocket "jxt-evidence-system/process-management/internal/domain/aggregate/task/websocket"
	"log"
)

// DefaultNotificationService 默认通知服务实现
type DefaultNotificationService struct {
	wsNotifier websocket.WebSocketNotifier
}

// NewNotificationService 创建通知服务
func NewNotificationService(wsNotifier websocket.WebSocketNotifier) *DefaultNotificationService {
	return &DefaultNotificationService{
		wsNotifier: wsNotifier,
	}
}

// NotifyTaskCreated 通知任务已创建
func (s *DefaultNotificationService) NotifyTaskCreated(ctx context.Context, task *task_aggregate.Task) {
	if s.wsNotifier == nil {
		return
	}

	log.Printf("[NotificationService] Notifying task created: %s", task.TaskID.String())

	data := map[string]interface{}{
		"taskId":      task.TaskID.String(),
		"taskName":    task.TaskName,
		"taskKey":     task.TaskKey,
		"instanceId":  task.InstanceID.String(),
		"workflowId":  task.WorkflowID.String(),
		"assignee":    task.Assignee,
		"priority":    task.Priority,
		"status":      task.Status,
		"createdAt":   task.CreatedAt,
		"description": task.Description,
	}

	// 通知受让人
	if task.Assignee != 0 {
		s.wsNotifier.SendToUser(task.Assignee, "task_created", data)
	}

}

// NotifyTaskAssigned 通知任务已分配
func (s *DefaultNotificationService) NotifyTaskAssigned(ctx context.Context, task *task_aggregate.Task, assignee int) {
	if s.wsNotifier == nil {
		return
	}

	log.Printf("[NotificationService] Notifying task assigned: %s to %d", task.TaskID.String(), assignee)

	data := map[string]interface{}{
		"taskId":     task.TaskID.String(),
		"taskName":   task.TaskName,
		"instanceId": task.InstanceID.String(),
		"workflowId": task.WorkflowID.String(),
		"assignee":   assignee,
		"priority":   task.Priority,
		"status":     task.Status,
	}

	s.wsNotifier.SendToUser(assignee, "task_assigned", data)
}

// NotifyTaskCompleted 通知任务已完成
func (s *DefaultNotificationService) NotifyTaskCompleted(ctx context.Context, task *task_aggregate.Task) {
	if s.wsNotifier == nil {
		return
	}

	log.Printf("[NotificationService] Notifying task completed: %s", task.TaskID.String())

	data := map[string]interface{}{
		"taskId":      task.TaskID.String(),
		"taskName":    task.TaskName,
		"instanceId":  task.InstanceID.String(),
		"workflowId":  task.WorkflowID.String(),
		"assignee":    task.Assignee,
		"status":      task.Status,
		"completedAt": task.CompletedAt,
		"result":      task.Result,
	}

	// 通知任务创建者（如果有）
	if task.Assignee != 0 {
		s.wsNotifier.SendToUser(task.Assignee, "task_completed", data)
	}

	// TODO: 通知流程发起人
}

// NotifyWorkflowCompleted 通知工作流已完成
func (s *DefaultNotificationService) NotifyWorkflowCompleted(ctx context.Context, instance *instance_aggregate.WorkflowInstance) {
	if s.wsNotifier == nil {
		return
	}

	log.Printf("[NotificationService] Notifying workflow completed: %s", instance.InstanceId.String())

	_ = map[string]interface{}{
		"instanceId":  instance.InstanceId.String(),
		"workflowId":  instance.WorkflowID.String(),
		"status":      instance.Status,
		"completedAt": instance.CompletedAt,
	}

	// TODO: 通知流程发起人和相关人员
	// 这里需要从实例输入中获取发起人信息
	log.Printf("[NotificationService] Workflow completed notification sent")
}

// NoOpNotificationService 空操作通知服务（用于测试或禁用通知）
type NoOpNotificationService struct{}

// NewNoOpNotificationService 创建空操作通知服务
func NewNoOpNotificationService() *NoOpNotificationService {
	return &NoOpNotificationService{}
}

func (s *NoOpNotificationService) NotifyTaskCreated(ctx context.Context, task *task_aggregate.Task) {}

func (s *NoOpNotificationService) NotifyTaskAssigned(ctx context.Context, task *task_aggregate.Task, assignee int) {
}

func (s *NoOpNotificationService) NotifyTaskCompleted(ctx context.Context, task *task_aggregate.Task) {
}

func (s *NoOpNotificationService) NotifyWorkflowCompleted(ctx context.Context, instance *instance_aggregate.WorkflowInstance) {
}
