package middleware

import (
	"errors"
	"jxt-evidence-system/process-management/shared/common/global"
	"net/http"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	"github.com/casbin/casbin/v2/util"
	"go.uber.org/zap"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthCheckRole 权限检查中间件
func AuthCheckRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetRequestLogger(c)
		// 从上下文中获取tenantID
		ctx := c.Request.Context()
		tenantID, ok := ctx.Value(global.TenantIDKey).(string)
		if !ok {
			log.Error("tenant id not exist")
			response.Error(c, 500, errors.New("tenant id not exist"), "")
			return
		}
		data, _ := c.Get(jwtauth.JwtPayloadKey)
		v := data.(jwtauth.MapClaims)
		e := sdk.Runtime.GetTenantCasbin(tenantID)
		var res, casbinExclude bool
		var err error
		//检查权限
		if v["rolekey"] == "admin" {
			res = true
			c.Next()
			return
		}
		for _, i := range CasbinExclude {
			if util.KeyMatch2(c.Request.URL.Path, i.Url) && c.Request.Method == i.Method {
				casbinExclude = true
				break
			}
		}
		if casbinExclude {
			log.Info("Casbin exclusion, no validation", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
			c.Next()
			return
		}
		res, err = e.Enforce(v["rolekey"], c.Request.URL.Path, c.Request.Method)
		if err != nil {
			log.Error("AuthCheckRole error", zap.Error(err), zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
			response.Error(c, 500, err, "")
			return
		}

		if res {
			log.Info("Request details", zap.Bool("isTrue", res), zap.String("role", v["rolekey"].(string)), zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
			c.Next()
		} else {
			log.Warn("当前request无权限，请管理员确认！", zap.Bool("isTrue", res), zap.String("role", v["rolekey"].(string)), zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusOK, gin.H{
				"code": 403,
				"msg":  "对不起，您没有该接口访问权限，请联系管理员",
			})
			c.Abort()
			return
		}

	}
}
