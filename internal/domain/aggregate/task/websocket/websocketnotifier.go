package websocket

// WebSocketNotifier WebSocket通知器
type WebSocketNotifier interface {
	// 发送消息方法
	SendToUser(userID string, msgType string, data map[string]interface{})
	SendToUsers(userIDs []string, msgType string, data map[string]interface{})

	// 查询在线状态方法
	GetOnlineUsers() []string
	IsUserOnline(userID string) bool

	// 生命周期方法
	Close() error
}
