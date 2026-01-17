package global

import (
	"errors"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	"github.com/ChenBigdata421/jxt-core/sdk"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoadPolicy(c *gin.Context) (*casbin.SyncedEnforcer, error) {
	log := logger.GetRequestLogger(c)
	// 从上下文中获取tenantID
	ctx := c.Request.Context()
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	if !ok {
		err := errors.New("tenant id not exist")
		log.Error("casbin rbac_model or policy init error, ", zap.Error(err))
		return nil, err

	}
	if err := sdk.Runtime.GetTenantCasbin(tenantID).LoadPolicy(); err == nil {
		return sdk.Runtime.GetTenantCasbin(tenantID), err
	} else {
		log.Error("casbin rbac_model or policy init error, ", zap.Error(err))
		return nil, err
	}
}
