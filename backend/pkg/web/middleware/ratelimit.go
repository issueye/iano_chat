package middleware

import (
	"fmt"
	"sync"
	"time"

	"iano_chat/pkg/web"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	rate       int       // 每秒产生令牌数
	capacity   int       // 桶容量
	tokens     float64   // 当前令牌数
	lastUpdate time.Time // 上次更新时间
	mu         sync.Mutex
}

// NewTokenBucket 创建新的令牌桶
func NewTokenBucket(rate, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     float64(capacity),
		lastUpdate: time.Now(),
	}
}

// Allow 检查是否允许通过（消耗一个令牌）
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.lastUpdate = now

	// 添加新产生的令牌
	tb.tokens += elapsed * float64(tb.rate)
	if tb.tokens > float64(tb.capacity) {
		tb.tokens = float64(tb.capacity)
	}

	// 尝试消耗一个令牌
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

// GetRate 获取当前限流率
func (tb *TokenBucket) GetRate() int {
	return tb.rate
}

// IPRateLimiter 基于 IP 的限流器
type IPRateLimiter struct {
	limiters map[string]*TokenBucket
	rate     int
	capacity int
	mu       sync.RWMutex
}

// NewIPRateLimiter 创建基于 IP 的限流器
func NewIPRateLimiter(rate, capacity int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*TokenBucket),
		rate:     rate,
		capacity: capacity,
	}
}

// Allow 检查 IP 是否允许通过
func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter = NewTokenBucket(rl.rate, rl.capacity)
		rl.limiters[ip] = limiter
		rl.mu.Unlock()
	}

	return limiter.Allow()
}

// GetLimiterCount 获取限流器数量
func (rl *IPRateLimiter) GetLimiterCount() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return len(rl.limiters)
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Requests  int                         // 请求数
	Per       time.Duration               // 时间窗口
	KeyFunc   func(c *web.Context) string // 自定义限流键函数
	OnLimited func(c *web.Context)        // 被限流时的回调
	SkipFunc  func(c *web.Context) bool   // 跳过的条件
}

// DefaultRateLimitConfig 默认限流配置
var DefaultRateLimitConfig = RateLimitConfig{
	Requests: 100,
	Per:      time.Minute,
	KeyFunc: func(c *web.Context) string {
		// 默认使用 IP 作为限流键
		return c.Request.RemoteAddr
	},
	OnLimited: func(c *web.Context) {
		c.String(429, "Too Many Requests")
	},
}

// RateLimit 限流中间件（使用默认配置：100请求/分钟）
func RateLimit() web.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig)
}

// RateLimitWithRequests 自定义请求数的限流中间件
func RateLimitWithRequests(requests int) web.HandlerFunc {
	config := DefaultRateLimitConfig
	config.Requests = requests
	return RateLimitWithConfig(config)
}

// RateLimitWithDuration 自定义时间窗口的限流中间件
func RateLimitWithDuration(requests int, per time.Duration) web.HandlerFunc {
	config := DefaultRateLimitConfig
	config.Requests = requests
	config.Per = per
	return RateLimitWithConfig(config)
}

// 存储全局限流器
var globalRateLimiters = struct {
	sync.RWMutex
	limiters map[string]*TokenBucket
}{limiters: make(map[string]*TokenBucket)}

// RateLimitWithConfig 带配置的限流中间件
func RateLimitWithConfig(config RateLimitConfig) web.HandlerFunc {
	// 使用默认值填充未设置的配置
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultRateLimitConfig.KeyFunc
	}
	if config.OnLimited == nil {
		config.OnLimited = DefaultRateLimitConfig.OnLimited
	}
	if config.Requests == 0 {
		config.Requests = DefaultRateLimitConfig.Requests
	}
	if config.Per == 0 {
		config.Per = DefaultRateLimitConfig.Per
	}

	return func(c *web.Context) {
		// 检查是否应该跳过限流
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// 获取限流键
		key := config.KeyFunc(c)

		// 获取或创建限流器
		globalRateLimiters.RLock()
		limiter, exists := globalRateLimiters.limiters[key]
		globalRateLimiters.RUnlock()

		if !exists {
			globalRateLimiters.Lock()
			limiter = NewTokenBucket(config.Requests, config.Requests)
			globalRateLimiters.limiters[key] = limiter
			globalRateLimiters.Unlock()
		}

		// 检查是否允许通过
		if !limiter.Allow() {
			// 添加限流响应头
			c.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", config.Requests))
			c.SetHeader("X-RateLimit-Window", config.Per.String())
			c.SetHeader("Retry-After", fmt.Sprintf("%d", int(config.Per.Seconds())))

			// 执行限流回调
			if config.OnLimited != nil {
				config.OnLimited(c)
			} else {
				c.String(429, "Too Many Requests")
			}
			c.Abort()
			return
		}

		// 添加限流信息头
		c.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", config.Requests))
		c.SetHeader("X-RateLimit-Window", config.Per.String())

		c.Next()
	}
}

// IPRateLimit 基于 IP 的限流中间件
func IPRateLimit(requests int, per time.Duration) web.HandlerFunc {
	limiter := NewIPRateLimiter(requests, requests)

	return func(c *web.Context) {
		ip := c.Request.RemoteAddr

		if !limiter.Allow(ip) {
			c.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
			c.SetHeader("X-RateLimit-Window", per.String())
			c.String(429, "Too Many Requests")
			c.Abort()
			return
		}

		c.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
		c.SetHeader("X-RateLimit-Window", per.String())
		c.Next()
	}
}

// PerSecond 每秒限流
func PerSecond(requests int) web.HandlerFunc {
	return RateLimitWithDuration(requests, time.Second)
}

// PerMinute 每分钟限流
func PerMinute(requests int) web.HandlerFunc {
	return RateLimitWithDuration(requests, time.Minute)
}

// PerHour 每小时限流
func PerHour(requests int) web.HandlerFunc {
	return RateLimitWithDuration(requests, time.Hour)
}
