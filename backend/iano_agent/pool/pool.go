package pool

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
)

type Pool struct {
	instances     sync.Map
	chatModel     model.ToolCallingChatModel
	mu            sync.RWMutex
	maxIdleTime   time.Duration
	cleanupTicker *time.Ticker
	done          chan struct{}
}

type PoolConfig struct {
	MaxIdleTime     time.Duration
	CleanupInterval time.Duration
}

func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxIdleTime:     30 * time.Minute,
		CleanupInterval: 5 * time.Minute,
	}
}

type Instance struct {
	Agent      interface{}
	LastActive time.Time
	SessionID  string
	AgentID    string
	CreatedAt  time.Time
}

func NewPool(chatModel model.ToolCallingChatModel, cfg *PoolConfig) *Pool {
	if cfg == nil {
		cfg = DefaultPoolConfig()
	}

	p := &Pool{
		chatModel:   chatModel,
		maxIdleTime: cfg.MaxIdleTime,
		done:        make(chan struct{}),
	}

	p.cleanupTicker = time.NewTicker(cfg.CleanupInterval)
	go p.cleanupLoop()

	return p
}

func (p *Pool) Store(key string, instance *Instance) {
	instance.LastActive = time.Now()
	p.instances.Store(key, instance)
}

func (p *Pool) Load(key string) (*Instance, bool) {
	if v, ok := p.instances.Load(key); ok {
		instance := v.(*Instance)
		instance.LastActive = time.Now()
		return instance, true
	}
	return nil, false
}

func (p *Pool) Delete(key string) {
	p.instances.Delete(key)
}

func (p *Pool) Range(f func(key string, instance *Instance) bool) {
	p.instances.Range(func(k, v interface{}) bool {
		return f(k.(string), v.(*Instance))
	})
}

func (p *Pool) Count() int {
	count := 0
	p.instances.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (p *Pool) Clear() {
	p.instances.Range(func(key, value interface{}) bool {
		p.instances.Delete(key)
		return true
	})
}

func (p *Pool) Close() {
	close(p.done)
	p.cleanupTicker.Stop()
	p.Clear()
}

func (p *Pool) GetChatModel() model.ToolCallingChatModel {
	return p.chatModel
}

func (p *Pool) cleanupLoop() {
	for {
		select {
		case <-p.done:
			return
		case <-p.cleanupTicker.C:
			p.cleanup()
		}
	}
}

func (p *Pool) cleanup() {
	now := time.Now()
	var deleted int

	p.instances.Range(func(key, value interface{}) bool {
		instance := value.(*Instance)
		if now.Sub(instance.LastActive) > p.maxIdleTime {
			p.instances.Delete(key)
			deleted++
		}
		return true
	})

	if deleted > 0 {
		slog.Debug("清理过期实例", "count", deleted, "remaining", p.Count())
	}
}

func (p *Pool) Stats() map[string]interface{} {
	return map[string]interface{}{
		"totalInstances": p.Count(),
		"maxIdleTime":    p.maxIdleTime.String(),
	}
}

func MakeKey(sessionID string, agentID string) string {
	if agentID == "" {
		return sessionID
	}
	return fmt.Sprintf("%s:%s", sessionID, agentID)
}
