package actions

import (
	"jxt-evidence-system/process-management/shared/common/models"
	"jxt-evidence-system/process-management/shared/common/query"
	"net/http"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreateAction 通用新增动作
func CreateAction(control query.Control) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetRequestLogger(c)
		db, err := pkg.GetOrm(c)
		if err != nil {
			log.Error("get db failure", zap.Error(err))
			return
		}

		//新增操作
		req := control.Generate()
		err = req.Bind(c)
		if err != nil {
			response.Error(c, http.StatusUnprocessableEntity, err, err.Error())
			return
		}
		var object models.ActiveRecord
		object, err = req.GenerateM()
		if err != nil {
			response.Error(c, 500, err, "模型生成失败")
			return
		}
		object.SetCreateBy(user.GetUserId(c))
		err = db.WithContext(c).Create(object).Error
		if err != nil {
			log.Error("Create error", zap.Error(err))
			response.Error(c, 500, err, "创建失败")
			return
		}
		response.OK(c, object.GetId(), "创建成功")
		c.Next()
	}
}
