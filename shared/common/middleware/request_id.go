package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestId 自动增加requestId
func RequestId(trafficKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		requestId := c.GetHeader(trafficKey)
		if requestId == "" {
			requestId = c.GetHeader(strings.ToLower(trafficKey))
		}
		if requestId == "" {
			requestId = uuid.New().String()
		}
		c.Request.Header.Set(trafficKey, requestId)
		c.Set(trafficKey, requestId)
		//不再使用go-admin的logger，使用我们自己基于zaplogger的封装
		//c.Set(pkg.LoggerKey,
		//	logger.NewHelper(logger.DefaultLogger).
		//		WithFields(map[string]interface{}{
		//			trafficKey: requestId,
		//		}))
		c.Next()
	}
}
