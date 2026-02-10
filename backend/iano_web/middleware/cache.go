package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	web "iano_web"
)

// CacheEntry 缓存条目
type CacheEntry struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	ExpiresAt  time.Time
}

// IsExpired 检查缓存是否过期
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// MemoryCache 内存缓存
type MemoryCache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		entries: make(map[string]*CacheEntry),
	}
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists || entry.IsExpired() {
		return nil, false
	}
	return entry, true
}

// Set 设置缓存
func (c *MemoryCache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = entry
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// Size 返回缓存条目数
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// 全局缓存实例
var globalCache = NewMemoryCache()

// CacheConfig 缓存配置
type CacheConfig struct {
	Duration     time.Duration               // 缓存时长
	KeyGenerator func(c *web.Context) string // 自定义缓存键生成器
	SkipPaths    []string                    // 跳过的路径
	Cacheable    func(c *web.Context) bool   // 自定义缓存判断函数
}

// DefaultCacheConfig 默认缓存配置
var DefaultCacheConfig = CacheConfig{
	Duration: 5 * time.Minute,
	KeyGenerator: func(c *web.Context) string {
		// 生成缓存键：方法 + URL + 查询参数
		return generateCacheKey(c.Method, c.Path, c.AllQuery())
	},
	Cacheable: func(c *web.Context) bool {
		// 只缓存 GET 和 HEAD 请求
		return c.Method == "GET" || c.Method == "HEAD"
	},
}

// Cache 响应缓存中间件（使用默认配置）
func Cache() web.HandlerFunc {
	return CacheWithConfig(DefaultCacheConfig)
}

// CacheWithDuration 带自定义时长的缓存中间件
func CacheWithDuration(duration time.Duration) web.HandlerFunc {
	config := DefaultCacheConfig
	config.Duration = duration
	return CacheWithConfig(config)
}

// CacheWithConfig 带配置的缓存中间件
func CacheWithConfig(config CacheConfig) web.HandlerFunc {
	// 使用默认值填充未设置的配置
	if config.KeyGenerator == nil {
		config.KeyGenerator = DefaultCacheConfig.KeyGenerator
	}
	if config.Cacheable == nil {
		config.Cacheable = DefaultCacheConfig.Cacheable
	}
	if config.Duration == 0 {
		config.Duration = DefaultCacheConfig.Duration
	}

	return func(c *web.Context) {
		// 检查是否应该跳过缓存
		if !shouldCache(c, config) {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := config.KeyGenerator(c)

		// 尝试从缓存获取
		if entry, found := globalCache.Get(cacheKey); found {
			// 写入缓存的响应
			writeCachedResponse(c, entry)
			return
		}

		// 包装 ResponseWriter 以捕获响应
		recorder := &responseRecorder{
			ResponseWriter: c.Writer,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}
		c.Writer = recorder

		// 执行后续处理器
		c.Next()

		// 缓存响应
		if recorder.statusCode >= 200 && recorder.statusCode < 300 {
			entry := &CacheEntry{
				StatusCode: recorder.statusCode,
				Headers:    recorder.Header().Clone(),
				Body:       recorder.body.Bytes(),
				ExpiresAt:  time.Now().Add(config.Duration),
			}
			globalCache.Set(cacheKey, entry)
		}
	}
}

// responseRecorder 用于捕获响应数据
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	written    bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.written {
		r.statusCode = code
		r.written = true
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

// shouldCache 判断是否 should 缓存当前请求
func shouldCache(c *web.Context, config CacheConfig) bool {
	// 检查是否是可缓存的方法
	if config.Cacheable != nil && !config.Cacheable(c) {
		return false
	}

	// 检查是否是跳过的路径
	for _, path := range config.SkipPaths {
		if c.Path == path {
			return false
		}
	}

	return true
}

// writeCachedResponse 写入缓存的响应
func writeCachedResponse(c *web.Context, entry *CacheEntry) {
	// 设置响应头
	for key, values := range entry.Headers {
		for _, value := range values {
			c.SetHeader(key, value)
		}
	}

	// 添加缓存标记头
	c.SetHeader("X-Cache", "HIT")

	// 写入响应
	c.Status(entry.StatusCode)
	c.Writer.Write(entry.Body)

	// 中止后续处理
	c.Abort()
}

// generateCacheKey 生成缓存键
func generateCacheKey(method, path string, query map[string][]string) string {
	// 构建查询字符串
	var queryStr string
	if len(query) > 0 {
		queryStr = fmt.Sprintf("%v", query)
	}

	// 组合并哈希
	data := fmt.Sprintf("%s:%s:%s", method, path, queryStr)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CacheClear 清除所有缓存的处理器
func CacheClear() web.HandlerFunc {
	return func(c *web.Context) {
		globalCache.Clear()
		c.JSON(200, map[string]string{
			"message": "Cache cleared",
		})
	}
}

// CacheStats 获取缓存统计信息
func CacheStats() web.HandlerFunc {
	return func(c *web.Context) {
		c.JSON(200, map[string]interface{}{
			"size": globalCache.Size(),
		})
	}
}

// CacheDelete 删除指定缓存
func CacheDelete(key string) web.HandlerFunc {
	return func(c *web.Context) {
		globalCache.Delete(key)
		c.JSON(200, map[string]string{
			"message": "Cache deleted",
		})
	}
}
