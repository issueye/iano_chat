package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SSEvent SSE 事件结构
type SSEvent struct {
	ID    string      `json:"id,omitempty"`
	Event string      `json:"event,omitempty"`
	Data  interface{} `json:"data"`
	Retry int         `json:"retry,omitempty"`
}

// String 将事件转换为 SSE 格式字符串
func (e *SSEvent) String() string {
	var result string

	if e.ID != "" {
		result += fmt.Sprintf("id: %s\n", e.ID)
	}

	if e.Event != "" {
		result += fmt.Sprintf("event: %s\n", e.Event)
	}

	if e.Retry > 0 {
		result += fmt.Sprintf("retry: %d\n", e.Retry)
	}

	// 处理数据
	var dataStr string
	switch v := e.Data.(type) {
	case string:
		dataStr = v
	case []byte:
		dataStr = string(v)
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			dataStr = fmt.Sprintf("%v", v)
		} else {
			dataStr = string(bytes)
		}
	}

	// SSE 数据可以有多行，每行都要加 "data: " 前缀
	lines := splitLines(dataStr)
	for _, line := range lines {
		result += fmt.Sprintf("data: %s\n", line)
	}

	result += "\n"
	return result
}

// splitLines 将字符串按行分割
func splitLines(s string) []string {
	var lines []string
	var current string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" || len(lines) > 0 {
		lines = append(lines, current)
	}
	return lines
}

// SSEContext SSE 上下文
type SSEContext struct {
	*Context
	flusher http.Flusher
	closeCh chan struct{}
	mu      sync.Mutex
	closed  bool
}

// SSE 创建 SSE 上下文
func (c *Context) SSE() (*SSEContext, error) {
	// 检查是否支持 SSE
	if c.GetHeader("Accept") != "text/event-stream" {
		// 不强制检查，允许直接调用
	}

	// 设置 SSE 响应头
	c.SetHeader("Content-Type", "text/event-stream")
	c.SetHeader("Cache-Control", "no-cache")
	c.SetHeader("Connection", "keep-alive")
	c.SetHeader("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 获取 Flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming unsupported")
	}

	// 写入状态码和头信息
	c.Status(200)
	flusher.Flush()

	return &SSEContext{
		Context: c,
		flusher: flusher,
		closeCh: make(chan struct{}),
	}, nil
}

// Emit 发送事件
func (s *SSEContext) Emit(event *SSEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("sse connection closed")
	}

	_, err := fmt.Fprint(s.Writer, event.String())
	if err != nil {
		return err
	}

	s.flusher.Flush()
	return nil
}

// EmitData 发送简单数据事件
func (s *SSEContext) EmitData(data interface{}) error {
	return s.Emit(&SSEvent{Data: data})
}

// EmitEvent 发送指定类型的事件
func (s *SSEContext) EmitEvent(eventType string, data interface{}) error {
	return s.Emit(&SSEvent{Event: eventType, Data: data})
}

// EmitID 发送带 ID 的事件
func (s *SSEContext) EmitID(id string, data interface{}) error {
	return s.Emit(&SSEvent{ID: id, Data: data})
}

// Ping 发送心跳（空数据）
func (s *SSEContext) Ping() error {
	return s.Emit(&SSEvent{Event: "ping"})
}

// Close 关闭连接
func (s *SSEContext) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.closed {
		s.closed = true
		close(s.closeCh)
	}
}

// IsClosed 检查连接是否已关闭
func (s *SSEContext) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

// Done 返回关闭信号通道
func (s *SSEContext) Done() <-chan struct{} {
	return s.closeCh
}

// StartHeartbeat 启动心跳机制
func (s *SSEContext) StartHeartbeat(interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := s.Ping(); err != nil {
					s.Close()
					return
				}
			case <-s.closeCh:
				ticker.Stop()
				return
			}
		}
	}()

	return ticker
}

// HandleSSE 创建 SSE 处理器
func HandleSSE(handler func(sse *SSEContext, c *Context)) HandlerFunc {
	return func(c *Context) {
		sse, err := c.SSE()
		if err != nil {
			c.String(500, "SSE not supported: %v", err)
			return
		}

		// 启动心跳（30 秒）
		sse.StartHeartbeat(30 * time.Second)

		// 调用处理器
		handler(sse, c)
	}
}

// SSEHub SSE 广播管理器
type SSEHub struct {
	clients    map[string]*SSEContext
	broadcast  chan *SSEvent
	register   chan *SSEContext
	unregister chan string
	mu         sync.RWMutex
}

// NewSSEHub 创建新的 SSE 管理器
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients:    make(map[string]*SSEContext),
		broadcast:  make(chan *SSEvent),
		register:   make(chan *SSEContext),
		unregister: make(chan string),
	}
}

// Run 启动 SSE 管理器
func (h *SSEHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 使用 RemoteAddr 作为客户端 ID
			clientID := client.Request.RemoteAddr
			h.clients[clientID] = client
			h.mu.Unlock()

		case clientID := <-h.unregister:
			h.mu.Lock()
			if client, ok := h.clients[clientID]; ok {
				delete(h.clients, clientID)
				client.Close()
			}
			h.mu.Unlock()

		case event := <-h.broadcast:
			h.mu.RLock()
			clients := make(map[string]*SSEContext)
			for k, v := range h.clients {
				clients[k] = v
			}
			h.mu.RUnlock()

			// 广播给所有客户端
			for clientID, client := range clients {
				if err := client.Emit(event); err != nil {
					// 发送失败，注销客户端
					h.unregister <- clientID
				}
			}
		}
	}
}

// Register 注册客户端
func (h *SSEHub) Register(client *SSEContext) {
	h.register <- client
}

// Unregister 注销客户端
func (h *SSEHub) Unregister(clientID string) {
	h.unregister <- clientID
}

// Broadcast 广播事件
func (h *SSEHub) Broadcast(event *SSEvent) {
	h.broadcast <- event
}

// BroadcastData 广播数据
func (h *SSEHub) BroadcastData(data interface{}) {
	h.Broadcast(&SSEvent{Data: data})
}

// BroadcastEvent 广播指定类型事件
func (h *SSEHub) BroadcastEvent(eventType string, data interface{}) {
	h.Broadcast(&SSEvent{Event: eventType, Data: data})
}

// ClientCount 返回客户端数量
func (h *SSEHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// SSEMiddlewareConfig SSE 中间件配置
type SSEMiddlewareConfig struct {
	HeartbeatInterval time.Duration
	OnConnect         func(sse *SSEContext, c *Context)
	OnDisconnect      func(clientID string)
}

// SSEMiddleware SSE 中间件
func SSEMiddleware(config SSEMiddlewareConfig) HandlerFunc {
	return func(c *Context) {
		sse, err := c.SSE()
		if err != nil {
			c.String(500, "SSE not supported: %v", err)
			return
		}

		// 启动心跳
		if config.HeartbeatInterval > 0 {
			sse.StartHeartbeat(config.HeartbeatInterval)
		} else {
			sse.StartHeartbeat(30 * time.Second)
		}

		// 调用连接回调
		if config.OnConnect != nil {
			config.OnConnect(sse, c)
		}

		// 等待连接关闭
		<-sse.Done()

		// 调用断开连接回调
		if config.OnDisconnect != nil {
			config.OnDisconnect(c.Request.RemoteAddr)
		}
	}
}

// ContextWithCancel 创建可取消的 SSE 上下文
func (s *SSEContext) ContextWithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	go func() {
		select {
		case <-s.closeCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}
