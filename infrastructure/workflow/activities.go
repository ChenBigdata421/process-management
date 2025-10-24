package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ProcessData 处理数据结构
type ProcessData struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Status string `json:"status"`
}

// ValidateActivity 验证活动
func ValidateActivity(ctx context.Context, input string) (string, error) {
	log.Printf("ValidateActivity: input=%s", input)

	if input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	data := &ProcessData{
		Input:  input,
		Output: fmt.Sprintf("Validated: %s", input),
		Status: "validated",
	}

	result, _ := json.Marshal(data)
	return string(result), nil
}

// ProcessActivity 处理活动
func ProcessActivity(ctx context.Context, input string) (string, error) {
	log.Printf("ProcessActivity: input=%s", input)

	var data ProcessData
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return "", err
	}

	data.Output = fmt.Sprintf("Processed: %s", data.Input)
	data.Status = "processed"

	result, _ := json.Marshal(data)
	return string(result), nil
}

// NotifyActivity 通知活动
func NotifyActivity(ctx context.Context, input string) (string, error) {
	log.Printf("NotifyActivity: input=%s", input)

	var data ProcessData
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return "", err
	}

	data.Output = fmt.Sprintf("Notified: %s", data.Output)
	data.Status = "notified"

	result, _ := json.Marshal(data)
	return string(result), nil
}

// CompleteActivity 完成活动
func CompleteActivity(ctx context.Context, input string) (string, error) {
	log.Printf("CompleteActivity: input=%s", input)

	var data ProcessData
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return "", err
	}

	data.Status = "completed"

	result, _ := json.Marshal(data)
	return string(result), nil
}

// ========== 订单处理活动 ==========

// OrderData 订单数据结构
type OrderData struct {
	OrderID       string  `json:"order_id"`
	CustomerID    string  `json:"customer_id"`
	TotalAmount   float64 `json:"total_amount"`
	CustomerEmail string  `json:"customer_email"`
	CustomerPhone string  `json:"customer_phone"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
}

// OrderValidateActivity 订单验证活动
func OrderValidateActivity(ctx context.Context, input string) (string, error) {
	log.Printf("OrderValidateActivity: input=%s", input)

	var order OrderData
	if err := json.Unmarshal([]byte(input), &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	// 验证订单数据
	if order.OrderID == "" {
		return "", fmt.Errorf("order_id cannot be empty")
	}
	if order.CustomerID == "" {
		return "", fmt.Errorf("customer_id cannot be empty")
	}
	if order.TotalAmount <= 0 {
		return "", fmt.Errorf("total_amount must be greater than 0")
	}
	if order.CustomerEmail == "" {
		return "", fmt.Errorf("customer_email cannot be empty")
	}

	// 设置状态
	order.Status = "validated"
	order.Message = fmt.Sprintf("订单 %s 验证成功", order.OrderID)

	log.Printf("OrderValidateActivity: order validated - %s", order.OrderID)

	result, _ := json.Marshal(order)
	return string(result), nil
}

// OrderProcessActivity 订单处理活动
func OrderProcessActivity(ctx context.Context, input string) (string, error) {
	log.Printf("OrderProcessActivity: input=%s", input)

	var order OrderData
	if err := json.Unmarshal([]byte(input), &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	// 处理订单
	// 在实际应用中，这里会调用数据库、支付系统等
	order.Status = "processing"
	order.Message = fmt.Sprintf("订单 %s 正在处理，金额: %.2f", order.OrderID, order.TotalAmount)

	log.Printf("OrderProcessActivity: order processing - %s, amount: %.2f", order.OrderID, order.TotalAmount)

	result, _ := json.Marshal(order)
	return string(result), nil
}

// OrderNotifyActivity 订单通知活动
func OrderNotifyActivity(ctx context.Context, input string) (string, error) {
	log.Printf("OrderNotifyActivity: input=%s", input)

	var order OrderData
	if err := json.Unmarshal([]byte(input), &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	// 发送通知
	// 在实际应用中，这里会调用邮件服务、短信服务等
	order.Status = "notified"
	order.Message = fmt.Sprintf("已向 %s 发送订单通知", order.CustomerEmail)

	log.Printf("OrderNotifyActivity: notification sent to %s (phone: %s)", order.CustomerEmail, order.CustomerPhone)

	result, _ := json.Marshal(order)
	return string(result), nil
}

// OrderCompleteActivity 订单完成活动
func OrderCompleteActivity(ctx context.Context, input string) (string, error) {
	log.Printf("OrderCompleteActivity: input=%s", input)

	var order OrderData
	if err := json.Unmarshal([]byte(input), &order); err != nil {
		return "", fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	// 完成订单处理
	order.Status = "completed"
	order.Message = fmt.Sprintf("订单 %s 处理完成", order.OrderID)

	log.Printf("OrderCompleteActivity: order completed - %s", order.OrderID)

	result, _ := json.Marshal(order)
	return string(result), nil
}

// RegisterActivities 注册所有活动到后端
// 注意：go-workflows 不使用 Registry 对象，而是直接在 worker 中注册
// 这个函数保留用于文档目的
func RegisterActivities() {
	log.Println("Activities registered successfully")
}
