package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestIsWebSocket(t *testing.T) {
	tests := []struct {
		name     string
		upgrade  string
		wantTrue bool
	}{
		{
			name:     "WebSocket Request",
			upgrade:  "websocket",
			wantTrue: true,
		},
		{
			name:     "Normal HTTP Request",
			upgrade:  "",
			wantTrue: false,
		},
		{
			name:     "Other Upgrade",
			upgrade:  "h2c",
			wantTrue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			engine.GET("/test", func(c *Context) {
				if c.IsWebSocket() != tt.wantTrue {
					t.Errorf("IsWebSocket() = %v, want %v", c.IsWebSocket(), tt.wantTrue)
				}
				c.String(200, "OK")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.upgrade != "" {
				req.Header.Set("Upgrade", tt.upgrade)
			}
			engine.ServeHTTP(w, req)
		})
	}
}

func TestWebSocketConfig(t *testing.T) {
	// 测试设置 WebSocket 配置
	config := WebSocketConfig{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == "http://localhost:3000"
		},
	}

	SetWebSocketConfig(config)

	// 验证配置是否生效
	if upgrader.ReadBufferSize != 2048 {
		t.Errorf("Expected ReadBufferSize 2048, got %d", upgrader.ReadBufferSize)
	}
	if upgrader.WriteBufferSize != 2048 {
		t.Errorf("Expected WriteBufferSize 2048, got %d", upgrader.WriteBufferSize)
	}
	if upgrader.CheckOrigin == nil {
		t.Error("Expected CheckOrigin to be set")
	}

	// 重置为默认值
	SetWebSocketConfig(WebSocketConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})
}

func TestWebSocketHub(t *testing.T) {
	hub := NewWebSocketHub()

	// 测试初始状态
	if hub.ClientCount() != 0 {
		t.Errorf("Expected 0 clients initially, got %d", hub.ClientCount())
	}

	// 注意：由于 WebSocket 需要真实的 HTTP 连接，
	// 完整的 WebSocket 测试需要在集成测试中进行
	// 这里只测试 Hub 的基本结构和接口
}

// mockWebSocketServer 创建一个用于测试的 WebSocket 服务器
func mockWebSocketServer(t *testing.T) (*httptest.Server, *websocket.Conn) {
	engine := New()

	engine.GET("/ws", HandleWebSocket(func(conn *websocket.Conn, c *Context) {
		// 简单的 echo 服务器
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	}))

	server := httptest.NewServer(engine)

	// 转换为 WebSocket URL
	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}

	return server, conn
}

func TestWebSocketEcho(t *testing.T) {
	// 创建测试服务器和 WebSocket 连接
	engine := New()

	messageReceived := make(chan string, 1)

	engine.GET("/ws", HandleWebSocket(func(conn *websocket.Conn, c *Context) {
		// 读取一条消息并回复
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		messageReceived <- string(msg)
		conn.WriteMessage(msgType, msg)
	}))

	server := httptest.NewServer(engine)
	defer server.Close()

	// 转换为 WebSocket URL
	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// 发送测试消息
	testMessage := "Hello, WebSocket!"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(testMessage)); err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// 读取响应
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if string(response) != testMessage {
		t.Errorf("Expected '%s', got '%s'", testMessage, string(response))
	}
}

func TestWebSocketUpgradeFail(t *testing.T) {
	engine := New()
	engine.GET("/ws", HandleWebSocket(func(conn *websocket.Conn, c *Context) {
		// 不应该执行到这里
		t.Error("Handler should not be called for non-WebSocket request")
	}))

	// 发送普通 HTTP 请求（非 WebSocket）
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ws", nil)
	engine.ServeHTTP(w, req)

	// 应该返回 400 错误
	if w.Code != 400 {
		t.Errorf("Expected status 400 for non-WebSocket request, got %d", w.Code)
	}
}
