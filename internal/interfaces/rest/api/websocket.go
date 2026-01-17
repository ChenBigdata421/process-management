package api

import (
	"net/http"

	domain_websocket "jxt-evidence-system/process-management/internal/domain/aggregate/task/websocket"
	infra_websocket "jxt-evidence-system/process-management/internal/infrastructure/websocket"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	hub domain_websocket.WebSocketNotifier
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(hub domain_websocket.WebSocketNotifier) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "user_id is required",
		})
		return
	}

	infra_websocket.ServeWs(h.hub, c.Writer, c.Request, userID)
}

// GetOnlineUsers 获取在线用户列表
func (h *WebSocketHandler) GetOnlineUsers(c *gin.Context) {
	users := h.hub.GetOnlineUsers()
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"users": users,
			"count": len(users),
		},
		"msg": "success",
	})
}

// CheckUserOnline 检查用户是否在线
func (h *WebSocketHandler) CheckUserOnline(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "user_id is required",
		})
		return
	}

	online := h.hub.IsUserOnline(userID)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"user_id": userID,
			"online":  online,
		},
		"msg": "success",
	})
}

// SendTestMessage 发送测试消息（用于调试）
func (h *WebSocketHandler) SendTestMessage(c *gin.Context) {
	var req struct {
		UserID  string                 `json:"user_id"`
		Type    string                 `json:"type"`
		Message map[string]interface{} `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}

	h.hub.SendToUser(req.UserID, req.Type, req.Message)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "message sent",
	})
}
