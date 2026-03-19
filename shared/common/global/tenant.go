package global

import (
	tenantmiddleware "github.com/ChenBigdata421/jxt-core/sdk/pkg/tenant/middleware"
)

// TenantIDKey 是租户ID在context中的键
// 直接使用 jxt-core 导出的键，确保类型一致
var TenantIDKey = tenantmiddleware.TenantContextKey

// UserIDKey 用户ID上下文键（process-management 特殊需求）
type contextKey string

const (
	UserIDKey contextKey = "JXT-User"
)
