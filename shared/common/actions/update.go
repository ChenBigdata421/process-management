package actions

import (
	"net/http"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jxt-evidence-system/process-management/shared/common/models"
	"jxt-evidence-system/process-management/shared/common/query"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

// UpdateAction 通用更新动作
func UpdateAction(control query.Control) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetRequestLogger(c)
		db, err := pkg.GetOrm(c)
		if err != nil {
			log.Error("get db failure", zap.Error(err))
			return
		}

		msgID := pkg.GenerateMsgIDFromContext(c)
		req := control.Generate()
		//更新操作
		err = req.Bind(c)
		if err != nil {
			response.Error(c, http.StatusUnprocessableEntity, err, "参数验证失败")
			return
		}
		var object models.ActiveRecord
		object, err = req.GenerateM()
		if err != nil {
			response.Error(c, 500, err, "模型生成失败")
			return
		}
		object.SetUpdateBy(user.GetUserId(c))

		//数据权限检查
		p := GetPermissionFromContext(c)

		db = db.WithContext(c).Scopes(
			Permission(object.TableName(), p),
		).Where(req.GetId()).Updates(object)
		if err = db.Error; err != nil {
			log.Error("Update error", zap.Error(err), zap.String("msgID", msgID))
			response.Error(c, 500, err, "更新失败")
			return
		}
		if db.RowsAffected == 0 {
			response.Error(c, http.StatusForbidden, nil, "无权更新该数据")
			return
		}
		response.OK(c, object.GetId(), "更新成功")
		c.Next()
	}
}
