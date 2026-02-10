package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"iano_server/pkg/web"
)

func TestCache(t *testing.T) {
	// 清空缓存
	globalCache.Clear()

	engine := web.New()
	engine.Use(Cache())

	callCount := 0
	engine.GET("/cached", func(c *web.Context) {
		callCount++
		c.String(200, "Response %d", callCount)
	})

	// 第一次请求（未缓存）
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/cached", nil)
	engine.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("Expected status 200, got %d", w1.Code)
	}
	if w1.Body.String() != "Response 1" {
		t.Errorf("Expected 'Response 1', got '%s'", w1.Body.String())
	}
	if callCount != 1 {
		t.Errorf("Expected handler to be called 1 time, called %d times", callCount)
	}

	// 第二次请求（应该命中缓存）
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/cached", nil)
	engine.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("Expected status 200, got %d", w2.Code)
	}
	if w2.Body.String() != "Response 1" {
		t.Errorf("Expected cached 'Response 1', got '%s'", w2.Body.String())
	}
	if callCount != 1 {
		t.Errorf("Expected handler to still be called 1 time (cached), called %d times", callCount)
	}

	// 检查缓存标记头
	cacheHeader := w2.Header().Get("X-Cache")
	if cacheHeader != "HIT" {
		t.Errorf("Expected X-Cache header 'HIT', got '%s'", cacheHeader)
	}
}

func TestCacheWithDuration(t *testing.T) {
	// 清空缓存
	globalCache.Clear()

	engine := web.New()
	engine.Use(CacheWithDuration(100 * time.Millisecond))

	callCount := 0
	engine.GET("/cached", func(c *web.Context) {
		callCount++
		c.String(200, "Response %d", callCount)
	})

	// 第一次请求
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/cached", nil)
	engine.ServeHTTP(w1, req1)

	// 第二次请求（应该命中缓存）
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/cached", nil)
	engine.ServeHTTP(w2, req2)

	if callCount != 1 {
		t.Errorf("Expected handler to be called 1 time, called %d times", callCount)
	}

	// 等待缓存过期
	time.Sleep(150 * time.Millisecond)

	// 第三次请求（缓存已过期）
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/cached", nil)
	engine.ServeHTTP(w3, req3)

	if callCount != 2 {
		t.Errorf("Expected handler to be called 2 times after expiry, called %d times", callCount)
	}
}

func TestCacheSkipMethods(t *testing.T) {
	// 清空缓存
	globalCache.Clear()

	engine := web.New()
	engine.Use(Cache())

	callCount := 0
	engine.POST("/not-cached", func(c *web.Context) {
		callCount++
		c.String(200, "Response %d", callCount)
	})

	// 第一次 POST 请求
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/not-cached", nil)
	engine.ServeHTTP(w1, req1)

	if callCount != 1 {
		t.Errorf("Expected handler to be called 1 time, called %d times", callCount)
	}

	// 第二次 POST 请求（不应该缓存）
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/not-cached", nil)
	engine.ServeHTTP(w2, req2)

	if callCount != 2 {
		t.Errorf("Expected handler to be called 2 times (no cache for POST), called %d times", callCount)
	}
}

func TestCacheSkipPaths(t *testing.T) {
	// 清空缓存
	globalCache.Clear()

	config := CacheConfig{
		Duration:  5 * time.Minute,
		SkipPaths: []string{"/skip"},
	}

	engine := web.New()
	engine.Use(CacheWithConfig(config))

	callCount := 0
	engine.GET("/cached", func(c *web.Context) {
		callCount++
		c.String(200, "Response %d", callCount)
	})
	engine.GET("/skip", func(c *web.Context) {
		callCount++
		c.String(200, "Response %d", callCount)
	})

	// 请求 /cached 两次（应该缓存）
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/cached", nil)
		engine.ServeHTTP(w, req)
	}

	// 请求 /skip 两次（不应该缓存）
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/skip", nil)
		engine.ServeHTTP(w, req)
	}

	// /cached 应该只调用 1 次（缓存），/skip 应该调用 2 次
	if callCount != 3 {
		t.Errorf("Expected handler to be called 3 times (1 cached + 2 not cached), called %d times", callCount)
	}
}

func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache()

	// 测试设置和获取
	entry := &CacheEntry{
		StatusCode: 200,
		Headers:    http.Header{},
		Body:       []byte("test data"),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}

	cache.Set("key1", entry)

	// 测试获取
	if _, found := cache.Get("key1"); !found {
		t.Error("Expected to find cached entry")
	}

	// 测试过期
	expiredEntry := &CacheEntry{
		StatusCode: 200,
		Headers:    http.Header{},
		Body:       []byte("expired"),
		ExpiresAt:  time.Now().Add(-1 * time.Second),
	}
	cache.Set("key2", expiredEntry)

	if _, found := cache.Get("key2"); found {
		t.Error("Expected expired entry to not be found")
	}

	// 测试删除
	cache.Delete("key1")
	if _, found := cache.Get("key1"); found {
		t.Error("Expected deleted entry to not be found")
	}

	// 测试清空
	cache.Set("key3", entry)
	cache.Clear()
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
}

func TestCacheStats(t *testing.T) {
	// 清空缓存
	globalCache.Clear()

	engine := web.New()
	engine.GET("/stats", CacheStats())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCacheClear(t *testing.T) {
	// 清空缓存并添加一些数据
	globalCache.Clear()
	globalCache.Set("test", &CacheEntry{
		StatusCode: 200,
		Headers:    http.Header{},
		Body:       []byte("test"),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	})

	if globalCache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", globalCache.Size())
	}

	engine := web.New()
	engine.GET("/clear", CacheClear())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/clear", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if globalCache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", globalCache.Size())
	}
}
