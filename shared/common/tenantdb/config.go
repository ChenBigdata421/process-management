// shared/common/tenantdb/config.go
package tenantdb

import "time"

// WatcherConfig 动态租户连接配置
type WatcherConfig struct {
	Enabled        bool                          `json:"enabled" yaml:"enabled"`
	MaxConcurrent  int                           `json:"max_concurrent" yaml:"max_concurrent"`
	RetryTimes     int                           `json:"retry_times" yaml:"retry_times"`
	RetryInterval  time.Duration                 `json:"retry_interval" yaml:"retry_interval"`
	ConnectTimeout time.Duration                 `json:"connect_timeout" yaml:"connect_timeout"`
	OnTenantAdded  func(tenantID int) error      // 新租户添加回调，用于自定义初始化逻辑（如 Casbin 策略加载）
}

// DefaultWatcherConfig 返回默认配置
func DefaultWatcherConfig() *WatcherConfig {
	return &WatcherConfig{
		Enabled:        true,
		MaxConcurrent:  5,
		RetryTimes:     3,
		RetryInterval:  5 * time.Second,
		ConnectTimeout: 30 * time.Second,
		OnTenantAdded:  nil,
	}
}
