package workflow

import (
	"fmt"
	"log"

	"github.com/cschleiden/go-workflows/workflow"
)

// ProcessWorkflow 业务流程工作流
// 这是一个完整的工作流定义，包含验证、处理、通知和完成四个活动
func ProcessWorkflow(ctx workflow.Context) error {
	// 第一步：验证输入
	validateResult, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		ValidateActivity,
		"test input",
	).Get(ctx)
	if err != nil {
		return fmt.Errorf("validate activity failed: %w", err)
	}
	log.Printf("Validate result: %s", validateResult)

	// 第二步：处理数据
	processResult, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		ProcessActivity,
		validateResult,
	).Get(ctx)
	if err != nil {
		return fmt.Errorf("process activity failed: %w", err)
	}
	log.Printf("Process result: %s", processResult)

	// 第三步：发送通知
	notifyResult, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		NotifyActivity,
		processResult,
	).Get(ctx)
	if err != nil {
		return fmt.Errorf("notify activity failed: %w", err)
	}
	log.Printf("Notify result: %s", notifyResult)

	// 第四步：完成处理
	completeResult, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		CompleteActivity,
		notifyResult,
	).Get(ctx)
	if err != nil {
		return fmt.Errorf("complete activity failed: %w", err)
	}
	log.Printf("Complete result: %s", completeResult)

	return nil
}

// OrderProcessingWorkflow 订单处理工作流
// 这是一个完整的订单处理工作流，包含验证、处理、通知和完成四个活动
func OrderProcessingWorkflow(ctx workflow.Context) error {
	// 第一步：验证订单
	validateResult, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		OrderValidateActivity,
		"", // 输入数据将由工作流实例提供
	).Get(ctx)
	if err != nil {
		return fmt.Errorf("order validate activity failed: %w", err)
	}
	log.Printf("Order validate result: %s", validateResult)

	// 第二步：处理订单
	processResult, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		OrderProcessActivity,
		validateResult,
	).Get(ctx)
	if err != nil {
		return fmt.Errorf("order process activity failed: %w", err)
	}
	log.Printf("Order process result: %s", processResult)

	// 第三步：发送通知
	notifyResult, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		OrderNotifyActivity,
		processResult,
	).Get(ctx)
	if err != nil {
		return fmt.Errorf("order notify activity failed: %w", err)
	}
	log.Printf("Order notify result: %s", notifyResult)

	// 第四步：完成订单处理
	completeResult, err := workflow.ExecuteActivity[string](
		ctx,
		workflow.DefaultActivityOptions,
		OrderCompleteActivity,
		notifyResult,
	).Get(ctx)
	if err != nil {
		return fmt.Errorf("order complete activity failed: %w", err)
	}
	log.Printf("Order complete result: %s", completeResult)

	return nil
}

// RegisterWorkflows 注册所有工作流
// 注意：go-workflows 不使用 Registry 对象，而是直接在 worker 中注册
// 这个函数保留用于文档目的
func RegisterWorkflows() {
	log.Println("Workflows registered successfully")
}
