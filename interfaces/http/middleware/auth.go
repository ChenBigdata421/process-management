package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTClaims JWT载荷结构
type JWTClaims struct {
	Identity int    `json:"identity"` // 用户ID
	RoleID   int    `json:"roleid"`
	RoleKey  string `json:"rolekey"`
	RoleName string `json:"rolename"`
	OrgID    int    `json:"org_id"`
}

// AuthMiddleware 认证中间件
// 从Authorization头中提取user_id
// 支持三种格式：
// 1. Authorization: Bearer {jwt_token} - JWT Token（从identity字段提取用户ID）
// 2. Authorization: Bearer {user_id} - 纯用户ID
// 3. Authorization: {user_id} - 纯用户ID（无Bearer前缀）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"code": 401,
				"msg":  "unauthorized: missing authorization header",
			})
			c.Abort()
			return
		}

		// 解析Authorization头
		var token string
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			token = authHeader
		}

		if token == "" {
			c.JSON(http.StatusOK, gin.H{
				"code": 401,
				"msg":  "unauthorized: invalid authorization header",
			})
			c.Abort()
			return
		}

		// 尝试解析JWT Token
		userID := parseJWTToken(token)
		if userID == "" {
			// 如果不是JWT，直接使用token作为user_id
			userID = token
		}

		// 将user_id存入context
		c.Set("user_id", userID)
		c.Next()
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseJWTToken 解析JWT Token，提取用户ID
// JWT格式：header.payload.signature
// 我们只需要解析payload部分
func parseJWTToken(token string) string {
	// JWT Token包含两个点号
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		// 不是JWT格式
		return ""
	}

	// 解码payload（第二部分）
	payload := parts[1]

	// JWT使用base64 URL编码，需要添加padding
	if l := len(payload) % 4; l > 0 {
		payload += strings.Repeat("=", 4-l)
	}

	// Base64解码
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// 解码失败，不是有效的JWT
		return ""
	}

	// 解析JSON
	var claims JWTClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		// JSON解析失败
		return ""
	}

	// 返回用户ID（转换为字符串）
	if claims.Identity > 0 {
		return fmt.Sprintf("%d", claims.Identity)
	}

	return ""
}

// OptionalAuthMiddleware 可选认证中间件
// 如果提供了Authorization头，则验证；否则继续
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			var token string
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				token = authHeader
			}

			if token != "" {
				// 尝试解析JWT Token
				userID := parseJWTToken(token)
				if userID == "" {
					// 如果不是JWT，直接使用token作为user_id
					userID = token
				}
				c.Set("user_id", userID)
			}
		}
		c.Next()
	}
}
