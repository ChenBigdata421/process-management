package router

import (
	common "jxt-evidence-system/process-management/shared/common/middleware"

	jwt "github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth"
	"github.com/gin-gonic/gin"
)

var (
	routerNoCheckRole = make([]func(*gin.RouterGroup), 0)
	routerCheckRole   = make([]func(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware), 0)
)

// initRouter 路由示例
func initRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) *gin.Engine {
	// 注册所有路由依赖
	//RegisterDependencies()

	println("🔧 [DEBUG] initRouter() 被调用")
	println("🔧 [DEBUG] routerNoCheckRole 数量:", len(routerNoCheckRole))
	println("🔧 [DEBUG] routerCheckRole 数量:", len(routerCheckRole))

	// 无需认证的路由
	noCheckRoleRouter(r)
	// 需要认证的路由
	checkRoleRouter(r, authMiddleware)

	return r
}

// noCheckRoleRouter 无需认证的路由示例
func noCheckRoleRouter(r *gin.Engine) {
	// 可根据业务需求来设置接口版本
	v1 := r.Group("/api/v1")
	v1.Use(common.TenantResolver)

	for _, f := range routerNoCheckRole {
		f(v1)
	}
}

// checkRoleRouter 需要认证的路由示例
func checkRoleRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	// 可根据业务需求来设置接口版本
	v1 := r.Group("/api/v1")
	v1.Use(common.TenantResolver)

	for _, f := range routerCheckRole {
		f(v1, authMiddleware)
	}
}
