package ratelimit

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Allow() bool
	Wait(ctx context.Context) error
	Limit() rate.Limit
}

type defaultRateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(r rate.Limit, b int) RateLimiter {
	return &defaultRateLimiter{
		limiter: rate.NewLimiter(r, b),
	}
}

func (rl *defaultRateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

func (rl *defaultRateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

func (rl *defaultRateLimiter) Limit() rate.Limit {
	return rl.limiter.Limit()
}

type AgentRateLimiter struct {
	globalLimiter RateLimiter
	userLimiters  map[string]RateLimiter
	mu            sync.RWMutex
	userRate      rate.Limit
	userBurst     int
}

func NewAgentRateLimiter(globalRPS float64, globalBurst int, userRPS float64, userBurst int) *AgentRateLimiter {
	return &AgentRateLimiter{
		globalLimiter: NewRateLimiter(rate.Limit(globalRPS), globalBurst),
		userLimiters:  make(map[string]RateLimiter),
		userRate:      rate.Limit(userRPS),
		userBurst:     userBurst,
	}
}

func (arl *AgentRateLimiter) Allow() bool {
	return arl.globalLimiter.Allow()
}

func (arl *AgentRateLimiter) AllowForUser(userID string) bool {
	if !arl.globalLimiter.Allow() {
		return false
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

	return userLimiter.Allow()
}

func (arl *AgentRateLimiter) Wait(ctx context.Context) error {
	return arl.globalLimiter.Wait(ctx)
}

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

type Config struct {
	Enabled     bool
	GlobalRPS   float64
	GlobalBurst int
	UserRPS     float64
	UserBurst   int
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:     true,
		GlobalRPS:   100,
		GlobalBurst: 150,
		UserRPS:     10,
		UserBurst:   20,
	}
}
