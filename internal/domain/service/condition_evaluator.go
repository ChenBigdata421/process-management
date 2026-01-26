package domain_service

import (
	"context"
	"encoding/json"
	"fmt"
	instance_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/instance"
	task_aggregate "jxt-evidence-system/process-management/internal/domain/aggregate/task"
	task_repository "jxt-evidence-system/process-management/internal/domain/aggregate/task/repository"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// ConditionEvaluator 条件表达式求值器
type ConditionEvaluator struct {
	instance *instance_aggregate.WorkflowInstance
	taskRepo task_repository.TaskRepository
}

// NewConditionEvaluator 创建条件求值器
func NewConditionEvaluator(instance *instance_aggregate.WorkflowInstance, taskRepo task_repository.TaskRepository) *ConditionEvaluator {
	return &ConditionEvaluator{
		instance: instance,
		taskRepo: taskRepo,
	}
}

// Evaluate 求值条件表达式
// 支持的表达式格式：
// - ${variable} == "value"
// - ${step_id.field} == true
// - ${step_id.field} > 100
// - ${step_id.field} != null
// - 逻辑运算：&&, ||, !
func (e *ConditionEvaluator) Evaluate(condition string) (bool, error) {
	if condition == "" {
		return true, nil
	}

	log.Printf("[ConditionEvaluator] Evaluating condition: %s", condition)

	// 处理逻辑运算符
	if strings.Contains(condition, "&&") {
		parts := strings.Split(condition, "&&")
		for _, part := range parts {
			result, err := e.Evaluate(strings.TrimSpace(part))
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil
	}

	if strings.Contains(condition, "||") {
		parts := strings.Split(condition, "||")
		for _, part := range parts {
			result, err := e.Evaluate(strings.TrimSpace(part))
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	// 处理取反
	if strings.HasPrefix(strings.TrimSpace(condition), "!") {
		innerCondition := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(condition), "!"))
		result, err := e.Evaluate(innerCondition)
		if err != nil {
			return false, err
		}
		return !result, nil
	}

	// 处理比较运算符
	return e.evaluateComparison(condition)
}

// evaluateComparison 求值比较表达式
func (e *ConditionEvaluator) evaluateComparison(condition string) (bool, error) {
	// 支持的运算符：==, !=, >, <, >=, <=
	operators := []string{"==", "!=", ">=", "<=", ">", "<"}

	for _, op := range operators {
		if strings.Contains(condition, op) {
			parts := strings.SplitN(condition, op, 2)
			if len(parts) != 2 {
				continue
			}

			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// 解析左值
			leftValue, err := e.resolveValue(left)
			if err != nil {
				return false, err
			}

			// 解析右值
			rightValue, err := e.resolveValue(right)
			if err != nil {
				return false, err
			}

			// 执行比较
			return e.compare(leftValue, rightValue, op)
		}
	}

	// 如果没有运算符，尝试作为布尔值解析
	value, err := e.resolveValue(condition)
	if err != nil {
		return false, err
	}

	// 转换为布尔值
	return e.toBool(value), nil
}

// resolveValue 解析值
// 支持：
// - ${variable} - 从实例输入中获取
// - ${step_id.field} - 从步骤输出中获取
// - "string" - 字符串字面量
// - 123 - 数字字面量
// - true/false - 布尔字面量
// - null - 空值
func (e *ConditionEvaluator) resolveValue(value string) (interface{}, error) {
	value = strings.TrimSpace(value)

	// 处理字符串字面量
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\""), nil
	}

	// 处理布尔字面量
	if value == "true" {
		return true, nil
	}
	if value == "false" {
		return false, nil
	}

	// 处理null
	if value == "null" {
		return nil, nil
	}

	// 处理数字字面量
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num, nil
	}

	// 处理变量引用 ${...}
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		varPath := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")

		// 检查是否是步骤输出引用 step_id.field
		if strings.Contains(varPath, ".") {
			return e.resolveStepOutput(varPath)
		}

		// 从实例输入中获取
		return e.resolveInstanceInput(varPath)
	}

	// 默认作为字符串返回
	return value, nil
}

// resolveInstanceInput 从实例输入中解析变量
func (e *ConditionEvaluator) resolveInstanceInput(varName string) (interface{}, error) {
	var input map[string]interface{}
	inputData := e.instance.Input

	// 检查双重编码
	inputStr := string(inputData)
	if strings.HasPrefix(inputStr, "\"") && strings.HasSuffix(inputStr, "\"") {
		var tempStr string
		if err := json.Unmarshal(inputData, &tempStr); err == nil {
			inputData = []byte(tempStr)
		}
	}

	if err := json.Unmarshal(inputData, &input); err != nil {
		return nil, fmt.Errorf("failed to parse instance input: %v", err)
	}

	if val, ok := input[varName]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("variable not found in instance input: %s", varName)
}

// resolveStepOutput 从步骤输出中解析变量
// 格式：step_id.field
func (e *ConditionEvaluator) resolveStepOutput(varPath string) (interface{}, error) {
	parts := strings.SplitN(varPath, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid step output reference: %s", varPath)
	}

	stepKey := parts[0]
	fieldName := parts[1]

	ctx := context.Background()

	tasks, err := e.taskRepo.FindByInstanceID(ctx, e.instance.InstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %v", err)
	}

	var targetTask *task_aggregate.Task
	for _, task := range tasks {
		if task.TaskKey == stepKey {
			targetTask = task
			break
		}
	}

	if targetTask == nil {
		return nil, fmt.Errorf("task not found for step: %s", stepKey)
	}

	// 从任务输出中获取字段值
	if len(targetTask.Output) == 0 {
		return nil, fmt.Errorf("task output is empty for step: %s", stepKey)
	}

	var output map[string]interface{}
	if err := json.Unmarshal(targetTask.Output, &output); err != nil {
		return nil, fmt.Errorf("failed to parse task output: %v", err)
	}

	if val, ok := output[fieldName]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("field not found in task output: %s.%s", stepKey, fieldName)
}

// compare 比较两个值
func (e *ConditionEvaluator) compare(left, right interface{}, operator string) (bool, error) {
	switch operator {
	case "==":
		return e.equals(left, right), nil
	case "!=":
		return !e.equals(left, right), nil
	case ">":
		return e.greaterThan(left, right)
	case "<":
		return e.lessThan(left, right)
	case ">=":
		gt, err := e.greaterThan(left, right)
		if err != nil {
			return false, err
		}
		return gt || e.equals(left, right), nil
	case "<=":
		lt, err := e.lessThan(left, right)
		if err != nil {
			return false, err
		}
		return lt || e.equals(left, right), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// equals 判断相等
func (e *ConditionEvaluator) equals(left, right interface{}) bool {
	// 处理 nil
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	// 类型转换
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	return leftStr == rightStr
}

// greaterThan 判断大于
func (e *ConditionEvaluator) greaterThan(left, right interface{}) (bool, error) {
	leftNum, err := e.toFloat(left)
	if err != nil {
		return false, err
	}

	rightNum, err := e.toFloat(right)
	if err != nil {
		return false, err
	}

	return leftNum > rightNum, nil
}

// lessThan 判断小于
func (e *ConditionEvaluator) lessThan(left, right interface{}) (bool, error) {
	leftNum, err := e.toFloat(left)
	if err != nil {
		return false, err
	}

	rightNum, err := e.toFloat(right)
	if err != nil {
		return false, err
	}

	return leftNum < rightNum, nil
}

// toFloat 转换为浮点数
func (e *ConditionEvaluator) toFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert to float: %v", value)
	}
}

// toBool 转换为布尔值
func (e *ConditionEvaluator) toBool(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "false" && v != "0"
	case float64:
		return v != 0
	case int:
		return v != 0
	default:
		return true
	}
}

// ValidateCondition 验证条件表达式语法
func ValidateCondition(condition string) error {
	if condition == "" {
		return nil
	}

	// 基本语法检查
	// 检查括号匹配
	openCount := strings.Count(condition, "(")
	closeCount := strings.Count(condition, ")")
	if openCount != closeCount {
		return fmt.Errorf("unmatched parentheses in condition: %s", condition)
	}
	/*匹配格式：${variable_name}
	允许的字符：字母、数字、下划线、点号
	至少一个字符*/
	// 检查变量引用格式
	varPattern := regexp.MustCompile(`\$\{[a-zA-Z0-9_\.]+\}`)
	matches := varPattern.FindAllString(condition, -1) //在字符串中查找所有匹配正则表达式的子串
	for _, match := range matches {
		varPath := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		if strings.Contains(varPath, "..") {
			return fmt.Errorf("invalid variable reference: %s", match)
		}
	}

	return nil
}
