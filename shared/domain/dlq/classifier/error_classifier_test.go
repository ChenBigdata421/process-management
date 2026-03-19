package classifier

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClassify_PoisonMessage 测试毒消息错误分类
func TestClassify_PoisonMessage(t *testing.T) {
	classifier := NewDefaultErrorClassifier()

	testCases := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "Payload为空",
			err:      errors.New("payload 为空"),
			expected: ErrorTypePoison,
		},
		{
			name:     "Payload解析失败",
			err:      errors.New("payload 解析失败"),
			expected: ErrorTypePoison,
		},
		{
			name:     "无效的参数",
			err:      errors.New("无效的 event type"),
			expected: ErrorTypePoison,
		},
		{
			name:     "格式错误",
			err:      errors.New("格式错误: JSON 解析失败"),
			expected: ErrorTypePoison,
		},
		{
			name:     "参数错误",
			err:      errors.New("参数错误: 租户ID为空"),
			expected: ErrorTypePoison,
		},
		{
			name:     "重复记录",
			err:      errors.New("重复的事件ID"),
			expected: ErrorTypePoison,
		},
		{
			name:     "记录不存在",
			err:      errors.New("record not found"),
			expected: ErrorTypePoison,
		},
		{
			name:     "权限不足",
			err:      errors.New("权限不足: 无法访问资源"),
			expected: ErrorTypePoison,
		},
		{
			name:     "未授权访问",
			err:      errors.New("unauthorized access"),
			expected: ErrorTypePoison,
		},
		{
			name:     "违反约束",
			err:      errors.New("constraint violation: foreign key"),
			expected: ErrorTypePoison,
		},
		{
			name:     "非法值",
			err:      errors.New("illegal value for field"),
			expected: ErrorTypePoison,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.Classify(tc.err)
			assert.Equal(t, tc.expected, result, "错误应被分类为毒消息")
		})
	}
}

// TestClassify_UnrecoverableError 测试不可恢复错误分类
func TestClassify_UnrecoverableError(t *testing.T) {
	classifier := NewDefaultErrorClassifier()

	testCases := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "配置错误",
			err:      errors.New("配置错误: 数据库连接字符串为空"),
			expected: ErrorTypeUnrecoverable,
		},
		{
			name:     "初始化失败",
			err:      errors.New("初始化失败: 无法加载配置文件"),
			expected: ErrorTypeUnrecoverable,
		},
		{
			name:     "连接池耗尽",
			err:      errors.New("连接池耗尽: 无法获取数据库连接"),
			expected: ErrorTypeUnrecoverable,
		},
		{
			name:     "数据库不可用",
			err:      errors.New("数据库不可用"),
			expected: ErrorTypeUnrecoverable,
		},
		{
			name:     "服务不可用",
			err:      errors.New("service unavailable"),
			expected: ErrorTypeUnrecoverable,
		},
		{
			name:     "Panic错误",
			err:      errors.New("panic: runtime error"),
			expected: ErrorTypeUnrecoverable,
		},
		{
			name:     "Fatal错误",
			err:      errors.New("fatal: system crash"),
			expected: ErrorTypeUnrecoverable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.Classify(tc.err)
			assert.Equal(t, tc.expected, result, "错误应被分类为不可恢复")
		})
	}
}

// TestClassify_RecoverableError 测试可恢复错误分类
func TestClassify_RecoverableError(t *testing.T) {
	classifier := NewDefaultErrorClassifier()

	testCases := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "连接超时",
			err:      errors.New("connection timeout"),
			expected: ErrorTypeRecoverable,
		},
		{
			name:     "数据库连接失败",
			err:      errors.New("dial tcp: connection refused"),
			expected: ErrorTypeRecoverable,
		},
		{
			name:     "网络错误",
			err:      errors.New("network is unreachable"),
			expected: ErrorTypeRecoverable,
		},
		{
			name:     "死锁",
			err:      errors.New("Deadlock found when trying to get lock"),
			expected: ErrorTypeRecoverable,
		},
		{
			name:     "连接池耗尽",
			err:      errors.New("too many connections"),
			expected: ErrorTypeRecoverable,
		},
		{
			name:     "临时IO错误",
			err:      errors.New("i/o timeout"),
			expected: ErrorTypeRecoverable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.Classify(tc.err)
			assert.Equal(t, tc.expected, result, "错误应被分类为可恢复")
		})
	}
}

// TestClassify_UnknownError 测试未知错误的默认分类（保守策略）
func TestClassify_UnknownError(t *testing.T) {
	classifier := NewDefaultErrorClassifier()

	testCases := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "完全未知的错误",
			err:      errors.New("some unknown error"),
			expected: ErrorTypeRecoverable, // 默认为可恢复，保守策略
		},
		{
			name:     "自定义业务错误",
			err:      errors.New("business logic error"),
			expected: ErrorTypeRecoverable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.Classify(tc.err)
			assert.Equal(t, tc.expected, result, "未知错误应默认为可恢复（保守策略）")
		})
	}
}

// TestIsPoison 测试 IsPoison 方法
func TestIsPoison(t *testing.T) {
	classifier := NewDefaultErrorClassifier()

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "毒消息错误",
			err:      errors.New("invalid parameter"),
			expected: true,
		},
		{
			name:     "不可恢复错误",
			err:      errors.New("fatal error"),
			expected: false,
		},
		{
			name:     "可恢复错误",
			err:      errors.New("connection timeout"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.IsPoison(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestIsRecoverable 测试 IsRecoverable 方法
func TestIsRecoverable(t *testing.T) {
	classifier := NewDefaultErrorClassifier()

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "可恢复错误",
			err:      errors.New("connection timeout"),
			expected: true,
		},
		{
			name:     "毒消息错误",
			err:      errors.New("invalid parameter"),
			expected: false,
		},
		{
			name:     "不可恢复错误",
			err:      errors.New("fatal error"),
			expected: false,
		},
		{
			name:     "未知错误（默认可恢复）",
			err:      errors.New("unknown error"),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.IsRecoverable(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestClassify_NilError 测试 nil 错误的处理
func TestClassify_NilError(t *testing.T) {
	classifier := NewDefaultErrorClassifier()

	// nil 错误应该返回可恢复（虽然实际不应该传 nil）
	result := classifier.Classify(nil)
	assert.Equal(t, ErrorTypeRecoverable, result)
}

// TestClassify_CaseSensitivity 测试错误消息大小写不敏感
func TestClassify_CaseSensitivity(t *testing.T) {
	classifier := NewDefaultErrorClassifier()

	testCases := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "大写 INVALID",
			err:      errors.New("INVALID PARAMETER"),
			expected: ErrorTypePoison,
		},
		{
			name:     "小写 invalid",
			err:      errors.New("invalid parameter"),
			expected: ErrorTypePoison,
		},
		{
			name:     "混合大小写 Invalid",
			err:      errors.New("Invalid Parameter"),
			expected: ErrorTypePoison,
		},
		{
			name:     "大写 TIMEOUT",
			err:      errors.New("CONNECTION TIMEOUT"),
			expected: ErrorTypeRecoverable,
		},
		{
			name:     "小写 timeout",
			err:      errors.New("connection timeout"),
			expected: ErrorTypeRecoverable,
		},
		{
			name:     "大写 FATAL",
			err:      errors.New("FATAL ERROR"),
			expected: ErrorTypeUnrecoverable,
		},
		{
			name:     "小写 fatal",
			err:      errors.New("fatal error"),
			expected: ErrorTypeUnrecoverable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.Classify(tc.err)
			assert.Equal(t, tc.expected, result, "错误分类应该大小写不敏感")
		})
	}
}

// BenchmarkClassify 性能测试：验证分类延迟 < 1ms
func BenchmarkClassify(b *testing.B) {
	classifier := NewDefaultErrorClassifier()
	err := errors.New("connection timeout")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifier.Classify(err)
	}
}

// BenchmarkClassify_PoisonMessage 毒消息分类性能测试
func BenchmarkClassify_PoisonMessage(b *testing.B) {
	classifier := NewDefaultErrorClassifier()
	err := errors.New("failed to unmarshal event")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifier.Classify(err)
	}
}

// BenchmarkClassify_UnrecoverableError 不可恢复错误分类性能测试
func BenchmarkClassify_UnrecoverableError(b *testing.B) {
	classifier := NewDefaultErrorClassifier()
	err := errors.New("Data too long for column 'media_name'")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifier.Classify(err)
	}
}
