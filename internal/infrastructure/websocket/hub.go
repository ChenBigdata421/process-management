package websocket

import (
	"encoding/json"
	websocket "jxt-evidence-system/process-management/internal/domain/aggregate/task/websocket"
	"log"
	"sync"
	"time"
)

// Hub WebSocket连接管理中心
type Hub struct {
	// 用户ID -> 连接列表的映射
	clients map[string]map[*Client]bool

	// 注册请求
	register chan *Client

	// 注销请求
	unregister chan *Client

	// 广播消息
	broadcast chan *Message

	// 互斥锁
	mu sync.RWMutex
}

// Message WebSocket消息
type Message struct {
	Type      string                 `json:"type"`      // 消息类型：task_created, task_updated, task_assigned, workflow_completed
	UserID    string                 `json:"user_id"`   // 目标用户ID
	Data      map[string]interface{} `json:"data"`      // 消息数据
	Timestamp string                 `json:"timestamp"` // 时间戳
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			log.Printf("[WebSocket] Client registered: user=%s, total=%d", client.UserID, len(h.clients[client.UserID]))

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WebSocket] Client unregistered: user=%s", client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.UserID]
			h.mu.RUnlock()

			if len(clients) == 0 {
				log.Printf("[WebSocket] No clients for user: %s", message.UserID)
				continue
			}

			// 序列化消息
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("[WebSocket] Failed to marshal message: %v", err)
				continue
			}

			// 发送给该用户的所有连接
			for client := range clients {
				select {
				case client.send <- data:
					log.Printf("[WebSocket] Message sent to user: %s, type: %s", message.UserID, message.Type)
				default:
					// 发送失败，关闭连接
					h.mu.Lock()
					close(client.send)
					delete(h.clients[message.UserID], client)
					if len(h.clients[message.UserID]) == 0 {
						delete(h.clients, message.UserID)
					}
					h.mu.Unlock()
					log.Printf("[WebSocket] Client send buffer full, closing: user=%s", message.UserID)
				}
			}
		}
	}
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userID string, msgType string, data map[string]interface{}) {
	message := &Message{
		Type:      msgType,
		UserID:    userID,
		Data:      data,
		Timestamp: getCurrentTimestamp(),
	}

	select {
	case h.broadcast <- message:
		log.Printf("[WebSocket] Message queued for user: %s, type: %s", userID, msgType)
	default:
		log.Printf("[WebSocket] Broadcast channel full, message dropped for user: %s", userID)
	}
}

// SendToUsers 发送消息给多个用户
func (h *Hub) SendToUsers(userIDs []string, msgType string, data map[string]interface{}) {
	for _, userID := range userIDs {
		h.SendToUser(userID, msgType, data)
	}
}

// GetOnlineUsers 获取在线用户列表
func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetUserConnectionCount 获取用户的连接数
func (h *Hub) GetUserConnectionCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		return len(clients)
	}
	return 0
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Close 优雅关闭 Hub，关闭所有连接和 channel
func (h *Hub) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 关闭所有客户端连接
	for _, clients := range h.clients {
		for client := range clients {
			close(client.send)
		}
	}
	h.clients = make(map[string]map[*Client]bool)

	// 关闭 channel（停止 Run() 循环）
	close(h.register)
	close(h.unregister)
	close(h.broadcast)

	log.Println("[WebSocket] Hub closed successfully")
	return nil
}

// 编译时检查：确保 Hub 实现了 WebSocketNotifier 接口
var _ websocket.WebSocketNotifier = (*Hub)(nil)
