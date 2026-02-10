package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	web "iano_web"
)

func clearRateLimiters() {
	globalRateLimiters.Lock()
	globalRateLimiters.limiters = make(map[string]*TokenBucket)
	globalRateLimiters.Unlock()
}

func TestTokenBucket(t *testing.T) {
	tb := NewTokenBucket(10, 10) // 10 tokens/sec, capacity 10

	// 应该允许前 10 个请求
	for i := 0; i < 10; i++ {
		if !tb.Allow() {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}

	// 第 11 个请求应该被拒绝（桶已空）
	if tb.Allow() {
		t.Error("Expected 11th request to be denied")
	}

	// 等待 1 秒，应该产生新的令牌
	time.Sleep(1 * time.Second)

	// 现在应该允许新请求
	if !tb.Allow() {
		t.Error("Expected request to be allowed after waiting")
	}
}

func TestRateLimit(t *testing.T) {
	clearRateLimiters()

	engine := web.New()
	engine.Use(RateLimitWithRequests(3)) // 每分钟 3 个请求

	callCount := 0
	engine.GET("/test", func(c *web.Context) {
		callCount++
		c.String(200, "OK")
	})

	// 前 3 个请求应该成功
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for request %d, got %d", i+1, w.Code)
		}
	}

	// 第 4 个请求应该被限流
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("Expected status 429 (Too Many Requests), got %d", w.Code)
	}

	// 检查限流头
	rateLimit := w.Header().Get("X-RateLimit-Limit")
	if rateLimit != "3" {
		t.Errorf("Expected X-RateLimit-Limit '3', got '%s'", rateLimit)
	}

	if callCount != 3 {
		t.Errorf("Expected handler to be called 3 times, called %d times", callCount)
	}
}

func TestPerSecondRateLimit(t *testing.T) {
	clearRateLimiters()

	engine := web.New()
	engine.Use(PerSecond(2)) // 每秒 2 个请求

	callCount := 0
	engine.GET("/test", func(c *web.Context) {
		callCount++
		c.String(200, "OK")
	})

	// 前 2 个请求应该成功
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for request %d, got %d", i+1, w.Code)
		}
	}

	// 第 3 个请求应该被限流
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestPerMinuteRateLimit(t *testing.T) {
	clearRateLimiters()

	engine := web.New()
	engine.Use(PerMinute(5)) // 每分钟 5 个请求

	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	// 前 5 个请求应该成功
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for request %d, got %d", i+1, w.Code)
		}
	}

	// 第 6 个请求应该被限流
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestRateLimitWithCustomCallback(t *testing.T) {
	clearRateLimiters()

	config := RateLimitConfig{
		Requests: 1,
		Per:      time.Minute,
		OnLimited: func(c *web.Context) {
			c.JSON(429, map[string]string{
				"error": "Rate limit exceeded",
				"retry": "Please try again later",
			})
		},
	}

	engine := web.New()
	engine.Use(RateLimitWithConfig(config))

	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	// 第一个请求成功
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("Expected status 200, got %d", w1.Code)
	}

	// 第二个请求触发自定义限流回调
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w2, req2)

	if w2.Code != 429 {
		t.Errorf("Expected status 429, got %d", w2.Code)
	}

	body := w2.Body.String()
	if !contains(body, "Rate limit exceeded") {
		t.Errorf("Expected custom error message, got '%s'", body)
	}
}

func TestRateLimitSkipFunc(t *testing.T) {
	clearRateLimiters()

	config := RateLimitConfig{
		Requests: 1,
		Per:      time.Minute,
		SkipFunc: func(c *web.Context) bool {
			// 跳过 /skip 路径
			return c.Path == "/skip"
		},
	}

	engine := web.New()
	engine.Use(RateLimitWithConfig(config))

	callCount := 0
	engine.GET("/test", func(c *web.Context) {
		callCount++
		c.String(200, "OK")
	})
	engine.GET("/skip", func(c *web.Context) {
		callCount++
		c.String(200, "Skipped")
	})

	// /test 应该被限流
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("Expected status 200 for first request, got %d", w1.Code)
	}

	// 第二次 /test 应该被限流
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w2, req2)

	if w2.Code != 429 {
		t.Errorf("Expected status 429 for second request, got %d", w2.Code)
	}

	// /skip 不应该被限流，即使请求多次
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/skip", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for /skip request %d, got %d", i+1, w.Code)
		}
	}

	if callCount != 6 { // 1 (test) + 5 (skip)
		t.Errorf("Expected handler to be called 6 times, called %d times", callCount)
	}
}

func TestIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(3, 3)

	// 同一个 IP 应该被限制
	ip1 := "192.168.1.1"
	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip1) {
			t.Errorf("Expected request %d from %s to be allowed", i+1, ip1)
		}
	}

	if limiter.Allow(ip1) {
		t.Error("Expected 4th request from same IP to be denied")
	}

	// 不同的 IP 应该独立计数
	ip2 := "192.168.1.2"
	if !limiter.Allow(ip2) {
		t.Error("Expected request from different IP to be allowed")
	}
}

func TestIPRateLimitMiddleware(t *testing.T) {
	clearRateLimiters()

	engine := web.New()
	engine.Use(IPRateLimit(2, time.Minute)) // 每 IP 每分钟 2 个请求

	callCount := 0
	engine.GET("/test", func(c *web.Context) {
		callCount++
		c.String(200, "OK")
	})

	// 使用相同的 RemoteAddr（默认）
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		engine.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for request %d, got %d", i+1, w.Code)
		}
	}

	// 第 3 个请求应该被限流
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestPerHourRateLimit(t *testing.T) {
	clearRateLimiters()

	engine := web.New()
	engine.Use(PerHour(1000)) // 每小时 1000 个请求

	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查限流窗口头
	window := w.Header().Get("X-RateLimit-Window")
	if window != "1h0m0s" {
		t.Errorf("Expected window '1h0m0s', got '%s'", window)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
