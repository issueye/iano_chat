package sse

import (
	web "iano_web"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEventString(t *testing.T) {
	tests := []struct {
		name     string
		event    web.SSEvent
		expected string
	}{
		{
			name: "Simple Data",
			event: web.SSEvent{
				Data: "Hello, World!",
			},
			expected: "data: Hello, World!\n\n",
		},
		{
			name: "With Event Type",
			event: web.SSEvent{
				Event: "message",
				Data:  "Hello",
			},
			expected: "event: message\ndata: Hello\n\n",
		},
		{
			name: "With ID",
			event: web.SSEvent{
				ID:   "123",
				Data: "Hello",
			},
			expected: "id: 123\ndata: Hello\n\n",
		},
		{
			name: "With Retry",
			event: web.SSEvent{
				Data:  "Hello",
				Retry: 5000,
			},
			expected: "retry: 5000\ndata: Hello\n\n",
		},
		{
			name: "Complete Event",
			event: web.SSEvent{
				ID:    "123",
				Event: "message",
				Data:  "Hello",
				Retry: 5000,
			},
			expected: "id: 123\nevent: message\nretry: 5000\ndata: Hello\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.event.String()
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestSSEventWithJSON(t *testing.T) {
	event := web.SSEvent{
		Event: "user",
		Data: map[string]string{
			"name": "John",
			"age":  "30",
		},
	}

	result := event.String()

	if !strings.Contains(result, "event: user") {
		t.Error("Expected event type 'user'")
	}

	if !strings.Contains(result, "data: {") {
		t.Error("Expected JSON data")
	}
}

func TestSSEContext(t *testing.T) {
	engine := web.New()

	engine.GET("/sse", func(c *web.Context) {
		sse, err := c.SSE()
		if err != nil {
			t.Errorf("Failed to create SSE context: %v", err)
			return
		}

		// 发送事件
		sse.Emit(&web.SSEvent{
			Event: "test",
			Data:  "Hello",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sse", nil)
	req.Header.Set("Accept", "text/event-stream")

	go func() {
		engine.ServeHTTP(w, req)
	}()

	// 等待响应
	time.Sleep(100 * time.Millisecond)

	// 检查响应头
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got '%s'", w.Header().Get("Content-Type"))
	}

	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Error("Expected Cache-Control 'no-cache'")
	}
}

func TestHandleSSE(t *testing.T) {
	engine := web.New()

	engine.GET("/sse", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
		// 发送多个事件
		sse.EmitEvent("message", "Event 1")
		sse.EmitEvent("message", "Event 2")
		sse.EmitData("Event 3")

		// 关闭连接
		go func() {
			time.Sleep(100 * time.Millisecond)
			sse.Close()
		}()
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sse", nil)

	engine.ServeHTTP(w, req)

	body := w.Body.String()

	if !strings.Contains(body, "event: message") {
		t.Error("Expected 'event: message' in response")
	}

	if !strings.Contains(body, "data: Event 1") {
		t.Errorf("Expected 'Event 1' in response, got: %s", body)
	}
}

func TestSSEPing(t *testing.T) {
	engine := web.New()

	engine.GET("/sse", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
		sse.Ping()
		sse.Close()
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sse", nil)

	engine.ServeHTTP(w, req)

	body := w.Body.String()

	if !strings.Contains(body, "event: ping") {
		t.Error("Expected 'event: ping' in response")
	}
}

func TestSSEHub(t *testing.T) {
	hub := web.NewSSEHub()

	// 检查初始状态
	if hub.ClientCount() != 0 {
		t.Errorf("Expected 0 clients initially, got %d", hub.ClientCount())
	}

	// 启动 hub
	go hub.Run()

	// 注意：完整的 Hub 测试需要真实的 HTTP 连接
	// 这里只测试基本结构和接口
}

func TestSSEContextClose(t *testing.T) {
	engine := web.New()

	closed := false
	engine.GET("/sse", func(c *web.Context) {
		sse, _ := c.SSE()

		sse.Close()
		closed = sse.IsClosed()
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sse", nil)

	engine.ServeHTTP(w, req)

	if !closed {
		t.Error("Expected SSE context to be closed")
	}
}

func TestSSEContextEmitAfterClose(t *testing.T) {
	engine := web.New()

	engine.GET("/sse", func(c *web.Context) {
		sse, _ := c.SSE()

		sse.Close()

		err := sse.Emit(&web.SSEvent{Data: "test"})
		if err == nil {
			t.Error("Expected error when emitting to closed connection")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sse", nil)

	engine.ServeHTTP(w, req)
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "Hello",
			expected: []string{"Hello"},
		},
		{
			input:    "Hello\nWorld",
			expected: []string{"Hello", "World"},
		},
		{
			input:    "Line1\nLine2\nLine3",
			expected: []string{"Line1", "Line2", "Line3"},
		},
		{
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		result := web.SplitLines(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("Expected line %d to be '%s', got '%s'", i, tt.expected[i], result[i])
			}
		}
	}
}

func TestSSEHeartbeat(t *testing.T) {
	engine := web.New()

	pingCount := 0
	engine.GET("/sse", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
		// 启动快速心跳（100ms）
		ticker := sse.StartHeartbeat(100 * time.Millisecond)
		defer ticker.Stop()

		// 等待 2 个心跳
		time.Sleep(250 * time.Millisecond)
		sse.Close()
	}))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sse", nil)

	engine.ServeHTTP(w, req)

	// 检查是否有心跳事件
	body := w.Body.String()
	pingCount = strings.Count(body, "event: ping")

	if pingCount < 2 {
		t.Errorf("Expected at least 2 ping events, got %d", pingCount)
	}
}

func TestSSEMiddleware(t *testing.T) {
	engine := web.New()

	connected := false
	disconnected := false

	config := web.SSEMiddlewareConfig{
		HeartbeatInterval: 100 * time.Millisecond,
		OnConnect: func(sse *web.SSEContext, c *web.Context) {
			connected = true
			// 发送欢迎消息
			sse.EmitData("Welcome!")
			// 立即关闭以便测试
			go func() {
				time.Sleep(50 * time.Millisecond)
				sse.Close()
			}()
		},
		OnDisconnect: func(clientID string) {
			disconnected = true
		},
	}

	engine.GET("/sse", web.SSEMiddleware(config))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sse", nil)

	engine.ServeHTTP(w, req)

	if !connected {
		t.Error("Expected OnConnect to be called")
	}

	if !disconnected {
		t.Error("Expected OnDisconnect to be called")
	}

	body := w.Body.String()
	if !strings.Contains(body, "data: Welcome!") {
		t.Errorf("Expected 'Welcome!' in response, got: %s", body)
	}
}

func TestSSEWithMultilineData(t *testing.T) {
	event := web.SSEvent{
		Event: "message",
		Data:  "Line 1\nLine 2\nLine 3",
	}

	result := event.String()

	expectedLines := []string{
		"event: message",
		"data: Line 1",
		"data: Line 2",
		"data: Line 3",
		"",
	}

	for _, line := range expectedLines {
		if !strings.Contains(result, line) {
			t.Errorf("Expected '%s' in result", line)
		}
	}
}

// SSEClient 用于测试的 SSE 客户端
type SSEClient struct {
	Events chan web.SSEvent
	Errors chan error
}

func NewSSEClient() *SSEClient {
	return &SSEClient{
		Events: make(chan web.SSEvent, 10),
		Errors: make(chan error, 1),
	}
}

func (c *SSEClient) Connect(url string) {
	// 实际测试中使用 httptest
}
