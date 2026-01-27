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

// findPreviousCompletedTask 查找上一个已完成的任务
func (s *WorkflowDomainService) FindPreviousCompletedTask(tasks []*task_aggregate.Task, currentTaskID valueobject.TaskID) *task_aggregate.Task {
	for i := len(tasks) - 1; i >= 0; i-- {
		t := tasks[i]
		if t.TaskID != currentTaskID && t.Status == status.TaskStatusCompleted && t.CompletedAt != nil {
			return t
		}
	}
	return nil
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
