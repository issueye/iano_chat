package web

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader 是 WebSocket 升级器
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源，生产环境应该配置更严格的策略
		return true
	},
}

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool
}

// SetWebSocketConfig 设置全局 WebSocket 配置
func SetWebSocketConfig(config WebSocketConfig) {
	if config.ReadBufferSize > 0 {
		upgrader.ReadBufferSize = config.ReadBufferSize
	}
	if config.WriteBufferSize > 0 {
		upgrader.WriteBufferSize = config.WriteBufferSize
	}
	if config.CheckOrigin != nil {
		upgrader.CheckOrigin = config.CheckOrigin
	}
}

// UpgradeWebSocket 将 HTTP 连接升级为 WebSocket
func (c *Context) UpgradeWebSocket() (*websocket.Conn, error) {
	return upgrader.Upgrade(c.Writer, c.Request, nil)
}

// UpgradeWebSocketWithHeader 带自定义响应头的 WebSocket 升级
func (c *Context) UpgradeWebSocketWithHeader(responseHeader http.Header) (*websocket.Conn, error) {
	return upgrader.Upgrade(c.Writer, c.Request, responseHeader)
}

// IsWebSocket 检查请求是否为 WebSocket 升级请求
func (c *Context) IsWebSocket() bool {
	return c.GetHeader("Upgrade") == "websocket"
}

// WebSocket 简化的 WebSocket 处理函数类型
type WebSocketHandler func(conn *websocket.Conn, c *Context)

// HandleWebSocket 创建一个处理 WebSocket 的路由处理器
func HandleWebSocket(handler WebSocketHandler) HandlerFunc {
	return func(c *Context) {
		conn, err := c.UpgradeWebSocket()
		if err != nil {
			c.String(400, "Failed to upgrade to WebSocket: %v", err)
			return
		}
		defer conn.Close()
		handler(conn, c)
	}
}

// WebSocketHub WebSocket 连接管理器
type WebSocketHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// NewWebSocketHub 创建新的 WebSocket 管理器
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run 启动 WebSocket 管理器
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(h.clients, client)
				}
			}
		}
	}
}

// Register 注册新连接
func (h *WebSocketHub) Register(conn *websocket.Conn) {
	h.register <- conn
}

// Unregister 注销连接
func (h *WebSocketHub) Unregister(conn *websocket.Conn) {
	h.unregister <- conn
}

// Broadcast 广播消息给所有客户端
func (h *WebSocketHub) Broadcast(message []byte) {
	h.broadcast <- message
}

// BroadcastString 广播字符串消息
func (h *WebSocketHub) BroadcastString(message string) {
	h.Broadcast([]byte(message))
}

// SendTo 发送消息给指定客户端
func (h *WebSocketHub) SendTo(conn *websocket.Conn, messageType int, data []byte) error {
	return conn.WriteMessage(messageType, data)
}

// ClientCount 返回当前连接数
func (h *WebSocketHub) ClientCount() int {
	return len(h.clients)
}
