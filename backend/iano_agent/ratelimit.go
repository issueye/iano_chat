package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	// Allow 检查是否允许执行
	Allow() bool
	// Wait 等待直到允许执行
	Wait(ctx context.Context) error
	// Limit 获取当前限制
	Limit() rate.Limit
}

// defaultRateLimiter 默认限流器实现
type defaultRateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter 创建新的限流器
// r: 每秒允许的请求数
// b: 桶容量（突发流量）
func NewRateLimiter(r rate.Limit, b int) RateLimiter {
	return &defaultRateLimiter{
		limiter: rate.NewLimiter(r, b),
	}
}

// Allow 检查是否允许执行
func (rl *defaultRateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

// Wait 等待直到允许执行
func (rl *defaultRateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// Limit 获取当前限制
func (rl *defaultRateLimiter) Limit() rate.Limit {
	return rl.limiter.Limit()
}

// AgentRateLimiter Agent 级别的限流管理
type AgentRateLimiter struct {
	// 全局限流器（所有请求）
	globalLimiter RateLimiter
	// 每用户限流器
	userLimiters map[string]RateLimiter
	mu           sync.RWMutex
	// 每用户限制配置
	userRate  rate.Limit
	userBurst int
}

// NewAgentRateLimiter 创建 Agent 限流器
func NewAgentRateLimiter(globalRPS float64, globalBurst int, userRPS float64, userBurst int) *AgentRateLimiter {
	return &AgentRateLimiter{
		globalLimiter: NewRateLimiter(rate.Limit(globalRPS), globalBurst),
		userLimiters:  make(map[string]RateLimiter),
		userRate:      rate.Limit(userRPS),
		userBurst:     userBurst,
	}
}

// Allow 检查是否允许执行（全局）
func (arl *AgentRateLimiter) Allow() bool {
	return arl.globalLimiter.Allow()
}

// AllowForUser 检查指定用户是否允许执行
func (arl *AgentRateLimiter) AllowForUser(userID string) bool {
	if !arl.globalLimiter.Allow() {
		return false
	}

	arl.mu.RLock()
	userLimiter, exists := arl.userLimiters[userID]
	arl.mu.RUnlock()

	if !exists {
		arl.mu.Lock()
		// 双重检查
		userLimiter, exists = arl.userLimiters[userID]
		if !exists {
			userLimiter = NewRateLimiter(arl.userRate, arl.userBurst)
			arl.userLimiters[userID] = userLimiter
		}
		arl.mu.Unlock()
	}

	return userLimiter.Allow()
}

// Wait 等待直到允许执行（全局）
func (arl *AgentRateLimiter) Wait(ctx context.Context) error {
	return arl.globalLimiter.Wait(ctx)
}

// WaitForUser 等待指定用户直到允许执行
func (arl *AgentRateLimiter) WaitForUser(ctx context.Context, userID string) error {
	if err := arl.globalLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("全局限流等待失败: %w", err)
	}

	arl.mu.RLock()
	userLimiter, exists := arl.userLimiters[userID]
	arl.mu.RUnlock()

	if !exists {
		arl.mu.Lock()
		userLimiter, exists = arl.userLimiters[userID]
		if !exists {
			userLimiter = NewRateLimiter(arl.userRate, arl.userBurst)
			arl.userLimiters[userID] = userLimiter
		}
		arl.mu.Unlock()
	}

	if err := userLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("用户限流等待失败: %w", err)
	}

	return nil
}

// CleanupInactiveUsers 清理不活跃用户的限流器
func (arl *AgentRateLimiter) CleanupInactiveUsers(inactiveDuration time.Duration) {
	// 注意：当前实现不追踪活跃时间，如果需要可以实现更复杂的逻辑
	// 这里提供一个简单的清理接口
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 是否启用限流
	Enabled bool
	// 全局每秒请求数
	GlobalRPS float64
	// 全局突发流量
	GlobalBurst int
	// 每用户每秒请求数
	UserRPS float64
	// 每用户突发流量
	UserBurst int
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Enabled:     true,
		GlobalRPS:   100, // 全局每秒 100 请求
		GlobalBurst: 150, // 全局突发 150 请求
		UserRPS:     10,  // 每用户每秒 10 请求
		UserBurst:   20,  // 每用户突发 20 请求
	}
}

// TokenBucketRateLimiter 基于 Token Bucket 的简单限流器
type TokenBucketRateLimiter struct {
	tokens   float64
	capacity float64
	rate     float64
	lastTime time.Time
	mu       sync.Mutex
}

// NewTokenBucketRateLimiter 创建 Token Bucket 限流器
// rate: 每秒产生的 token 数
// capacity: 桶容量
func NewTokenBucketRateLimiter(rate, capacity float64) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		tokens:   capacity,
		capacity: capacity,
		rate:     rate,
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许执行
func (tb *TokenBucketRateLimiter) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.lastTime = now

	// 添加新产生的 token
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

// Wait 等待直到允许执行
func (tb *TokenBucketRateLimiter) Wait(ctx context.Context) error {
	for {
		if tb.Allow() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// 继续尝试
		}
	}
}

// Limit 获取当前限制
func (tb *TokenBucketRateLimiter) Limit() rate.Limit {
	return rate.Limit(tb.rate)
}
