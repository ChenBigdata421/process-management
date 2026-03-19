package middleware

import (
	"jxt-evidence-system/process-management/shared/common/actions"
	"jxt-evidence-system/process-management/shared/common/tenantdb"

	"github.com/ChenBigdata421/jxt-core/sdk"
	jwt "github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth"
	tenantmiddleware "github.com/ChenBigdata421/jxt-core/sdk/pkg/tenant/middleware"
	"github.com/gin-gonic/gin"
)

const (
	JwtTokenCheck   string = "JwtToken"
	RoleCheck       string = "AuthCheckRole"
	PermissionCheck string = "PermissionAction"
)

// InitMiddleware initializes the middleware chain for the application.
// tenantCache: tenant cache for domain-based identification fallback (can be nil)
func InitMiddleware(r *gin.Engine, tenantCache *tenantdb.Cache) {
	//r.Use(DemoEvn())

	// 1. Header-based tenant identification
	// If X-Tenant-ID is valid, sets tenant_id and continues
	// If missing/invalid, continues without setting tenant_id (allows fallback)
	r.Use(tenantmiddleware.ExtractTenantID(
		tenantmiddleware.WithResolverType("header"),
		tenantmiddleware.WithHeaderName("X-Tenant-ID"),
		tenantmiddleware.WithOnMissingTenant("Continue"),
	))

	// 2. Domain-based tenant identification (fallback)
	// Only executes if tenant_id not already set by Header middleware
	// If cache/provider is nil, middleware is not registered
	if tenantCache != nil && tenantCache.GetProvider() != nil {
		r.Use(DomainFallbackMiddleware(tenantCache.GetProvider()))
	}

	// 日志处理
	r.Use(LoggerToFile())
	// 自定义错误处理
	r.Use(CustomError)
	// NoCache is a middleware function that appends headers
	r.Use(NoCache)
	// 跨域处理
	r.Use(Options)
	// Secure is a middleware function that appends security
	r.Use(Secure)
	// 链路追踪
	//r.Use(middleware.Trace())

	sdk.Runtime.SetMiddleware(JwtTokenCheck, (*jwt.GinJWTMiddleware).MiddlewareFunc)
	sdk.Runtime.SetMiddleware(RoleCheck, AuthCheckRole())
	sdk.Runtime.SetMiddleware(PermissionCheck, actions.PermissionAction())
}
