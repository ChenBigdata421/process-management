package infrastructure_service

import "time"

// 统一配置：适度重试 + 快速降级
const (
	MaxRetries   = 2                      // 最大重试2次，平衡可靠性和性能
	RetryBackoff = 500 * time.Millisecond // 固定500ms间隔，简单有效
)
