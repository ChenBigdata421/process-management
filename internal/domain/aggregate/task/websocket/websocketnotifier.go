package websocket

// WebSocketNotifier WebSocket通知器
type WebSocketNotifier interface {
	// 发送消息方法
	SendToUser(userID int, msgType string, data map[string]interface{})
	SendToUsers(userIDs []int, msgType string, data map[string]interface{})

	// 查询在线状态方法
	GetOnlineUsers() []int
	IsUserOnline(userID int) bool

	// 生命周期方法
	Close() error
}
