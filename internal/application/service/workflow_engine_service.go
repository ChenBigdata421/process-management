package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"jxt-evidence-system/process-management/internal/application/command"
	"jxt-evidence-system/process-management/internal/application/service/port"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	instance_repository "jxt-evidence-system/process-management/internal/domain/aggregate/instance/repository"
	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	task_repository "jxt-evidence-system/process-management/internal/domain/aggregate/task/repository"
	workflow_repository "jxt-evidence-system/process-management/internal/domain/aggregate/workflow/repository"
	domain_service "jxt-evidence-system/process-management/internal/domain/service"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	errors_ "jxt-evidence-system/process-management/shared/common/errors"
	"jxt-evidence-system/process-management/shared/common/global"
	"jxt-evidence-system/process-management/shared/common/status"
	"log"
	"strconv"
	"strings"
	"time"
)

// InstanceCompletedCallback 工作流实例完成回调函数类型
type InstanceCompletedCallback func(ctx context.Context, instance *instance_aggregate.WorkflowInstance) error

// WorkflowEngineService 工作流引擎服务（应用层）
// 负责工作流执行的应用协调，依赖领域服务和仓储
type WorkflowEngineService struct {
	workflowRepo        workflow_repository.WorkflowRepository
	instanceRepo        instance_repository.WorkflowInstanceRepository
	taskRepo            task_repository.TaskRepository
	domainService       domain_service.WorkflowDomainService
	notificationSvc     port.NotificationService // 通知服务（可选）
	downloadApprovalSvc *DownloadApprovalService
}

// NewWorkflowEngineService 创建工作流引擎服务
func NewWorkflowEngineService(
	workflowRepo workflow_repository.WorkflowRepository,
	instanceRepo instance_repository.WorkflowInstanceRepository,
	taskRepo task_repository.TaskRepository,
	domainService domain_service.WorkflowDomainService,
) *WorkflowEngineService {
	return &WorkflowEngineService{
		workflowRepo:    workflowRepo,
		instanceRepo:    instanceRepo,
		taskRepo:        taskRepo,
		domainService:   domainService,
		notificationSvc: NewNoOpNotificationService(), // 默认使用空操作通知服务
	}
}

// NewWorkflowEngineServiceWithNotification 创建工作流引擎服务（带通知服务）
func NewWorkflowEngineServiceWithNotification(
	workflowRepo workflow_repository.WorkflowRepository,
	instanceRepo instance_repository.WorkflowInstanceRepository,
	taskRepo task_repository.TaskRepository,
	domainService domain_service.WorkflowDomainService,
	notificationSvc port.NotificationService,
	downloadApprovalSvc *DownloadApprovalService,
) *WorkflowEngineService {
	return &WorkflowEngineService{
		workflowRepo:        workflowRepo,
		instanceRepo:        instanceRepo,
		taskRepo:            taskRepo,
		domainService:       domainService,
		notificationSvc:     notificationSvc,
		downloadApprovalSvc: downloadApprovalSvc,
	}
}

// SetNotificationService 设置通知服务
func (s *WorkflowEngineService) SetNotificationService(svc port.NotificationService) {
	s.notificationSvc = svc
}

// StepDefinition 步骤定义（从领域服务导入）
type StepDefinition = domain_service.StepDefinition

// WorkflowDefinitionStruct 工作流定义（从领域服务导入）
type WorkflowDefinitionStruct = domain_service.WorkflowDefinitionStruct

// StartInstance 启动工作流实例并执行第一步
func (s *WorkflowEngineService) StartInstance(ctx context.Context, instanceID valueobject.InstanceID) error {
	log.Printf("[EngineService] Starting instance: %s", instanceID)

	// 获取实例
	instance, err := s.instanceRepo.FindByID(ctx, instanceID)
	if err != nil {
		if errors.Is(err, errors_.ErrInstanceNotFound) {
			return fmt.Errorf("instance not found: %s", instanceID)
		}
		return fmt.Errorf("failed to find instance: %w", err)
	}
	// 获取工作流定义
	wf, err := s.workflowRepo.FindByID(ctx, instance.WorkflowID)
	if err != nil {
		if errors.Is(err, errors_.ErrWorkflowNotFound) {
			return fmt.Errorf("workflow not found: %s", instance.WorkflowID.String())
		}
		return fmt.Errorf("failed to find workflow: %w", err)
	}

	// 解析工作流定义
	var definition WorkflowDefinitionStruct
	if err := json.Unmarshal([]byte(wf.Definition), &definition); err != nil {
		return fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	if len(definition.Steps) == 0 {
		return fmt.Errorf("workflow has no steps")
	}

	// 更新实例状态为运行中
	instance.Status = status.InstanceStatusRunning
	instance.StartedAt = time.Now()

	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	log.Printf("[EngineService] Instance started, executing first step: %s", definition.Steps[0].Name)

	// 执行第一步
	return s.executeStep(ctx, instance, &definition.Steps[0], &definition)
}

// executeStep 执行工作流步骤
func (s *WorkflowEngineService) executeStep(ctx context.Context, instance *instance_aggregate.WorkflowInstance, step *StepDefinition, definition *WorkflowDefinitionStruct) error {
	log.Printf("[EngineService] Executing step: %s (type: %s) for instance: %s", step.Name, step.Type, instance.InstanceId.String())

	switch step.Type {
	case "userTask":
		return s.executeUserTask(ctx, instance, step)
	case "process":
		return s.executeProcessTask(ctx, instance, step, definition)
	case "parallel":
		return s.executeParallelTasks(ctx, instance, step, definition)
	case "complete":
		return s.completeInstance(ctx, instance, step, definition)
	default:
		log.Printf("[EngineService] Unknown step type: %s, skipping", step.Type)
		// 未知类型，尝试执行下一步
		return s.executeNextStep(ctx, instance, step, definition)
	}
}

// executeUserTask 执行用户任务步骤
func (s *WorkflowEngineService) executeUserTask(ctx context.Context, instance *instance_aggregate.WorkflowInstance, step *StepDefinition) error {
	log.Printf("[EngineService] Creating user task for step: %s", step.Name)

	// 创建用户任务
	task := task_aggregate.NewTask(instance.InstanceId, instance.WorkflowID)
	if userID, ok := ctx.Value(global.UserIDKey).(int); ok {
		task.Assignee = userID
	}
	// 从步骤参数设置任务属性
	s.domainService.ApplyStepParamsToTask(task, step, instance)

	tasks, err := s.taskRepo.FindByInstanceID(ctx, instance.InstanceId)
	if err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	var taskHistories []command.TaskHistoryItem

	if len(tasks) > 0 {
		taskHistories = s.domainService.BuildTaskHistories(tasks)
	}

	// 尝试从上下文中获取 next_task_approver（来自前一个任务的输出）
	if nextApprover, ok := ctx.Value("next_task_approver").(int); ok {
		if nextApprover != 0 {
			task.Assignee = nextApprover
			log.Printf("[EngineService] Set task assignee from next_task_approver: %d", nextApprover)
		}
	}

	// 构建任务数据
	task.TaskData = s.domainService.BuildTaskData(instance, taskHistories, nil)

	// 保存任务
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	log.Printf("[EngineService] User task created: %s (ID: %s)", task.TaskName, task.TaskID.String())

	// 发送任务创建通知
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyTaskCreated(ctx, task)
	}

	log.Printf("[EngineService] Instance paused, waiting for task completion")

	return nil
}

// executeProcessTask 执行自动化处理任务
func (s *WorkflowEngineService) executeProcessTask(ctx context.Context, instance *instance_aggregate.WorkflowInstance, step *StepDefinition, definition *WorkflowDefinitionStruct) error {
	log.Printf("[EngineService] Executing automated process task: %s", step.Name)

	// 自动化任务直接执行完成，继续下一步
	// 这里可以根据需要添加实际的处理逻辑
	// 创建用户任务
	task := task_aggregate.NewTask(instance.InstanceId, instance.WorkflowID)

	// 从步骤参数设置任务属性
	s.domainService.ApplyStepParamsToTask(task, step, instance)
	task.Status = status.TaskStatusCompleted
	task.Result = status.TaskResultApproved
	// 保存任务
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	return s.executeNextStep(ctx, instance, step, definition)
}

// completeInstance 完成工作流实例
func (s *WorkflowEngineService) completeInstance(ctx context.Context, instance *instance_aggregate.WorkflowInstance, step *StepDefinition, definition *WorkflowDefinitionStruct) error {
	log.Printf("[EngineService] Completing instance: %s", instance.InstanceId.String())

	task := task_aggregate.NewTask(instance.InstanceId, instance.WorkflowID)
	// 从步骤参数设置任务属性
	s.domainService.ApplyStepParamsToTask(task, step, instance)
	task.Status = status.TaskStatusCompleted
	task.Result = status.TaskResultApproved
	// 保存任务
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}
	now := time.Now()
	instance.Status = status.InstanceStatusCompleted
	instance.CompletedAt = &now

	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	// 用于更新下载审批等业务状态
	if err := s.downloadApprovalSvc.ApproveDownload(ctx, instance.InstanceId, 30); err != nil {
		log.Printf("[EngineService] 更新下载审批业务状态失败！: %v", err)
	}

	log.Printf("[EngineService] Instance completed successfully")

	return nil
}

// ContinueAfterTask 任务完成后继续执行流程
func (s *WorkflowEngineService) ContinueAfterTask(ctx context.Context, task *task_aggregate.Task) error {
	log.Printf("[EngineService] Continuing workflow after task completion: %s", task.TaskID.String())

	// 获取实例
	instance, err := s.instanceRepo.FindByID(ctx, task.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to find instance: %w", err)
	}

	if instance == nil {
		return fmt.Errorf("instance not found: %s", task.InstanceID.String())
	}

	if instance.Status != status.InstanceStatusRunning {
		return fmt.Errorf("instance is not running: %s", instance.InstanceId.String())
	}

	// 获取工作流定义
	wf, err := s.workflowRepo.FindByID(ctx, instance.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to find workflow: %w", err)
	}

	if wf == nil {
		return fmt.Errorf("workflow not found: %s", instance.WorkflowID.String())
	}

	// 解析工作流定义
	var definition WorkflowDefinitionStruct
	if err := json.Unmarshal([]byte(wf.Definition), &definition); err != nil {
		return fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	// 检查当前任务是否是并行任务的子任务
	// 并行子任务的 taskKey 格式为：parallel_step_id_1, parallel_step_id_2
	var parentStepID string
	var isParallelSubTask bool
	if idx := strings.LastIndex(task.TaskKey, "_"); idx > 0 {
		potentialSuffix := task.TaskKey[idx+1:]
		if _, err := strconv.Atoi(potentialSuffix); err == nil {
			// 是并行子任务
			parentStepID = task.TaskKey[:idx]
			isParallelSubTask = true
			log.Printf("[EngineService] Detected parallel sub-task, parent step ID: %s", parentStepID)
		}
	}

	// 如果是并行子任务，检查该并行组的所有任务是否都已完成
	if isParallelSubTask {
		allCompleted, err := s.checkParallelTasksCompleted(ctx, instance, parentStepID)
		if err != nil {
			log.Printf("[EngineService] Failed to check parallel tasks: %v", err)
			return fmt.Errorf("failed to check parallel tasks: %w", err)
		}

		if !allCompleted {
			log.Printf("[EngineService] Not all parallel tasks completed yet for group %s, waiting", parentStepID)
			return nil // 还有其他并行任务未完成，暂不继续
		}

		log.Printf("[EngineService] All parallel tasks for group %s completed, continuing", parentStepID)
	}

	// 找到当前步骤定义（支持并行任务的父步骤查找）
	currentStep := s.domainService.FindStepDefinitionByTaskKey(task.TaskKey, &definition)
	if currentStep == nil {
		return fmt.Errorf("current step not found for taskKey: %s", task.TaskKey)
	}

	log.Printf("[EngineService] Found current step: %s (type: %s)", currentStep.Name, currentStep.Type)

	// 用于更新下载审批等业务状态
	if currentStep.IsRoot {
		if err := s.downloadApprovalSvc.PendingDownload(ctx, instance.InstanceId); err != nil {
			log.Printf("[EngineService] 更新下载审批业务状态失败！: %v", err)
		}

	}
	// 执行下一步
	return s.executeNextStep(ctx, instance, currentStep, &definition)
}

// cancelParallelGroupTasks 取消并行组中的其他任务
// 当并行任务中的一个被驳回时，取消同组中其他已完成或待处理的任务
func (s *WorkflowEngineService) cancelParallelGroupTasks(ctx context.Context, tasks []*task_aggregate.Task, rejectedTask *task_aggregate.Task) error {
	// 提取并行组前缀
	idx := strings.LastIndex(rejectedTask.TaskKey, "_")
	if idx <= 0 {
		return nil // 不是并行任务
	}

	potentialSuffix := rejectedTask.TaskKey[idx+1:]
	if _, err := strconv.Atoi(potentialSuffix); err != nil {
		return nil // 后缀不是数字，不是并行任务
	}

	parallelGroupPrefix := rejectedTask.TaskKey[:idx]
	log.Printf("[EngineService] Cancelling parallel group tasks: %s", parallelGroupPrefix)

	var cancelledCount int
	for _, t := range tasks {
		if t.TaskID == rejectedTask.TaskID {
			continue // 跳过被驳回的任务本身
		}

		// 检查是否是同一并行组的任务
		if strings.HasPrefix(t.TaskKey, parallelGroupPrefix+"_") {
			if t.Status == status.TaskStatusCompleted || t.Status == status.TaskStatusPending {
				t.Status = status.TaskStatusCancelled
				if err := s.taskRepo.Update(ctx, t); err != nil {
					log.Printf("[EngineService] Failed to cancel parallel task %s: %v", t.TaskKey, err)
				} else {
					cancelledCount++
					log.Printf("[EngineService] Cancelled parallel task: %s", t.TaskKey)
				}
			}
		}
	}

	if cancelledCount > 0 {
		log.Printf("[EngineService] Cancelled %d tasks in parallel group %s", cancelledCount, parallelGroupPrefix)
	}

	return nil
}

// createParallelTasks 创建并行任务
// 根据并行步骤定义和原任务信息创建新的并行任务
func (s *WorkflowEngineService) createParallelTasks(ctx context.Context, instance *instance_aggregate.WorkflowInstance, previousStep *StepDefinition, previousTaskInfo *domain_service.PreviousTaskInfo, taskHistories []command.TaskHistoryItem) error {
	log.Printf("[EngineService] Creating parallel tasks for step: %s", previousStep.Name)

	for i, parallelStepDef := range previousStep.ParallelTasks {
		// 查找原来的任务，获取原始的 assignee
		var originalAssignee int
		originalTaskKey := fmt.Sprintf("%s_%d", previousTaskInfo.StepID, i+1)
		for _, oldTask := range previousTaskInfo.ParallelTasks {
			if oldTask.TaskKey == originalTaskKey {
				originalAssignee = oldTask.Assignee
				break
			}
		}

		// 创建新的并行任务
		newTask := task_aggregate.NewTask(instance.InstanceId, instance.WorkflowID)

		// 创建并行任务的副本，修改其ID以包含父步骤前缀
		parallelStepCopy := parallelStepDef
		parallelStepCopy.ID = fmt.Sprintf("%s_%d", previousStep.ID, i+1)

		// 应用步骤参数
		s.domainService.ApplyStepParamsToTask(newTask, &parallelStepCopy, instance)

		// 优先使用原任务的处理人
		if originalAssignee != 0 {
			newTask.Assignee = originalAssignee
		}

		// 构建任务数据
		newTask.TaskData = s.domainService.BuildTaskData(instance, taskHistories, nil)

		// 保存任务
		if err := s.taskRepo.Save(ctx, newTask); err != nil {
			log.Printf("[EngineService] Failed to save parallel task %s: %v", parallelStepCopy.ID, err)
			continue
		}

		log.Printf("[EngineService] Created parallel task: %s (Assignee: %d)", newTask.TaskName, newTask.Assignee)

		// 发送通知
		if s.notificationSvc != nil && originalAssignee != 0 {
			s.notificationSvc.NotifyTaskAssigned(ctx, newTask, originalAssignee)
		}
	}

	return nil
}

// createSingleTask 创建单个任务
// 根据步骤定义和原任务信息创建新的单个任务
func (s *WorkflowEngineService) createSingleTask(ctx context.Context, instance *instance_aggregate.WorkflowInstance, previousStep *StepDefinition, previousTask *task_aggregate.Task, taskHistories []command.TaskHistoryItem) error {
	log.Printf("[EngineService] Creating single task for step: %s", previousStep.Name)

	// 创建新任务
	newTask := task_aggregate.NewTask(instance.InstanceId, instance.WorkflowID)
	newTask.TaskType = previousStep.Type
	newTask.Description = previousStep.Description

	s.domainService.ApplyStepParamsToTask(newTask, previousStep, instance)

	// 设置任务分配：优先使用上一个任务的处理人
	if previousTask.Assignee != 0 {
		newTask.Assignee = previousTask.Assignee
	}

	// 构建任务数据
	newTask.TaskData = s.domainService.BuildTaskData(instance, taskHistories, nil)

	// 保存新任务
	if err := s.taskRepo.Save(ctx, newTask); err != nil {
		return fmt.Errorf("failed to save new task: %w", err)
	}

	log.Printf("[EngineService] Created new task for previous step: %s", newTask.TaskID.String())

	// 发送通知
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyTaskAssigned(ctx, newTask, previousTask.Assignee)
	}

	return nil
}

// updateInstanceStatusToRunning 更新实例状态为运行中
func (s *WorkflowEngineService) updateInstanceStatusToRunning(ctx context.Context, instance *instance_aggregate.WorkflowInstance) error {
	instance.Status = status.InstanceStatusRunning
	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance status: %w", err)
	}
	return nil
}

// RejectAndGoBack 驳回任务并回退到上一个步骤
// 支持并行任务场景：如果上一步是并行任务组，会重新创建该组的所有并行任务
func (s *WorkflowEngineService) RejectAndGoBack(ctx context.Context, task *task_aggregate.Task) error {
	log.Printf("[EngineService] Rejecting task and going back: %s", task.TaskID.String())

	// 获取实例
	instance, err := s.instanceRepo.FindByID(ctx, task.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to find instance: %w", err)
	}
	if instance == nil {
		return fmt.Errorf("instance not found: %s", task.InstanceID.String())
	}

	if instance.Status != status.InstanceStatusRunning {
		return fmt.Errorf("instance is not running: %s", instance.InstanceId.String())
	}

	// 获取工作流定义
	wf, err := s.workflowRepo.FindByID(ctx, instance.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to find workflow: %w", err)
	}
	if wf == nil {
		return fmt.Errorf("workflow not found: %s", instance.WorkflowID.String())
	}

	// 解析工作流定义
	var definition WorkflowDefinitionStruct
	if err := json.Unmarshal([]byte(wf.Definition), &definition); err != nil {
		return fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	tasks, err := s.taskRepo.FindByInstanceID(ctx, instance.InstanceId)
	if err != nil {
		return fmt.Errorf("查找实例的任务失败: %w", err)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("没有发现实例: %s的任务", instance.InstanceId.String())
	}

	// 特殊场景：检查当前被驳回的任务是否属于并行任务组
	// 如果是并行任务中的一个被驳回，需要取消同组中其他已完成的任务
	if err := s.cancelParallelGroupTasks(ctx, tasks, task); err != nil {
		log.Printf("[EngineService] Error cancelling parallel group tasks: %v", err)
		// 不中断流程，继续处理驳回
	}

	// 查找上一个任务的完整信息（支持并行任务组）
	previousTaskInfo := s.domainService.FindPreviousTaskInfo(tasks, task.TaskID)
	if previousTaskInfo == nil {
		return fmt.Errorf("cannot reject first task, no previous completed task found")
	}

	// 构建任务历史
	taskHistories := s.domainService.BuildTaskHistories(tasks)

	// 根据上一步是否为并行任务组，采取不同的处理策略
	if previousTaskInfo.IsParallelGroup {
		// 场景1：驳回到并行任务组
		log.Printf("[EngineService] Previous step is parallel group: %s with %d tasks",
			previousTaskInfo.StepID, len(previousTaskInfo.ParallelTasks))

		// 查找并行步骤定义
		previousStep := s.domainService.FindStepDefinitionByTaskKey(previousTaskInfo.StepID, &definition)
		if previousStep == nil {
			return fmt.Errorf("parallel step definition not found for step ID: %s", previousTaskInfo.StepID)
		}

		if previousStep.Type != "parallel" {
			return fmt.Errorf("expected parallel step type, got: %s", previousStep.Type)
		}

		log.Printf("[EngineService] Found parallel step definition: %s with %d parallel tasks",
			previousStep.Name, len(previousStep.ParallelTasks))

		// 创建并行任务
		if err := s.createParallelTasks(ctx, instance, previousStep, previousTaskInfo, taskHistories); err != nil {
			return fmt.Errorf("failed to create parallel tasks: %w", err)
		}

		// 更新实例状态为运行中（等待并行任务完成）
		if err := s.updateInstanceStatusToRunning(ctx, instance); err != nil {
			return err
		}

	} else {
		// 场景2：驳回到单个任务
		previousTask := previousTaskInfo.SingleTask
		log.Printf("[EngineService] Previous step is single task: %s (TaskKey: %s, Assignee: %d)",
			previousTask.TaskName, previousTask.TaskKey, previousTask.Assignee)

		// 从工作流定义中查找对应的步骤定义
		previousStep := s.domainService.FindStepDefinitionByTaskKey(previousTask.TaskKey, &definition)
		if previousStep == nil {
			return fmt.Errorf("step definition not found for task key: %s", previousTask.TaskKey)
		}

		log.Printf("[EngineService] Found step definition: %s (%s)", previousStep.Name, previousStep.ID)

		// 创建单个任务
		if err := s.createSingleTask(ctx, instance, previousStep, previousTask, taskHistories); err != nil {
			return err
		}

		// 用于更新下载审批等业务状态
		if previousStep.IsRoot {
			if err := s.downloadApprovalSvc.RejectDownload(ctx, instance.InstanceId, task.Comment); err != nil {
				log.Printf("[EngineService] 更新下载审批业务状态失败！: %v", err)
			}

		}

		// 更新实例状态为运行中（等待新任务完成）
		if err := s.updateInstanceStatusToRunning(ctx, instance); err != nil {
			return err
		}
	}

	log.Printf("[EngineService] Task rejection and rollback completed successfully")
	return nil
}

// executeNextStep 执行下一个步骤
func (s *WorkflowEngineService) executeNextStep(ctx context.Context, instance *instance_aggregate.WorkflowInstance, currentStep *StepDefinition, definition *WorkflowDefinitionStruct) error {
	// 找到下一个步骤
	nextStep := s.domainService.FindNextStep(currentStep, definition, instance)

	if nextStep == nil {
		// 没有下一步，完成流程
		log.Printf("[EngineService] No next step found, completing instance")
		return s.completeInstance(ctx, instance, currentStep, definition)
	}

	log.Printf("[EngineService] Found next step: %s", nextStep.Name)

	// 执行下一步
	return s.executeStep(ctx, instance, nextStep, definition)
}

// executeParallelTasks 执行并行任务
func (s *WorkflowEngineService) executeParallelTasks(ctx context.Context, instance *instance_aggregate.WorkflowInstance, step *StepDefinition, definition *WorkflowDefinitionStruct) error {
	log.Printf("[EngineService] Executing parallel tasks for step: %s", step.Name)

	if len(step.ParallelTasks) == 0 {
		log.Printf("[EngineService] No parallel tasks defined, continuing to next step")
		return s.executeNextStep(ctx, instance, step, definition)
	}

	// 创建所有并行任务
	var createdTasks []string
	for i, parallelStep := range step.ParallelTasks {
		// 检查条件
		if parallelStep.Condition != "" {
			if !s.domainService.EvaluateCondition(parallelStep.Condition, instance) {
				log.Printf("[EngineService] Parallel task condition not met: %s", parallelStep.Condition)
				continue
			}
		}

		// 创建并行任务的副本，修改其ID以包含父步骤前缀
		// 例如：parallel_approval -> parallel_approval_1, parallel_approval_2
		parallelStepCopy := parallelStep
		originalID := parallelStepCopy.ID
		parallelStepCopy.ID = fmt.Sprintf("%s_%d", step.ID, i+1)

		log.Printf("[EngineService] Creating parallel task: %s (original: %s)", parallelStepCopy.ID, originalID)

		// 执行并行步骤（通常是 user_task）
		if parallelStepCopy.Type == "userTask" {
			if err := s.executeUserTask(ctx, instance, &parallelStepCopy); err != nil {
				log.Printf("[EngineService] Failed to create parallel task %s: %v", parallelStepCopy.Name, err)
				continue
			}
			createdTasks = append(createdTasks, parallelStepCopy.ID)
		}
		if parallelStepCopy.Type == "process" {
			if err := s.executeProcessTask(ctx, instance, &parallelStepCopy, definition); err != nil {
				log.Printf("[EngineService] Failed to execute parallel task %s: %v", parallelStepCopy.Name, err)
				continue
			}
			createdTasks = append(createdTasks, parallelStepCopy.ID)
		}
	}

	log.Printf("[EngineService] Created %d parallel tasks", len(createdTasks))

	// 并行任务创建后，流程暂停，等待所有任务完成
	// 注意：需要在 ContinueAfterTask 中检查是否所有并行任务都已完成
	return nil
}

// checkParallelTasksCompleted 检查并行任务是否全部完成
func (s *WorkflowEngineService) checkParallelTasksCompleted(ctx context.Context, instance *instance_aggregate.WorkflowInstance, stepID string) (bool, error) {

	tasks, err := s.taskRepo.FindByInstanceID(ctx, instance.InstanceId)
	if err != nil {
		return false, fmt.Errorf("查找实例的任务失败: %w", err)
	}

	// 查找属于该并行步骤的所有任务
	var parallelTasks []*task_aggregate.Task
	for _, task := range tasks {
		// 检查任务的 taskKey 是否以 stepID 开头（表示是该并行步骤的子任务）
		if strings.HasPrefix(task.TaskKey, stepID+"_") {
			parallelTasks = append(parallelTasks, task)
		}
	}

	if len(parallelTasks) == 0 {
		return true, nil // 没有并行任务，认为已完成
	}

	// 检查是否所有并行任务都已完成
	for _, task := range parallelTasks {
		if task.Status == status.TaskStatusPending {
			log.Printf("[EngineService] Parallel task %s not completed yet (status: %s)", task.TaskID.String(), task.Status)
			return false, nil
		}
	}

	log.Printf("[EngineService] All %d parallel tasks completed", len(parallelTasks))
	return true, nil
}
