package domain_service

import (
	"encoding/json"
	command "jxt-evidence-system/process-management/internal/application/command"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	task_repository "jxt-evidence-system/process-management/internal/domain/aggregate/task/repository"
	"jxt-evidence-system/process-management/internal/domain/valueobject"
	"jxt-evidence-system/process-management/shared/common/status"
	"log"
	"strconv"
	"strings"
)

// WorkflowDomainService 工作流领域服务
// 负责工作流相关的领域逻辑，不涉及应用协调
type WorkflowDomainService struct {
	taskRepo task_repository.TaskRepository
}

func NewWorkflowDomainService(taskRepo task_repository.TaskRepository) *WorkflowDomainService {
	return &WorkflowDomainService{
		taskRepo: taskRepo,
	}
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
	NextSteps     []string               `json:"nextSteps"`     // 下一步步骤ID列表（支持并行）
	ParallelTasks []StepDefinition       `json:"parallelTasks"` // 并行任务列表
	IsRoot        bool                   `json:"isRoot"`
}

// WorkflowDefinitionStruct 工作流定义结构
type WorkflowDefinitionStruct struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Steps       []StepDefinition `json:"steps"`
}

// buildTaskHistories 构建任务历史列表
func (s *WorkflowDomainService) BuildTaskHistories(tasks []*task_aggregate.Task) []command.TaskHistoryItem {
	var taskHistories []command.TaskHistoryItem

	for _, t := range tasks {
		if (t.Status == status.TaskStatusCompleted || t.Status == status.TaskStatusRejected) && t.CompletedAt != nil {
			var output map[string]interface{}
			if len(t.Output) > 0 {
				if err := json.Unmarshal(t.Output, &output); err != nil {
					output = make(map[string]interface{})
				}
			} else {
				output = make(map[string]interface{})
			}

			resultText := ""
			switch t.Result {
			case status.TaskResultApproved:
				resultText = "通过"
			case status.TaskResultRejected:
				resultText = "驳回"
			case status.TaskResultCompleted:
				resultText = "完成"
			default:
				resultText = string(t.Result)
			}

			taskHistories = append(taskHistories, command.TaskHistoryItem{
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

	return taskHistories
}

// buildTaskHistories 构建实例任务列表，用于展示实例详情
func (s *WorkflowDomainService) BuildInstanceDetail(tasks []*task_aggregate.Task) []command.TaskHistoryItem {
	var taskHistories []command.TaskHistoryItem

	for _, t := range tasks {
		var output map[string]interface{}
		if len(t.Output) > 0 {
			if err := json.Unmarshal(t.Output, &output); err != nil {
				output = make(map[string]interface{})
			}
		} else {
			output = make(map[string]interface{})
		}

		resultText := ""
		switch t.Result {
		case status.TaskResultApproved:
			resultText = "通过"
		case status.TaskResultRejected:
			resultText = "驳回"
		case status.TaskResultCompleted:
			resultText = "完成"
		default:
			resultText = string(t.Result)
		}

		if t.Status == status.TaskStatusPending {
			resultText = "待处理"
		}

		completedAtStr := ""
		if t.CompletedAt != nil {
			completedAtStr = t.CompletedAt.Format("2006-01-02 15:04:05")
		}
		createdAtStr := t.CreatedAt.Format("2006-01-02 15:04:05")

		taskHistories = append(taskHistories, command.TaskHistoryItem{
			TaskName:    t.TaskName,
			TaskKey:     t.TaskKey,
			Assignee:    t.Assignee,
			Status:      string(t.Status),
			Result:      resultText,
			Comment:     t.Comment,
			Output:      output,
			CompletedAt: completedAtStr,
			CreatedAt:   createdAtStr,
		})

	}

	// 按创建时间排序
	for i := 0; i < len(taskHistories); i++ {
		for j := 0; j < len(taskHistories)-i-1; j++ {
			if taskHistories[j].CreatedAt > taskHistories[j+1].CreatedAt {
				taskHistories[j], taskHistories[j+1] = taskHistories[j+1], taskHistories[j]
			}
		}
	}

	return taskHistories
}

// applyStepParamsToTask 从步骤参数设置任务属性
func (s *WorkflowDomainService) ApplyStepParamsToTask(task *task_aggregate.Task, step *StepDefinition, instance *instance_aggregate.WorkflowInstance) {
	if step.Params == nil {
		log.Printf("[WorkflowDomainService] Step params is nil for step: %s", step.Name)
		return
	}
	task.TaskName = step.Name
	task.TaskKey = step.ID
	task.Description = step.Description
	task.IsFirstStep = step.IsRoot
	task.TaskType = step.Type
	// 处理 assignee
	if assignee, ok := step.Params["assignee"].(string); ok {
		log.Printf("[WorkflowDomainService] Found assignee param: %s", assignee)
		resolvedValue := s.resolveVariable(assignee, instance)
		log.Printf("[WorkflowDomainService] Resolved assignee value: %s", resolvedValue)
		// 尝试将解析后的值转换为 int
		if assigneeInt, err := strconv.Atoi(resolvedValue); err == nil {
			task.Assignee = assigneeInt
			log.Printf("[WorkflowDomainService] Set task assignee to: %d", assigneeInt)
		} else {
			log.Printf("[WorkflowDomainService] Failed to convert assignee to int: %s (error: %v)", resolvedValue, err)
			return
		}
	} else {
		log.Printf("[WorkflowDomainService] No assignee param found in step: %s, params: %v", step.Name, step.Params)
	}

	// 处理优先级
	if priority, ok := step.Params["priority"].(string); ok {
		switch priority {
		case "high":
			task.Priority = status.TaskPriorityHigh
		case "low":
			task.Priority = status.TaskPriorityLow
		default:
			task.Priority = status.TaskPriorityMedium
		}
	}

	// 处理表单字段
	if formFields, ok := step.Params["formFields"].([]interface{}); ok {
		fields := make([]string, 0, len(formFields))
		for _, field := range formFields {
			if fieldStr, ok := field.(string); ok {
				fields = append(fields, fieldStr)
			}
		}

		formData := map[string]interface{}{
			"formFields": fields,
		}
		formDataJSON, _ := json.Marshal(formData)
		task.FormData = formDataJSON
	}
}

// buildTaskData 构建任务数据（合并实例输入和历史记录）
func (s *WorkflowDomainService) BuildTaskData(instance *instance_aggregate.WorkflowInstance, taskHistories []command.TaskHistoryItem, extraData map[string]interface{}) []byte {
	taskData := make(map[string]interface{})

	// 1. 加载实例输入数据
	if len(instance.Input) > 0 {
		if err := json.Unmarshal(instance.Input, &taskData); err != nil {
			log.Printf("[EngineService] Failed to parse instance input: %v", err)
		}
	}

	// 2. 添加额外数据（如驳回信息）
	for k, v := range extraData {
		taskData[k] = v
	}

	// 3. 添加任务历史
	if len(taskHistories) > 0 {
		taskData["previousTasksHistory"] = taskHistories
		log.Printf("[EngineService] Added %d previous task histories to task data", len(taskHistories))
	}

	taskDataJSON, _ := json.Marshal(taskData)
	return taskDataJSON
}

// PreviousTaskInfo 上一个任务的信息
type PreviousTaskInfo struct {
	IsParallelGroup bool                   // 是否为并行任务组
	ParallelTasks   []*task_aggregate.Task // 并行任务组的所有任务（如果是并行组）
	SingleTask      *task_aggregate.Task   // 单个任务（如果不是并行组）
	StepID          string                 // 步骤ID
}

// extractParallelGroupPrefix 从 taskKey 中提取并行组前缀
// 如果 taskKey 是并行任务（格式：parentID_index，其中 index 是 1-9），返回 parentID
// 否则返回空字符串
// 注意：并行任务的 index 必须是单个数字（1-9），以区别于普通任务的时间戳后缀
func (s *WorkflowDomainService) extractParallelGroupPrefix(taskKey string) string {
	if taskKey == "" {
		return ""
	}

	idx := strings.LastIndex(taskKey, "_")
	if idx <= 0 {
		return ""
	}

	potentialSuffix := taskKey[idx+1:]

	// 并行任务的 index 必须是单个数字（1-9）
	if len(potentialSuffix) != 1 {
		return ""
	}

	index, err := strconv.Atoi(potentialSuffix)
	if err != nil || index < 1 || index > 9 {
		return ""
	}

	return taskKey[:idx]
}

// findCurrentTask 在任务列表中查找指定ID的任务
func (s *WorkflowDomainService) findCurrentTask(tasks []*task_aggregate.Task, taskID valueobject.TaskID) *task_aggregate.Task {
	for _, t := range tasks {
		if t.TaskID == taskID {
			return t
		}
	}
	return nil
}

// findPreviousCompletedTaskInList 从后向前查找上一个已完成的任务
// 支持跳过并行任务组中的兄弟任务
func (s *WorkflowDomainService) findPreviousCompletedTaskInList(tasks []*task_aggregate.Task, currentTask *task_aggregate.Task, parallelGroupPrefix string) *task_aggregate.Task {
	for i := len(tasks) - 1; i >= 0; i-- {
		t := tasks[i]

		// 跳过当前任务
		if t.TaskID == currentTask.TaskID {
			continue
		}

		// 如果当前任务是并行任务，跳过同一并行组内的其他任务
		if parallelGroupPrefix != "" && strings.HasPrefix(t.TaskKey, parallelGroupPrefix+"_") {
			log.Printf("[WorkflowDomainService] Skipping parallel sibling task: %s", t.TaskKey)
			continue
		}

		// 找到已完成的任务,注意去除同步骤的任务
		if t.Status == status.TaskStatusCompleted && t.CompletedAt != nil && t.TaskKey != currentTask.TaskKey {
			log.Printf("[WorkflowDomainService] Found previous completed task: %s (TaskKey: %s)", t.TaskName, t.TaskKey)
			return t
		}
	}

	log.Printf("[WorkflowDomainService] No previous completed task found")
	return nil
}

// collectParallelGroupTasks 收集并行任务组的所有已完成任务
// 注意：当存在相同 TaskKey 的多个任务时，只保留最新的任务（按 CreatedAt 时间排序）
func (s *WorkflowDomainService) collectParallelGroupTasks(tasks []*task_aggregate.Task, parallelGroupPrefix string) []*task_aggregate.Task {
	// 使用 map 来存储每个 TaskKey 对应的最新任务
	taskKeyMap := make(map[string]*task_aggregate.Task)

	for _, t := range tasks {
		if strings.HasPrefix(t.TaskKey, parallelGroupPrefix+"_") && t.Status == status.TaskStatusCompleted {
			// 如果该 TaskKey 已存在，比较时间，保留最新的
			if existing, exists := taskKeyMap[t.TaskKey]; exists {
				// 比较创建时间，保留较新的任务
				if t.CreatedAt.After(existing.CreatedAt) {
					log.Printf("[WorkflowDomainService] Replacing older task %s (created: %v) with newer task (created: %v)",
						t.TaskKey, existing.CreatedAt, t.CreatedAt)
					taskKeyMap[t.TaskKey] = t
				}
			} else {
				taskKeyMap[t.TaskKey] = t
			}
		}
	}

	// 将 map 转换为切片
	var parallelTasks []*task_aggregate.Task
	for _, t := range taskKeyMap {
		parallelTasks = append(parallelTasks, t)
	}

	return parallelTasks
}

// FindPreviousTaskInfo 查找上一个步骤的完整信息
// 支持并行任务场景：如果上一步是并行任务组，返回该组的所有任务
func (s *WorkflowDomainService) FindPreviousTaskInfo(tasks []*task_aggregate.Task, currentTaskID valueobject.TaskID) *PreviousTaskInfo {
	// 1. 查找当前任务
	currentTask := s.findCurrentTask(tasks, currentTaskID)
	if currentTask == nil {
		log.Printf("[WorkflowDomainService] Current task not found: %s", currentTaskID.String())
		return nil
	}

	// 2. 识别当前任务是否为并行任务
	currentParallelGroupPrefix := s.extractParallelGroupPrefix(currentTask.TaskKey)
	if currentParallelGroupPrefix != "" {
		log.Printf("[WorkflowDomainService] Current task is parallel task, group prefix: %s", currentParallelGroupPrefix)
	}

	// 3. 查找上一个已完成的任务
	previousCompletedTask := s.findPreviousCompletedTaskInList(tasks, currentTask, currentParallelGroupPrefix)
	if previousCompletedTask == nil {
		return nil
	}

	log.Printf("[WorkflowDomainService] Found previous completed task: %s (TaskKey: %s)", previousCompletedTask.TaskName, previousCompletedTask.TaskKey)

	// 4. 识别上一个任务是否为并行任务组的一部分
	previousParallelGroupPrefix := s.extractParallelGroupPrefix(previousCompletedTask.TaskKey)
	if previousParallelGroupPrefix == "" {
		// 上一个任务不是并行任务，返回单个任务
		return &PreviousTaskInfo{
			IsParallelGroup: false,
			SingleTask:      previousCompletedTask,
			StepID:          previousCompletedTask.TaskKey,
		}
	}

	// 5. 收集并行任务组的所有任务
	log.Printf("[WorkflowDomainService] Previous task is part of parallel group: %s", previousParallelGroupPrefix)
	parallelTasks := s.collectParallelGroupTasks(tasks, previousParallelGroupPrefix)

	if len(parallelTasks) > 0 {
		log.Printf("[WorkflowDomainService] Found %d parallel tasks in group %s", len(parallelTasks), previousParallelGroupPrefix)
		return &PreviousTaskInfo{
			IsParallelGroup: true,
			ParallelTasks:   parallelTasks,
			StepID:          previousParallelGroupPrefix,
		}
	}

	// 如果没有找到并行任务组，返回单个任务
	return &PreviousTaskInfo{
		IsParallelGroup: false,
		SingleTask:      previousCompletedTask,
		StepID:          previousCompletedTask.TaskKey,
	}
}

// findPreviousCompletedTask 查找上一个已完成的任务（保留向后兼容）
// 支持并行任务场景：如果当前任务是并行任务，会跳过同一并行组内的其他任务
// 注意：此方法只返回单个任务，无法处理并行任务组场景，建议使用 FindPreviousStepInfo
func (s *WorkflowDomainService) FindPreviousCompletedTask(tasks []*task_aggregate.Task, currentTaskID valueobject.TaskID) *task_aggregate.Task {
	// 找到当前任务
	var currentTask *task_aggregate.Task
	for _, t := range tasks {
		if t.TaskID == currentTaskID {
			currentTask = t
			break
		}
	}

	if currentTask == nil {
		log.Printf("[WorkflowDomainService] Current task not found: %s", currentTaskID.String())
		return nil
	}

	// 检查当前任务是否是并行任务（taskKey 包含下划线表示是并行任务的子任务）
	// 例如：parallel_approval_1, parallel_approval_2
	var parallelGroupPrefix string
	if idx := strings.LastIndex(currentTask.TaskKey, "_"); idx > 0 {
		// 尝试提取并行组前缀
		// 如果最后一个下划线后面是数字，则认为是并行任务
		potentialSuffix := currentTask.TaskKey[idx+1:]
		if _, err := strconv.Atoi(potentialSuffix); err == nil {
			parallelGroupPrefix = currentTask.TaskKey[:idx]
			log.Printf("[WorkflowDomainService] Detected parallel task, group prefix: %s", parallelGroupPrefix)
		}
	}

	// 从后向前查找已完成的任务
	for i := len(tasks) - 1; i >= 0; i-- {
		t := tasks[i]

		// 跳过当前任务
		if t.TaskID == currentTaskID {
			continue
		}

		// 如果当前任务是并行任务，跳过同一并行组内的其他任务
		if parallelGroupPrefix != "" && strings.HasPrefix(t.TaskKey, parallelGroupPrefix+"_") {
			log.Printf("[WorkflowDomainService] Skipping parallel sibling task: %s", t.TaskKey)
			continue
		}

		// 找到已完成的任务
		if t.Status == status.TaskStatusCompleted && t.CompletedAt != nil {
			log.Printf("[WorkflowDomainService] Found previous completed task: %s (TaskKey: %s)", t.TaskName, t.TaskKey)
			return t
		}
	}

	log.Printf("[WorkflowDomainService] No previous completed task found")
	return nil
}

// FindStepDefinitionByTaskKey 根据 taskKey 查找步骤定义
// 支持递归查找：可以在嵌套的 ParallelTasks 中查找
// 注意：并行任务的 taskKey 是修改后的格式（如 parallel_approval_1），不是真实的步骤ID
// 因此本方法只能查找真实的步骤定义，不能用于查找并行任务子任务
func (s *WorkflowDomainService) FindStepDefinitionByTaskKey(taskKey string, definition *WorkflowDefinitionStruct) *StepDefinition {
	if taskKey == "" || definition == nil {
		return nil
	}

	// 尝试直接匹配顶层步骤或递归查找 ParallelTasks 中的步骤
	step := s.findStepInList(taskKey, definition.Steps)
	if step != nil {
		return step
	}

	log.Printf("[WorkflowDomainService] Step definition not found for taskKey: %s", taskKey)
	return nil
}

// findStepInList 在步骤列表中查找指定ID的步骤
// 支持直接匹配和递归查找（在 ParallelTasks 中）
func (s *WorkflowDomainService) findStepInList(stepID string, steps []StepDefinition) *StepDefinition {
	for i := range steps {
		if steps[i].ID == stepID {
			return &steps[i]
		}

		// 递归查找：在并行任务中查找
		if steps[i].Type == "parallel" && len(steps[i].ParallelTasks) > 0 {
			if found := s.findStepInList(stepID, steps[i].ParallelTasks); found != nil {
				return found
			}
		}
	}
	return nil
}

// extractParallelTaskInfo 从 taskKey 中提取并行任务信息
// 返回值：(parentStepID, index)
// 如果不是并行任务，返回 ("", 0)
func (s *WorkflowDomainService) extractParallelTaskInfo(taskKey string) (string, int) {
	if taskKey == "" {
		return "", 0
	}

	// 查找最后一个下划线
	idx := strings.LastIndex(taskKey, "_")
	if idx <= 0 || idx >= len(taskKey)-1 {
		// 没有下划线，或下划线在末尾，都不是有效的并行任务格式
		return "", 0
	}

	potentialSuffix := taskKey[idx+1:]
	index, err := strconv.Atoi(potentialSuffix)
	if err != nil {
		// 后缀不是数字，不是并行任务
		return "", 0
	}

	// 并行任务的序号从1开始，0是无效的
	if index <= 0 {
		log.Printf("[WorkflowDomainService] Invalid parallel task index: %d (must be > 0)", index)
		return "", 0
	}

	parentStepID := taskKey[:idx]
	return parentStepID, index
}

// findNextStep 查找下一个步骤
func (s *WorkflowDomainService) FindNextStep(currentStep *StepDefinition, definition *WorkflowDefinitionStruct, instance *instance_aggregate.WorkflowInstance) *StepDefinition {
	// 优先使用 next_steps 字段（支持条件分支）
	if len(currentStep.NextSteps) > 0 {
		// 遍历所有可能的下一步，找到第一个满足条件的
		for _, nextStepID := range currentStep.NextSteps {
			nextStep := s.FindStepByID(nextStepID, definition)
			if nextStep == nil {
				log.Printf("[EngineService] Next step not found: %s", nextStepID)
				continue
			}

			// 检查步骤条件
			if nextStep.Condition != "" {
				if !s.EvaluateCondition(nextStep.Condition, instance) {
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
		if !s.EvaluateCondition(nextStep.Condition, instance) {
			log.Printf("[EngineService] Step condition not met: %s, skipping", nextStep.Condition)
			// 条件不满足，继续查找下一个步骤
			return s.FindNextStep(nextStep, definition, instance)
		}
	}

	return nextStep
}

// findStepByID 根据ID查找步骤
func (s *WorkflowDomainService) FindStepByID(stepID string, definition *WorkflowDefinitionStruct) *StepDefinition {
	for i := range definition.Steps {
		if definition.Steps[i].ID == stepID {
			return &definition.Steps[i]
		}
	}
	return nil
}

// evaluateCondition 评估条件表达式
func (s *WorkflowDomainService) EvaluateCondition(condition string, instance *instance_aggregate.WorkflowInstance) bool {
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
func (s *WorkflowDomainService) resolveVariable(value string, instance *instance_aggregate.WorkflowInstance) string {
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
