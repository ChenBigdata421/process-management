package global

// 租户ID上下文键
type contextKey string
type userId string

const (
	TenantIDKey contextKey = "JXT-Tenant"
	UserIDKey   contextKey = "JXT-User"
)
