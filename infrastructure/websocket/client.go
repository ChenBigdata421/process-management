package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入等待时间
	writeWait = 10 * time.Second

	// Pong等待时间
	pongWait = 60 * time.Second

	// Ping周期（必须小于pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// Client WebSocket客户端
type Client struct {
	// WebSocket连接
	conn *websocket.Conn

	// Hub引用
	hub *Hub

	// 用户ID
	UserID string

	// 发送消息的通道
	send chan []byte
}

// NewClient 创建新的客户端
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		UserID: userID,
		send:   make(chan []byte, 256),
	}
}

// readPump 从WebSocket连接读取消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] Error reading message: %v", err)
			}
			break
		}

		// 处理客户端发来的消息（如果需要）
		log.Printf("[WebSocket] Received message from user %s: %s", c.UserID, string(message))
	}
}

// writePump 向WebSocket连接写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将队列中的其他消息也一起发送
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Start 启动客户端的读写循环
func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源（生产环境应该限制）
	},
}

// ServeWs 处理WebSocket请求
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Failed to upgrade connection: %v", err)
		return
	}

	client := NewClient(hub, conn, userID)
	hub.register <- client

	client.Start()
}
