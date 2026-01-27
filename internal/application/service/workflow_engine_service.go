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
	"strings"
	"time"
)

// WorkflowEngineService 工作流引擎服务（应用层）
// 负责工作流执行的应用协调，依赖领域服务和仓储
type WorkflowEngineService struct {
	workflowRepo    workflow_repository.WorkflowRepository
	instanceRepo    instance_repository.WorkflowInstanceRepository
	taskRepo        task_repository.TaskRepository
	domainService   domain_service.WorkflowDomainService
	notificationSvc port.NotificationService // 通知服务（可选）
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
) *WorkflowEngineService {
	return &WorkflowEngineService{
		workflowRepo:    workflowRepo,
		instanceRepo:    instanceRepo,
		taskRepo:        taskRepo,
		domainService:   domainService,
		notificationSvc: notificationSvc,
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

	// 找到当前步骤
	var currentStep *StepDefinition
	for i := range definition.Steps {
		if definition.Steps[i].ID == task.TaskKey {
			currentStep = &definition.Steps[i]
			break
		}
	}

	if currentStep == nil {
		return fmt.Errorf("current step not found: %s", task.TaskKey)
	}

	// 检查是否是并行任务的一部分
	// 注意：只有当步骤类型是 "parallel" 时才需要检查并行任务
	// 不能简单地根据 taskKey 包含 "_" 来判断，因为步骤ID本身就可能包含 "_"
	if currentStep != nil && currentStep.Type == "parallel" {
		// 检查该并行步骤的所有任务是否都已完成
		allCompleted, err := s.checkParallelTasksCompleted(ctx, instance, currentStep.ID)
		if err != nil {
			log.Printf("[EngineService] Failed to check parallel tasks: %v", err)
		} else if !allCompleted {
			log.Printf("[EngineService] Not all parallel tasks completed yet, waiting")
			return nil // 还有其他并行任务未完成，暂不继续
		}
		log.Printf("[EngineService] All parallel tasks for step %s completed", currentStep.ID)
	}

	// 更新实例状态为运行中
	instance.Status = status.InstanceStatusRunning

	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	log.Printf("[EngineService] Instance resumed, finding next step")

	// 执行下一步
	return s.executeNextStep(ctx, instance, currentStep, &definition)
}

// RejectAndGoBack 驳回任务并回退到上一个步骤
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
		return fmt.Errorf("查找实例的任务失败")
	}

	if len(tasks) == 0 {
		return fmt.Errorf("没有发现实例: %s的任务", instance.InstanceId.String())
	}

	// 找出当前任务之前最后完成的任务
	previousTask := s.domainService.FindPreviousCompletedTask(tasks, task.TaskID)
	if previousTask == nil {
		return fmt.Errorf("cannot reject first task, no previous completed task found")
	}

	log.Printf("[EngineService] Found previous task: %s (TaskKey: %s, Assignee: %s)", previousTask.TaskName, previousTask.TaskKey, previousTask.Assignee)

	// 从工作流定义中查找对应的步骤定义
	previousStep := s.domainService.FindStepByID(previousTask.TaskKey, &definition)
	if previousStep == nil {
		return fmt.Errorf("step definition not found for task key: %s", previousTask.TaskKey)
	}

	log.Printf("[EngineService] Found step definition: %s (%s)", previousStep.Name, previousStep.ID)

	// 创建新任务回退到上一个步骤
	newTask := task_aggregate.NewTask(instance.InstanceId, instance.WorkflowID)
	newTask.TaskType = previousStep.Type
	newTask.Description = previousStep.Description

	s.domainService.ApplyStepParamsToTask(newTask, previousStep, instance)

	// 设置任务分配：优先使用上一个任务的处理人
	previousTaskAssignee := previousTask.Assignee
	if previousTask.Assignee != 0 {
		newTask.Assignee = previousTask.Assignee
		log.Printf("[EngineService] Set assignee from previous task: %s", previousTask.Assignee)
	}

	// 构建任务历史和任务数据
	taskHistories := s.domainService.BuildTaskHistories(tasks)
	newTask.TaskData = s.domainService.BuildTaskData(instance, taskHistories, nil)

	// 保存新任务
	if err := s.taskRepo.Save(ctx, newTask); err != nil {
		return fmt.Errorf("failed to save new task: %w", err)
	}

	log.Printf("[EngineService] Created new task for previous step: %s", newTask.TaskID.String())

	// 发送通知（如果有通知服务）
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyTaskAssigned(ctx, newTask, previousTaskAssignee)
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
	for _, parallelStep := range step.ParallelTasks {
		// 检查条件
		if parallelStep.Condition != "" {
			if !s.domainService.EvaluateCondition(parallelStep.Condition, instance) {
				log.Printf("[EngineService] Parallel task condition not met: %s", parallelStep.Condition)
				continue
			}
		}

		// 执行并行步骤（通常是 user_task）
		if parallelStep.Type == "userTask" {
			if err := s.executeUserTask(ctx, instance, &parallelStep); err != nil {
				log.Printf("[EngineService] Failed to create parallel task %s: %v", parallelStep.Name, err)
				continue
			}
			createdTasks = append(createdTasks, parallelStep.ID)
		}
		if parallelStep.Type == "process" {
			if err := s.executeProcessTask(ctx, instance, &parallelStep, definition); err != nil {
				log.Printf("[EngineService] Failed to execute parallel task %s: %v", parallelStep.Name, err)
				continue
			}
			createdTasks = append(createdTasks, parallelStep.ID)
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
	// 注意：rejected 状态的任务也视为"已完成"，因为它已经被回退任务替代
	for _, task := range parallelTasks {
		if task.Status != status.TaskStatusCompleted && task.Status != status.TaskStatusRejected {
			log.Printf("[EngineService] Parallel task %s not completed yet (status: %s)", task.TaskID.String(), task.Status)
			return false, nil
		}
	}

	log.Printf("[EngineService] All %d parallel tasks completed", len(parallelTasks))
	return true, nil
}
