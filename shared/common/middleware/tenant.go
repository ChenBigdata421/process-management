package middleware

import (
	"context"
	"net/http"

	"jxt-evidence-system/process-management/shared/common/global"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/gin-gonic/gin"
)

func TenantResolver(c *gin.Context) {

	// 如果没有启用多租户，tenantID设置为*，并直接返回
	if !config.TenantsConfig.Enabled {
		// 保留原来把主db存储在gin.Context供老代码auth，permission，job使用
		c.Set("db", sdk.Runtime.GetTenantDB("*").WithContext(c))
		// 记录tenantID到context中供后续使用
		ctx := context.WithValue(c.Request.Context(), global.TenantIDKey, "*")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		return
	} else if config.TenantsConfig.Resolver.Type == "host" {
		// 根据host解析TenantID，并把TenantID存储在通用context上下文中
		tenantID := sdk.Runtime.GetTenantID(c.Request.Host)
		if tenantID != "" {
			// 保留原来把主db存储在gin.Context供老代码auth，perssion，job使用
			c.Set("db", sdk.Runtime.GetTenantDB(tenantID).WithContext(c))
			// 记录tenantID到context中供后续使用
			ctx := context.WithValue(c.Request.Context(), global.TenantIDKey, tenantID)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 500,
				"msg":  "获取租户ID失败，请检查配置文件",
			})
			c.Abort()
			return
		}
	}
	c.Next()
}
