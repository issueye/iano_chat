package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"iano_server/pkg/web"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	engine := web.New()
	engine.Use(LoggerWithConfig(LoggerConfig{
		Output:    logger,
		Formatter: defaultFormatter,
	}))
	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected log output, got empty")
	}

	// 检查日志包含关键信息
	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected log to contain 'GET'")
	}
	if !strings.Contains(logOutput, "/test") {
		t.Error("Expected log to contain '/test'")
	}
	if !strings.Contains(logOutput, "200") {
		t.Error("Expected log to contain '200'")
	}
}

func TestSimpleLogger(t *testing.T) {
	var buf bytes.Buffer
	// 重定向标准日志输出
	oldLog := log.Default()
	log.SetOutput(&buf)
	defer log.SetOutput(oldLog.Writer())

	engine := web.New()
	engine.Use(SimpleLogger())
	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected log output, got empty")
	}
}

func TestCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	oldLog := log.Default()
	log.SetOutput(&buf)
	defer log.SetOutput(oldLog.Writer())

	format := "[{time}] {method} {path} {status} {latency}"

	engine := web.New()
	engine.Use(CustomLogger(format))
	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected log output, got empty")
	}

	// 检查自定义格式被应用
	if strings.Contains(logOutput, "{method}") {
		t.Error("Expected {method} to be replaced")
	}
	if strings.Contains(logOutput, "{path}") {
		t.Error("Expected {path} to be replaced")
	}
}

func TestDefaultFormatter(t *testing.T) {
	engine := web.New()
	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	// 创建上下文和模拟请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/users", nil)
	c := &web.Context{
		Writer:  w,
		Request: req,
		Method:  "POST",
		Path:    "/api/users",
	}

	// 模拟请求处理
	c.Status(200)

	latency := 5 * time.Millisecond
	result := defaultFormatter(c, latency)

	if result == "" {
		t.Error("Expected formatter to return non-empty string")
	}

	// 检查包含关键信息
	if !strings.Contains(result, "POST") {
		t.Error("Expected formatter to contain method 'POST'")
	}
	if !strings.Contains(result, "/api/users") {
		t.Error("Expected formatter to contain path '/api/users'")
	}
}

func TestLoggerWithCustomFormatter(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	customFormatter := func(c *web.Context, latency time.Duration) string {
		return "CUSTOM: " + c.Method + " " + c.Path
	}

	engine := web.New()
	engine.Use(LoggerWithConfig(LoggerConfig{
		Output:    logger,
		Formatter: customFormatter,
	}))
	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "CUSTOM:") {
		t.Error("Expected custom formatter to be applied")
	}
	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected log to contain 'GET'")
	}
	if !strings.Contains(logOutput, "/test") {
		t.Error("Expected log to contain '/test'")
	}
}

func TestLoggerMultipleRequests(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	engine := web.New()
	engine.Use(LoggerWithConfig(LoggerConfig{
		Output:    logger,
		Formatter: defaultFormatter,
	}))
	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})
	engine.POST("/data", func(c *web.Context) {
		c.String(201, "Created")
	})

	// 发送多个请求
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		engine.ServeHTTP(w, req)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/data", nil)
	engine.ServeHTTP(w, req)

	logOutput := buf.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}
}
