package actions

import (
	"errors"
	"net/http"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"jxt-evidence-system/process-management/shared/common/models"
	query "jxt-evidence-system/process-management/shared/common/query"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"
)

// IndexAction 通用查询动作
func IndexAction(m models.ActiveRecord, d query.Index, f func() interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetRequestLogger(c)
		db, err := pkg.GetOrm(c)
		if err != nil {
			log.Error("get db failure", zap.Error(err))
			return
		}

		msgID := pkg.GenerateMsgIDFromContext(c)
		list := f()
		object := m.Generate()
		req := d.Generate()
		var count int64

		//查询列表
		err = req.Bind(c)
		if err != nil {
			response.Error(c, http.StatusUnprocessableEntity, err, "参数验证失败")
			return
		}

		//数据权限检查
		p := GetPermissionFromContext(c)

		err = db.WithContext(c).Model(object).
			Scopes(
				query.MakeCondition(req.GetNeedSearch(), db.Dialector.Name()),
				query.Paginate(req.GetPageSize(), req.GetPageIndex()),
				Permission(object.TableName(), p),
			).
			Find(list).Limit(-1).Offset(-1).
			Count(&count).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error("Index error", zap.Error(err), zap.String("msgID", msgID))
			response.Error(c, 500, err, "查询失败")
			return
		}
		response.PageOK(c, list, int(count), req.GetPageIndex(), req.GetPageSize(), "查询成功")
		c.Next()
	}
}
