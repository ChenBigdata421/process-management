package workflow

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cschleiden/go-workflows/workflow"
)

// StepDefinition 步骤定义
type StepDefinition struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Timeout     int                    `json:"timeout"`
	Retries     int                    `json:"retries"`
	Params      map[string]interface{} `json:"params"`
}

// WorkflowDefinition 工作流定义
type WorkflowDefinition struct {
	Steps []StepDefinition `json:"steps"`
}

// DynamicWorkflowExecutor 动态工作流执行器
// 根据工作流定义动态执行工作流，无需修改后端代码
type DynamicWorkflowExecutor struct {
	activities map[string]interface{}
}

// NewDynamicWorkflowExecutor 创建动态工作流执行器
func NewDynamicWorkflowExecutor() *DynamicWorkflowExecutor {
	return &DynamicWorkflowExecutor{
		activities: make(map[string]interface{}),
	}
}

// RegisterActivity 注册活动
func (e *DynamicWorkflowExecutor) RegisterActivity(activityType string, activity interface{}) {
	e.activities[activityType] = activity
	log.Printf("Activity registered: %s", activityType)
}

// ExecuteWorkflowByDefinition 根据工作流定义执行工作流
// 这是一个通用的工作流执行函数，可以处理任何工作流定义
func (e *DynamicWorkflowExecutor) ExecuteWorkflowByDefinition(ctx workflow.Context, definitionJSON string, input string) (string, error) {
	// 解析工作流定义
	var definition WorkflowDefinition
	if err := json.Unmarshal([]byte(definitionJSON), &definition); err != nil {
		return "", fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	// 验证工作流定义
	if len(definition.Steps) == 0 {
		return "", fmt.Errorf("workflow definition has no steps")
	}

	// 执行工作流步骤
	currentInput := input
	for i, step := range definition.Steps {
		log.Printf("Executing step %d: %s (type: %s)", i+1, step.Name, step.Type)

		// 执行步骤
		result, err := e.executeStep(ctx, step, currentInput)
		if err != nil {
			return "", fmt.Errorf("step %s failed: %w", step.Name, err)
		}

		// 当前步骤的输出作为下一步骤的输入
		currentInput = result
		log.Printf("Step %s completed, output: %s", step.Name, result)
	}

	return currentInput, nil
}

// executeStep 执行单个步骤
func (e *DynamicWorkflowExecutor) executeStep(ctx workflow.Context, step StepDefinition, input string) (string, error) {
	// 根据步骤类型执行相应的活动
	switch step.Type {
	case "validate":
		return e.executeValidateStep(ctx, step, input)
	case "process":
		return e.executeProcessStep(ctx, step, input)
	case "notify":
		return e.executeNotifyStep(ctx, step, input)
	case "complete":
		return e.executeCompleteStep(ctx, step, input)
	case "custom":
		return e.executeCustomStep(ctx, step, input)
	default:
		return "", fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeValidateStep 执行验证步骤
func (e *DynamicWorkflowExecutor) executeValidateStep(ctx workflow.Context, step StepDefinition, input string) (string, error) {
	result, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		ValidateActivity,
		input,
	).Get(ctx)

	if err != nil {
		return "", fmt.Errorf("validate activity failed: %w", err)
	}

	return result, nil
}

// executeProcessStep 执行处理步骤
func (e *DynamicWorkflowExecutor) executeProcessStep(ctx workflow.Context, step StepDefinition, input string) (string, error) {
	result, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		ProcessActivity,
		input,
	).Get(ctx)

	if err != nil {
		return "", fmt.Errorf("process activity failed: %w", err)
	}

	return result, nil
}

// executeNotifyStep 执行通知步骤
func (e *DynamicWorkflowExecutor) executeNotifyStep(ctx workflow.Context, step StepDefinition, input string) (string, error) {
	result, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		NotifyActivity,
		input,
	).Get(ctx)

	if err != nil {
		return "", fmt.Errorf("notify activity failed: %w", err)
	}

	return result, nil
}

// executeCompleteStep 执行完成步骤
func (e *DynamicWorkflowExecutor) executeCompleteStep(ctx workflow.Context, step StepDefinition, input string) (string, error) {
	result, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		CompleteActivity,
		input,
	).Get(ctx)

	if err != nil {
		return "", fmt.Errorf("complete activity failed: %w", err)
	}

	return result, nil
}

// executeCustomStep 执行自定义步骤
func (e *DynamicWorkflowExecutor) executeCustomStep(ctx workflow.Context, step StepDefinition, input string) (string, error) {
	// 自定义步骤可以根据参数执行不同的逻辑
	// 这里作为示例，直接返回输入
	log.Printf("Custom step: %s, params: %v", step.Name, step.Params)
	return input, nil
}

// DynamicWorkflow 动态工作流函数
// 这个函数可以处理任何工作流定义
func DynamicWorkflow(ctx workflow.Context, definitionJSON string, input string) error {
	executor := NewDynamicWorkflowExecutor()

	// 注册所有活动
	executor.RegisterActivity("validate", ValidateActivity)
	executor.RegisterActivity("process", ProcessActivity)
	executor.RegisterActivity("notify", NotifyActivity)
	executor.RegisterActivity("complete", CompleteActivity)

	// 执行工作流
	result, err := executor.ExecuteWorkflowByDefinition(ctx, definitionJSON, input)
	if err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	log.Printf("Workflow completed successfully, result: %s", result)
	return nil
}
