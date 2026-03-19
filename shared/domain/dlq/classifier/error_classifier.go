package classifier

import (
	"strings"
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypePoison        ErrorType = "POISON"        // 毒消息：无法恢复的业务逻辑错误
	ErrorTypeUnrecoverable ErrorType = "UNRECOVERABLE" // 不可恢复：系统级错误（如配置错误）
	ErrorTypeRecoverable   ErrorType = "RECOVERABLE"   // 可恢复：临时性错误（如网络超时）
)

// ErrorClassifier 错误分类器接口
type ErrorClassifier interface {
	// Classify 分类错误
	Classify(err error) ErrorType

	// IsPoison 判断是否为毒消息
	IsPoison(err error) bool

	// IsRecoverable 判断是否可恢复
	IsRecoverable(err error) bool
}

// DefaultErrorClassifier 默认错误分类器实现
type DefaultErrorClassifier struct{}

// NewDefaultErrorClassifier 创建默认错误分类器
func NewDefaultErrorClassifier() ErrorClassifier {
	return &DefaultErrorClassifier{}
}

// Classify 分类错误
func (c *DefaultErrorClassifier) Classify(err error) ErrorType {
	if err == nil {
		return ErrorTypeRecoverable
	}

	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)

	// 1. 毒消息判断（业务逻辑错误，重试无意义）
	poisonPatterns := []string{
		"payload 为空",
		"payload 解析失败",
		"无效的",
		"不合法",
		"格式错误",
		"参数错误",
		"违反约束",
		"重复",
		"已存在",
		"不存在",
		"权限不足",
		"未授权",
		"forbidden",
		"invalid",
		"illegal",
		"malformed",
		"constraint violation",
		"duplicate",
		"not found",
		"unauthorized",
	}

	for _, pattern := range poisonPatterns {
		if strings.Contains(errMsgLower, strings.ToLower(pattern)) {
			return ErrorTypePoison
		}
	}

	// 2. 不可恢复错误判断（系统配置错误，需要人工介入）
	unrecoverablePatterns := []string{
		"配置错误",
		"初始化失败",
		"连接池耗尽",
		"数据库不可用",
		"服务不可用",
		"panic",
		"fatal",
		"configuration error",
		"initialization failed",
		"service unavailable",
	}

	for _, pattern := range unrecoverablePatterns {
		if strings.Contains(errMsgLower, strings.ToLower(pattern)) {
			return ErrorTypeUnrecoverable
		}
	}

	// 3. 可恢复错误判断（临时性错误，可以重试）
	recoverablePatterns := []string{
		"超时",
		"timeout",
		"连接被拒绝",
		"connection refused",
		"连接重置",
		"connection reset",
		"临时",
		"temporary",
		"重试",
		"retry",
		"网络",
		"network",
		"i/o timeout",
		"deadline exceeded",
		"context deadline exceeded",
		"context canceled",
	}

	for _, pattern := range recoverablePatterns {
		if strings.Contains(errMsgLower, strings.ToLower(pattern)) {
			return ErrorTypeRecoverable
		}
	}

	// 4. 默认判断：未知错误视为可恢复（保守策略）
	return ErrorTypeRecoverable
}

// IsPoison 判断是否为毒消息
func (c *DefaultErrorClassifier) IsPoison(err error) bool {
	return c.Classify(err) == ErrorTypePoison
}

// IsRecoverable 判断是否可恢复
func (c *DefaultErrorClassifier) IsRecoverable(err error) bool {
	return c.Classify(err) == ErrorTypeRecoverable
}
