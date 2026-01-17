package actions

import (
	"errors"

	"github.com/ChenBigdata421/jxt-core/sdk/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ChenBigdata421/jxt-core/sdk/config"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/jwtauth/user"
	"github.com/ChenBigdata421/jxt-core/sdk/pkg/response"
	"gorm.io/gorm"
)

type DataPermission struct {
	DataScope string
	UserId    int
	OrgId     int
	RoleId    int
}

func PermissionAction() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从JWT Token获取权限信息（新方案）
		if p := getDataPermissionFromToken(c); p != nil {
			c.Set(PermissionKey, p)
			c.Next()
			return
		}

		// 降级到数据库查询（兼容性方案）
		log := logger.GetRequestLogger(c)
		db, err := pkg.GetOrm(c)
		if err != nil {
			log.Error("get db failure", zap.Error(err))
			return
		}

		msgID := pkg.GenerateMsgIDFromContext(c)
		var p = new(DataPermission)
		if userId := user.GetUserIdStr(c); userId != "" {
			p, err = newDataPermission(db, userId)
			if err != nil {
				log.Error("PermissionAction error", zap.Error(err), zap.String("msgID", msgID))
				response.Error(c, 500, err, "权限范围鉴定错误")
				c.Abort()
				return
			}
		}
		c.Set(PermissionKey, p)
		c.Next()
	}
}

func newDataPermission(tx *gorm.DB, userId interface{}) (*DataPermission, error) {
	var err error
	p := &DataPermission{}

	err = tx.Table("sys_user").
		Select("sys_user.user_id", "sys_role.role_id", "sys_user.dept_id", "sys_role.data_scope").
		Joins("left join sys_role on sys_role.role_id = sys_user.role_id").
		Where("sys_user.user_id = ?", userId).
		Scan(p).Error
	if err != nil {
		err = errors.New("获取用户数据出错 msg:" + err.Error())
		return nil, err
	}
	return p, nil
}

func Permission(tableName string, p *DataPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !config.ApplicationConfig.EnableDP {
			return db
		}
		switch p.DataScope {
		case "2":
			return db.Where(tableName+".create_by in (select sys_user.user_id from sys_role_dept left join sys_user on sys_user.dept_id=sys_role_dept.dept_id where sys_role_dept.role_id = ?)", p.RoleId)
		case "3":
			return db.Where(tableName+".create_by in (SELECT user_id from sys_user where dept_id = ? )", p.OrgId)
		case "4":
			return db.Where(tableName+".create_by in (SELECT user_id from sys_user where sys_user.dept_id in(select dept_id from sys_dept where dept_path like ? ))", "%/"+pkg.IntToString(p.OrgId)+"/%")
		case "5":
			return db.Where(tableName+".create_by = ?", p.UserId)
		default:
			return db
		}
	}
}

func getPermissionFromContext(c *gin.Context) *DataPermission {
	p := new(DataPermission)
	if pm, ok := c.Get(PermissionKey); ok {
		switch pm.(type) {
		case *DataPermission:
			p = pm.(*DataPermission)
		}
	}
	return p
}

// GetPermissionFromContext 提供非action写法数据范围约束
func GetPermissionFromContext(c *gin.Context) *DataPermission {
	return getPermissionFromContext(c)
}

// getDataPermissionFromToken 从JWT Token中获取数据权限信息
func getDataPermissionFromToken(c *gin.Context) *DataPermission {
	// 获取JWT Token中的claims
	data, exists := c.Get("JWT_PAYLOAD")
	if !exists {
		return nil
	}

	claims, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}

	// 检查是否包含org_id字段
	orgIdFloat, ok := claims["org_id"].(float64)
	if !ok {
		return nil
	}

	// 构建DataPermission对象
	p := &DataPermission{
		UserId:    int(claims["identity"].(float64)),
		RoleId:    int(claims["roleid"].(float64)),
		OrgId:     int(orgIdFloat),
		DataScope: claims["datascope"].(string),
	}

	return p
}
