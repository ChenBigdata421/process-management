package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// WorkflowEngineService 工作流引擎服务
// 负责处理工作流实例的执行，包括用户任务的创建和流程推进
type WorkflowEngineService struct {
	workflowRepo    WorkflowRepository
	instanceRepo    WorkflowInstanceRepository
	taskRepo        TaskRepository
	userClient      UserServiceClient   // 用户服务客户端接口（可选）
	notificationSvc NotificationService // 通知服务（可选）
}

// UserServiceClient 用户服务客户端接口
type UserServiceClient interface {
	ResolveUserID(ctx context.Context, identifier string) (int32, error)
	ResolveOrgID(ctx context.Context, identifier string) (int32, error)
}

// NewWorkflowEngineService 创建工作流引擎服务
func NewWorkflowEngineService(
	workflowRepo WorkflowRepository,
	instanceRepo WorkflowInstanceRepository,
	taskRepo TaskRepository,
) *WorkflowEngineService {
	return &WorkflowEngineService{
		workflowRepo:    workflowRepo,
		instanceRepo:    instanceRepo,
		taskRepo:        taskRepo,
		userClient:      nil,                          // 可选，如果不设置则不使用用户服务
		notificationSvc: NewNoOpNotificationService(), // 默认使用空操作通知服务
	}
}

// SetUserClient 设置用户服务客户端
func (s *WorkflowEngineService) SetUserClient(client UserServiceClient) {
	s.userClient = client
}

// SetNotificationService 设置通知服务
func (s *WorkflowEngineService) SetNotificationService(svc NotificationService) {
	s.notificationSvc = svc
}

// StepDefinition 步骤定义
type StepDefinition struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Condition     string                 `json:"condition"` // 执行条件
	Timeout       int                    `json:"timeout"`
	Retries       int                    `json:"retries"`
	Params        map[string]interface{} `json:"params"`
	NextSteps     []string               `json:"next_steps"`     // 下一步步骤ID列表（支持并行）
	ParallelTasks []StepDefinition       `json:"parallel_tasks"` // 并行任务列表
}

// WorkflowDefinitionStruct 工作流定义结构
type WorkflowDefinitionStruct struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Steps       []StepDefinition `json:"steps"`
}

// StartInstance 启动工作流实例并执行第一步
func (s *WorkflowEngineService) StartInstance(ctx context.Context, instanceID string) error {
	log.Printf("[EngineService] Starting instance: %s", instanceID)

	// 获取实例
	instance, err := s.instanceRepo.FindByID(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to find instance: %w", err)
	}

	if instance == nil {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	// 获取工作流定义
	wf, err := s.workflowRepo.FindByID(ctx, instance.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to find workflow: %w", err)
	}

	if wf == nil {
		return fmt.Errorf("workflow not found: %s", instance.WorkflowID)
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
	instance.Status = InstanceStatusRunning
	instance.StartedAt = time.Now()

	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	log.Printf("[EngineService] Instance started, executing first step: %s", definition.Steps[0].Name)

	// 执行第一步
	return s.ExecuteStep(ctx, instance, &definition.Steps[0], &definition)
}

// ExecuteStep 执行工作流步骤
func (s *WorkflowEngineService) ExecuteStep(ctx context.Context, instance *WorkflowInstance, step *StepDefinition, definition *WorkflowDefinitionStruct) error {
	log.Printf("[EngineService] Executing step: %s (type: %s) for instance: %s", step.Name, step.Type, instance.ID)

	switch step.Type {
	case "user_task":
		return s.executeUserTask(ctx, instance, step)
	case "process":
		return s.executeProcessTask(ctx, instance, step, definition)
	case "parallel":
		return s.executeParallelTasks(ctx, instance, step, definition)
	case "complete":
		return s.completeInstance(ctx, instance)
	default:
		log.Printf("[EngineService] Unknown step type: %s, skipping", step.Type)
		// 未知类型，尝试执行下一步
		return s.executeNextStep(ctx, instance, step, definition)
	}
}

// executeUserTask 执行用户任务步骤
func (s *WorkflowEngineService) executeUserTask(ctx context.Context, instance *WorkflowInstance, step *StepDefinition) error {
	log.Printf("[EngineService] Creating user task for step: %s", step.Name)

	// 创建用户任务
	task := NewTask(instance.ID, instance.WorkflowID, step.Name, step.ID)
	task.Description = step.Description

	// 从步骤参数中提取任务分配信息
	if step.Params != nil {
		// 处理 assignee
		if assignee, ok := step.Params["assignee"].(string); ok {
			task.Assignee = s.resolveVariable(assignee, instance)
		}

		// 处理 candidate_users
		if candidateUsers, ok := step.Params["candidate_users"].([]interface{}); ok {
			for _, user := range candidateUsers {
				if userStr, ok := user.(string); ok {
					task.CandidateUsers = append(task.CandidateUsers, userStr)
				}
			}
		}

		// 处理 candidate_groups
		if candidateGroups, ok := step.Params["candidate_groups"].([]interface{}); ok {
			for _, group := range candidateGroups {
				if groupStr, ok := group.(string); ok {
					task.CandidateGroups = append(task.CandidateGroups, groupStr)
				}
			}
		}

		// 处理优先级
		if priority, ok := step.Params["priority"].(string); ok {
			switch priority {
			case "high":
				task.Priority = TaskPriorityHigh
			case "low":
				task.Priority = TaskPriorityLow
			default:
				task.Priority = TaskPriorityMedium
			}
		}

		// 处理表单字段
		if formFields, ok := step.Params["form_fields"].([]interface{}); ok {
			// 将 form_fields 数组转换为字符串数组
			fields := make([]string, 0, len(formFields))
			for _, field := range formFields {
				if fieldStr, ok := field.(string); ok {
					fields = append(fields, fieldStr)
				}
			}

			// 构造 FormData 对象，包含 form_fields 数组
			formData := map[string]interface{}{
				"form_fields": fields,
			}
			formDataJSON, _ := json.Marshal(formData)
			task.FormData = formDataJSON
		}
	}

	// 设置任务数据：合并实例输入和所有前序任务的输出
	taskData := make(map[string]interface{})

	// 1. 首先加载实例输入数据
	if len(instance.Input) > 0 {
		if err := json.Unmarshal(instance.Input, &taskData); err != nil {
			log.Printf("[EngineService] Failed to parse instance input: %v", err)
		}
	}

	// 2. 查找所有已处理的任务（包括已完成和已驳回），按处理时间排序
	allTasks, err := s.taskRepo.FindByInstanceID(ctx, instance.ID, 100, 0)
	if err == nil && len(allTasks) > 0 {
		// 收集所有已处理任务的历史记录
		type TaskHistory struct {
			TaskName    string                 `json:"task_name"`
			TaskKey     string                 `json:"task_key"`
			Assignee    string                 `json:"assignee"`
			Status      string                 `json:"status"`
			Result      string                 `json:"result"`
			Comment     string                 `json:"comment"`
			Output      map[string]interface{} `json:"output"`
			CompletedAt string                 `json:"completed_at"`
		}

		var taskHistories []TaskHistory

		// 遍历所有任务，找出已处理的任务（已完成或已驳回）
		for _, t := range allTasks {
			// 只包含已完成或已驳回的任务
			if (t.Status == TaskStatusCompleted || t.Status == TaskStatusRejected) && t.CompletedAt != nil {
				var output map[string]interface{}
				// 解析输出数据（可能为空）
				if len(t.Output) > 0 {
					if err := json.Unmarshal(t.Output, &output); err != nil {
						output = make(map[string]interface{})
					}
				} else {
					output = make(map[string]interface{})
				}

				// 确定处理结果的中文描述
				resultText := ""
				switch t.Result {
				case TaskResultApproved:
					resultText = "通过"
				case TaskResultRejected:
					resultText = "驳回"
				case TaskResultCompleted:
					resultText = "完成"
				default:
					resultText = string(t.Result)
				}

				taskHistories = append(taskHistories, TaskHistory{
					TaskName:    t.TaskName,
					TaskKey:     t.TaskKey,
					Assignee:    t.Assignee,
					Status:      string(t.Status),
					Result:      resultText,
					Comment:     t.Comment,
					Output:      output,
					CompletedAt: t.CompletedAt.Format("2006-01-02 15:04:05"),
				})
			}
		}

		// 按完成时间排序（使用冒泡排序，因为任务数量通常不多）
		for i := 0; i < len(taskHistories)-1; i++ {
			for j := 0; j < len(taskHistories)-i-1; j++ {
				// 比较完成时间字符串（格式：2006-01-02 15:04:05）
				if taskHistories[j].CompletedAt > taskHistories[j+1].CompletedAt {
					taskHistories[j], taskHistories[j+1] = taskHistories[j+1], taskHistories[j]
				}
			}
		}

		// 将所有前序任务的历史记录添加到任务数据中
		if len(taskHistories) > 0 {
			taskData["previous_tasks_history"] = taskHistories
			log.Printf("[EngineService] Added %d previous task histories to new task data", len(taskHistories))
		}
	}

	// 3. 将合并后的数据序列化为JSON
	taskDataJSON, _ := json.Marshal(taskData)
	task.TaskData = taskDataJSON

	// 保存任务
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	log.Printf("[EngineService] User task created: %s (ID: %s)", task.TaskName, task.ID)

	// 发送任务创建通知
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyTaskCreated(ctx, task)
	}

	// 更新实例状态（保持运行中，因为可能有多个并行任务）
	// 注意：这里不改变状态，因为实例仍在运行中，只是在等待用户任务完成

	log.Printf("[EngineService] Instance paused, waiting for task completion")

	return nil
}

// executeProcessTask 执行自动化处理任务
func (s *WorkflowEngineService) executeProcessTask(ctx context.Context, instance *WorkflowInstance, step *StepDefinition, definition *WorkflowDefinitionStruct) error {
	log.Printf("[EngineService] Executing automated process task: %s", step.Name)

	// 自动化任务直接执行完成，继续下一步
	// 这里可以根据需要添加实际的处理逻辑

	return s.executeNextStep(ctx, instance, step, definition)
}

// completeInstance 完成工作流实例
func (s *WorkflowEngineService) completeInstance(ctx context.Context, instance *WorkflowInstance) error {
	log.Printf("[EngineService] Completing instance: %s", instance.ID)

	now := time.Now()
	instance.Status = InstanceStatusCompleted
	instance.CompletedAt = &now

	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	log.Printf("[EngineService] Instance completed successfully")

	return nil
}

// ContinueAfterTask 任务完成后继续执行流程
func (s *WorkflowEngineService) ContinueAfterTask(ctx context.Context, task *Task) error {
	log.Printf("[EngineService] Continuing workflow after task completion: %s", task.ID)

	// 获取实例
	instance, err := s.instanceRepo.FindByID(ctx, task.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to find instance: %w", err)
	}

	if instance == nil {
		return fmt.Errorf("instance not found: %s", task.InstanceID)
	}

	// 获取工作流定义
	wf, err := s.workflowRepo.FindByID(ctx, instance.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to find workflow: %w", err)
	}

	if wf == nil {
		return fmt.Errorf("workflow not found: %s", instance.WorkflowID)
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
	// 不能简单地根据 task_key 包含 "_" 来判断，因为步骤ID本身就可能包含 "_"
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
	instance.Status = InstanceStatusRunning

	if err := s.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	log.Printf("[EngineService] Instance resumed, finding next step")

	// 执行下一步
	return s.executeNextStep(ctx, instance, currentStep, &definition)
}

// RejectAndGoBack 驳回任务并回退到上一个步骤
func (s *WorkflowEngineService) RejectAndGoBack(ctx context.Context, task *Task) error {
	log.Printf("[EngineService] Rejecting task and going back: %s", task.ID)

	// 获取实例
	instance, err := s.instanceRepo.FindByID(ctx, task.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to find instance: %w", err)
	}

	if instance == nil {
		return fmt.Errorf("instance not found: %s", task.InstanceID)
	}

	// 获取工作流定义
	wf, err := s.workflowRepo.FindByID(ctx, instance.WorkflowID)
	if err != nil {
		return fmt.Errorf("failed to find workflow: %w", err)
	}

	if wf == nil {
		return fmt.Errorf("workflow not found: %s", instance.WorkflowID)
	}

	// 解析工作流定义
	var definition WorkflowDefinitionStruct
	if err := json.Unmarshal([]byte(wf.Definition), &definition); err != nil {
		return fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	// 查找上一个步骤
	previousStep := s.findPreviousStep(task.TaskKey, &definition)
	if previousStep == nil {
		return fmt.Errorf("cannot reject first task, no previous step found")
	}

	log.Printf("[EngineService] Found previous step: %s (%s)", previousStep.Name, previousStep.ID)

	// 查找上一个步骤的已完成任务，获取原处理人
	previousTasks, err := s.taskRepo.FindByInstanceID(ctx, instance.ID, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to find previous tasks: %w", err)
	}

	var previousTaskAssignee string
	for _, t := range previousTasks {
		if t.TaskKey == previousStep.ID && t.Status == TaskStatusCompleted {
			previousTaskAssignee = t.Assignee
			log.Printf("[EngineService] Found previous task assignee: %s", previousTaskAssignee)
			break
		}
	}

	// 如果找不到上一个任务的处理人，尝试从步骤定义中解析
	if previousTaskAssignee == "" {
		if assignee, ok := previousStep.Params["assignee"].(string); ok {
			previousTaskAssignee = s.resolveVariable(assignee, instance)
			log.Printf("[EngineService] Resolved assignee from step params: %s", previousTaskAssignee)
		}
	}

	// 创建新任务回退到上一个步骤
	newTask := NewTask(instance.ID, instance.WorkflowID, previousStep.Name, previousStep.ID)
	newTask.TaskType = previousStep.Type
	newTask.Description = previousStep.Description

	// 设置任务分配
	if previousTaskAssignee != "" {
		newTask.Assignee = previousTaskAssignee
	} else if candidateGroups, ok := previousStep.Params["candidate_groups"].([]interface{}); ok {
		groups := make([]string, 0, len(candidateGroups))
		for _, g := range candidateGroups {
			if groupStr, ok := g.(string); ok {
				groups = append(groups, groupStr)
			}
		}
		newTask.CandidateGroups = groups
	} else if candidateUsers, ok := previousStep.Params["candidate_users"].([]interface{}); ok {
		users := make([]string, 0, len(candidateUsers))
		for _, u := range candidateUsers {
			if userStr, ok := u.(string); ok {
				users = append(users, userStr)
			}
		}
		newTask.CandidateUsers = users
	}

	// 设置优先级
	if priority, ok := previousStep.Params["priority"].(string); ok {
		newTask.Priority = TaskPriority(priority)
	}

	// 处理表单字段
	if formFields, ok := previousStep.Params["form_fields"].([]interface{}); ok {
		fields := make([]string, 0, len(formFields))
		for _, field := range formFields {
			if fieldStr, ok := field.(string); ok {
				fields = append(fields, fieldStr)
			}
		}

		// 构造 FormData 对象，包含 form_fields 数组
		formData := map[string]interface{}{
			"form_fields": fields,
		}
		formDataJSON, _ := json.Marshal(formData)
		newTask.FormData = formDataJSON
	}

	// 设置任务数据：合并实例输入、前序任务历史和驳回信息
	taskData := make(map[string]interface{})

	// 1. 首先加载实例输入数据
	if len(instance.Input) > 0 {
		if err := json.Unmarshal(instance.Input, &taskData); err != nil {
			log.Printf("[EngineService] Failed to parse instance input: %v", err)
		}
	}

	// 2. 添加驳回信息
	rejectionInfo := map[string]interface{}{
		"rejected_by":      task.Assignee,
		"rejected_at":      task.CompletedAt.Format("2006-01-02 15:04:05"),
		"rejection_reason": task.Comment,
		"rejected_task_id": task.ID,
	}
	taskData["rejected_by"] = rejectionInfo["rejected_by"]
	taskData["rejected_at"] = rejectionInfo["rejected_at"]
	taskData["rejection_reason"] = rejectionInfo["rejection_reason"]
	taskData["rejected_task_id"] = rejectionInfo["rejected_task_id"]

	// 3. 查找所有已处理的任务（包括已完成和已驳回），按处理时间排序
	allTasks, err := s.taskRepo.FindByInstanceID(ctx, instance.ID, 100, 0)
	if err == nil && len(allTasks) > 0 {
		// 收集所有已处理任务的历史记录
		type TaskHistory struct {
			TaskName    string                 `json:"task_name"`
			TaskKey     string                 `json:"task_key"`
			Assignee    string                 `json:"assignee"`
			Status      string                 `json:"status"`
			Result      string                 `json:"result"`
			Comment     string                 `json:"comment"`
			Output      map[string]interface{} `json:"output"`
			CompletedAt string                 `json:"completed_at"`
		}

		var taskHistories []TaskHistory

		// 遍历所有任务，找出已处理的任务（已完成或已驳回）
		for _, t := range allTasks {
			// 只包含已完成或已驳回的任务
			if (t.Status == TaskStatusCompleted || t.Status == TaskStatusRejected) && t.CompletedAt != nil {
				var output map[string]interface{}
				// 解析输出数据（可能为空）
				if len(t.Output) > 0 {
					if err := json.Unmarshal(t.Output, &output); err != nil {
						output = make(map[string]interface{})
					}
				} else {
					output = make(map[string]interface{})
				}

				// 确定处理结果的中文描述
				resultText := ""
				switch t.Result {
				case TaskResultApproved:
					resultText = "通过"
				case TaskResultRejected:
					resultText = "驳回"
				case TaskResultCompleted:
					resultText = "完成"
				default:
					resultText = string(t.Result)
				}

				taskHistories = append(taskHistories, TaskHistory{
					TaskName:    t.TaskName,
					TaskKey:     t.TaskKey,
					Assignee:    t.Assignee,
					Status:      string(t.Status),
					Result:      resultText,
					Comment:     t.Comment,
					Output:      output,
					CompletedAt: t.CompletedAt.Format("2006-01-02 15:04:05"),
				})
			}
		}

		// 按完成时间排序
		for i := 0; i < len(taskHistories)-1; i++ {
			for j := 0; j < len(taskHistories)-i-1; j++ {
				if taskHistories[j].CompletedAt > taskHistories[j+1].CompletedAt {
					taskHistories[j], taskHistories[j+1] = taskHistories[j+1], taskHistories[j]
				}
			}
		}

		// 将所有前序任务的历史记录添加到任务数据中
		if len(taskHistories) > 0 {
			taskData["previous_tasks_history"] = taskHistories
			log.Printf("[EngineService] Added %d previous task histories to rollback task data", len(taskHistories))
		}
	}

	// 4. 将合并后的数据序列化为JSON
	taskDataJSON, _ := json.Marshal(taskData)
	newTask.TaskData = taskDataJSON

	// 保存新任务
	if err := s.taskRepo.Save(ctx, newTask); err != nil {
		return fmt.Errorf("failed to save new task: %w", err)
	}

	log.Printf("[EngineService] Created new task for previous step: %s", newTask.ID)

	// 发送通知（如果有通知服务）
	if s.notificationSvc != nil {
		s.notificationSvc.NotifyTaskAssigned(ctx, newTask, previousTaskAssignee)
	}

	log.Printf("[EngineService] Task rejection and rollback completed successfully")
	return nil
}

// executeNextStep 执行下一个步骤
func (s *WorkflowEngineService) executeNextStep(ctx context.Context, instance *WorkflowInstance, currentStep *StepDefinition, definition *WorkflowDefinitionStruct) error {
	// 找到下一个步骤
	nextStep := s.findNextStep(currentStep, definition, instance)

	if nextStep == nil {
		// 没有下一步，完成流程
		log.Printf("[EngineService] No next step found, completing instance")
		return s.completeInstance(ctx, instance)
	}

	log.Printf("[EngineService] Found next step: %s", nextStep.Name)

	// 执行下一步
	return s.ExecuteStep(ctx, instance, nextStep, definition)
}

// findNextStep 查找下一个步骤
func (s *WorkflowEngineService) findNextStep(currentStep *StepDefinition, definition *WorkflowDefinitionStruct, instance *WorkflowInstance) *StepDefinition {
	// 优先使用 next_steps 字段（支持条件分支）
	if len(currentStep.NextSteps) > 0 {
		// 遍历所有可能的下一步，找到第一个满足条件的
		for _, nextStepID := range currentStep.NextSteps {
			nextStep := s.findStepByID(nextStepID, definition)
			if nextStep == nil {
				log.Printf("[EngineService] Next step not found: %s", nextStepID)
				continue
			}

			// 检查步骤条件
			if nextStep.Condition != "" {
				if !s.evaluateCondition(nextStep.Condition, instance) {
					log.Printf("[EngineService] Step condition not met for %s: %s", nextStep.ID, nextStep.Condition)
					continue
				}
			}

			log.Printf("[EngineService] Found next step via next_steps: %s", nextStep.Name)
			return nextStep
		}

		// 所有条件都不满足
		log.Printf("[EngineService] No next step condition satisfied")
		return nil
	}

	// 回退到顺序执行（兼容旧格式）
	currentIndex := -1
	for i, step := range definition.Steps {
		if step.ID == currentStep.ID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 || currentIndex >= len(definition.Steps)-1 {
		return nil
	}

	// 返回下一个步骤
	nextStep := &definition.Steps[currentIndex+1]

	// 检查步骤条件
	if nextStep.Condition != "" {
		if !s.evaluateCondition(nextStep.Condition, instance) {
			log.Printf("[EngineService] Step condition not met: %s, skipping", nextStep.Condition)
			// 条件不满足，继续查找下一个步骤
			return s.findNextStep(nextStep, definition, instance)
		}
	}

	return nextStep
}

// findStepByID 根据ID查找步骤
func (s *WorkflowEngineService) findStepByID(stepID string, definition *WorkflowDefinitionStruct) *StepDefinition {
	for i := range definition.Steps {
		if definition.Steps[i].ID == stepID {
			return &definition.Steps[i]
		}
	}
	return nil
}

// findPreviousStep 查找上一个步骤
func (s *WorkflowEngineService) findPreviousStep(currentStepID string, definition *WorkflowDefinitionStruct) *StepDefinition {
	// 按顺序查找上一个步骤
	for i, step := range definition.Steps {
		if step.ID == currentStepID {
			if i > 0 {
				return &definition.Steps[i-1]
			}
			// 已经是第一个步骤，没有上一步
			return nil
		}
	}
	return nil
}

// evaluateCondition 评估条件表达式
func (s *WorkflowEngineService) evaluateCondition(condition string, instance *WorkflowInstance) bool {
	if condition == "" {
		return true
	}

	// 使用条件求值器
	evaluator := NewConditionEvaluator(instance, s.taskRepo)
	result, err := evaluator.Evaluate(condition)
	if err != nil {
		log.Printf("[EngineService] Failed to evaluate condition '%s': %v, defaulting to false", condition, err)
		return false
	}

	log.Printf("[EngineService] Condition '%s' evaluated to: %v", condition, result)
	return result
}

// resolveVariable 解析变量
// 支持 ${variable} 格式的变量替换
func (s *WorkflowEngineService) resolveVariable(value string, instance *WorkflowInstance) string {
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return value
	}

	// 提取变量名
	varName := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")

	// 解析实例输入
	var input map[string]interface{}
	inputData := instance.Input

	// 检查是否是双重编码的JSON字符串
	// 如果Input以引号开头，说明它是一个被编码的字符串，需要先解码一次
	inputStr := string(inputData)
	if strings.HasPrefix(inputStr, "\"") && strings.HasSuffix(inputStr, "\"") {
		var tempStr string
		if err := json.Unmarshal(inputData, &tempStr); err == nil {
			inputData = []byte(tempStr)
		}
	}

	if err := json.Unmarshal(inputData, &input); err != nil {
		log.Printf("[EngineService] Failed to parse instance input: %v (input: %s)", err, string(inputData))
		return value
	}

	// 查找变量值
	if val, ok := input[varName]; ok {
		if strVal, ok := val.(string); ok {
			log.Printf("[EngineService] Resolved variable %s = %s", varName, strVal)
			return strVal
		}
	}

	log.Printf("[EngineService] Variable not found: %s", varName)
	return value
}

// executeParallelTasks 执行并行任务
func (s *WorkflowEngineService) executeParallelTasks(ctx context.Context, instance *WorkflowInstance, step *StepDefinition, definition *WorkflowDefinitionStruct) error {
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
			if !s.evaluateCondition(parallelStep.Condition, instance) {
				log.Printf("[EngineService] Parallel task condition not met: %s", parallelStep.Condition)
				continue
			}
		}

		// 执行并行步骤（通常是 user_task）
		if parallelStep.Type == "user_task" {
			if err := s.executeUserTask(ctx, instance, &parallelStep); err != nil {
				log.Printf("[EngineService] Failed to create parallel task %s: %v", parallelStep.Name, err)
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
func (s *WorkflowEngineService) checkParallelTasksCompleted(ctx context.Context, instance *WorkflowInstance, stepID string) (bool, error) {
	// 查找该实例的所有任务
	tasks, err := s.taskRepo.FindByInstanceID(ctx, instance.ID, 100, 0)
	if err != nil {
		return false, fmt.Errorf("failed to find tasks: %v", err)
	}

	// 查找属于该并行步骤的所有任务
	var parallelTasks []*Task
	for _, task := range tasks {
		// 检查任务的 task_key 是否以 stepID 开头（表示是该并行步骤的子任务）
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
		if task.Status != TaskStatusCompleted && task.Status != TaskStatusRejected {
			log.Printf("[EngineService] Parallel task %s not completed yet (status: %s)", task.ID, task.Status)
			return false, nil
		}
	}

	log.Printf("[EngineService] All %d parallel tasks completed", len(parallelTasks))
	return true, nil
}
