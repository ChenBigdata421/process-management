package middleware

import (
	tenantmiddleware "github.com/ChenBigdata421/jxt-core/sdk/pkg/tenant/middleware"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/tenant/provider"
	"github.com/gin-gonic/gin"
)

// DomainFallbackMiddleware creates a middleware that performs domain-based
// tenant identification as a fallback when Header-based identification fails.
//
// Behavior:
//   - If tenant_id already exists in context: skip, call c.Next()
//   - If provider is nil: delegate to jxt-core's ExtractTenantID (subdomain extraction only)
//   - Otherwise: delegate to jxt-core's ExtractTenantID with domain lookup
//
// The middleware wraps jxt-core's ExtractTenantID with skip logic to implement
// the "Header priority + Domain fallback" two-stage identification strategy.
func DomainFallbackMiddleware(p *provider.Provider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Skip if tenant_id already set by Header middleware
		if _, exists := c.Get("tenant_id"); exists {
			c.Next()
			return
		}

		// 2. Delegate to jxt-core's domain identification (Abort mode)
		// Note: Only pass domain lookup if provider is non-nil to avoid panic
		tenantmiddleware.ExtractTenantID(
			tenantmiddleware.WithProviderConfig(p), // 从 ETCD 读取配置
		)(c)
	}
}
