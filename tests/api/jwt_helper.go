package api_tests

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateTestToken 生成测试用 JWT token
// 参数: userID, roleID, username, roleName, orgID
// 返回: Bearer token 字符串
func GenerateTestToken(userID int, roleID int, username string, roleName string, orgID int) string {
	// JWT 密钥（与 config/settings.yml 中的 jwt.secret 保持一致）
	secret := []byte("jxt")

	// 构造 JWT claims，与生产环境一致
	claims := jwt.MapClaims{
		"identity":  userID,
		"roleid":    roleID,
		"rolekey":   "admin", // 简化：统一使用 admin
		"nice":      username,
		"datascope": 1, // 数据权限范围
		"rolename":  roleName,
		"org_id":    orgID,                                 // 组织ID
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
		"iat":       time.Now().Unix(),
	}

	// 使用 HS256 算法签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return ""
	}

	return "Bearer " + tokenString
}

// GenerateTestTokenWithOrgID 生成带组织ID的测试 token（别名函数，保持与 evidence-management 兼容）
func GenerateTestTokenWithOrgID(userID int, orgID int, username string, roleName string, roleID int) string {
	return GenerateTestToken(userID, roleID, username, roleName, orgID)
}
