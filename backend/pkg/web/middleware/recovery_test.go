package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"iano_chat/pkg/web"
)

func TestRecovery(t *testing.T) {
	engine := web.New()
	engine.Use(Recovery())
	engine.GET("/panic", func(c *web.Context) {
		panic("test panic")
	})
	engine.GET("/normal", func(c *web.Context) {
		c.String(200, "OK")
	})

	// 测试正常请求
	t.Run("Normal Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/normal", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
		}
	})

	// 测试 panic 恢复
	t.Run("Panic Recovery", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/panic", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 500 {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
		if w.Body.String() != "Internal Server Error" {
			t.Errorf("Expected body 'Internal Server Error', got '%s'", w.Body.String())
		}
	})
}

func TestRecoveryWithCustomHandler(t *testing.T) {
	customHandler := func(c *web.Context, err interface{}) {
		c.String(500, "Custom Error: %v", err)
	}

	engine := web.New()
	engine.Use(RecoveryWithConfig(RecoveryConfig{
		Handler: customHandler,
	}))
	engine.GET("/panic", func(c *web.Context) {
		panic("custom panic message")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	expectedBody := "Custom Error: custom panic message"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, w.Body.String())
	}
}

func TestRecoveryWithLog(t *testing.T) {
	var buf bytes.Buffer

	engine := web.New()
	slogLogger := slog.New(slog.NewTextHandler(&buf, nil))
	engine.Use(RecoveryWithLog(slogLogger))
	engine.GET("/panic", func(c *web.Context) {
		panic("logged panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "[PANIC]") {
		t.Error("Expected log to contain '[PANIC]'")
	}
	if !strings.Contains(logOutput, "logged panic") {
		t.Error("Expected log to contain panic message")
	}
	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected log to contain method 'GET'")
	}
	if !strings.Contains(logOutput, "/panic") {
		t.Error("Expected log to contain path '/panic'")
	}
}

func TestRecoveryWithNestedPanic(t *testing.T) {
	engine := web.New()
	engine.Use(Recovery())

	callCount := 0
	engine.GET("/nested", func(c *web.Context) {
		callCount++
		if callCount == 1 {
			panic("first panic")
		}
		c.String(200, "OK")
	})

	// 第一次请求触发 panic
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/nested", nil)
	engine.ServeHTTP(w1, req1)

	if w1.Code != 500 {
		t.Errorf("Expected status 500 for first request, got %d", w1.Code)
	}

	// 第二次请求应该正常工作
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/nested", nil)
	engine.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("Expected status 200 for second request, got %d", w2.Code)
	}
	if w2.Body.String() != "OK" {
		t.Errorf("Expected body 'OK' for second request, got '%s'", w2.Body.String())
	}
}

func TestRecoveryPreservesOtherMiddleware(t *testing.T) {
	middlewareCalled := false

	engine := web.New()
	engine.Use(func(c *web.Context) {
		middlewareCalled = true
		c.Next()
	})
	engine.Use(Recovery())
	engine.GET("/panic", func(c *web.Context) {
		panic("test")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	engine.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Expected middleware to be called before panic")
	}
}
