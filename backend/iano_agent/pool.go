package agent

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
)

type AgentPool struct {
	pools         sync.Map
	chatModel     model.ToolCallingChatModel
	mu            sync.RWMutex
	maxIdleTime   time.Duration
	cleanupTicker *time.Ticker
	done          chan struct{}
}

type AgentPoolConfig struct {
	MaxIdleTime     time.Duration
	CleanupInterval time.Duration
}

func DefaultAgentPoolConfig() *AgentPoolConfig {
	return &AgentPoolConfig{
		MaxIdleTime:     30 * time.Minute,
		CleanupInterval: 5 * time.Minute,
	}
}

func NewAgentPool(chatModel model.ToolCallingChatModel, cfg *AgentPoolConfig) *AgentPool {
	if cfg == nil {
		cfg = DefaultAgentPoolConfig()
	}

	pool := &AgentPool{
		chatModel:   chatModel,
		maxIdleTime: cfg.MaxIdleTime,
		done:        make(chan struct{}),
	}

	pool.cleanupTicker = time.NewTicker(cfg.CleanupInterval)
	go pool.cleanupLoop()

	return pool
}

type poolEntry struct {
	agent      *Agent
	lastActive time.Time
}

func (p *AgentPool) Get(sessionID string, agentID string, opts ...Option) (*Agent, error) {
	key := p.makeKey(sessionID, agentID)

	if entry, ok := p.pools.Load(key); ok {
		pe := entry.(*poolEntry)
		pe.lastActive = time.Now()
		return pe.agent, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if entry, ok := p.pools.Load(key); ok {
		pe := entry.(*poolEntry)
		pe.lastActive = time.Now()
		return pe.agent, nil
	}

	allOpts := append([]Option{
		WithSessionID(sessionID),
		WithAgentID(agentID),
	}, opts...)

	agent, err := NewAgent(p.chatModel, allOpts...)
	if err != nil {
		return nil, fmt.Errorf("创建 Agent 失败: %w", err)
	}

	p.pools.Store(key, &poolEntry{
		agent:      agent,
		lastActive: time.Now(),
	})

	slog.Debug("创建新 Agent", "sessionID", sessionID, "agentID", agentID)
	return agent, nil
}

func (p *AgentPool) GetOrCreate(sessionID string, agentID string, opts ...Option) (*Agent, error) {
	return p.Get(sessionID, agentID, opts...)
}

func (p *AgentPool) Remove(sessionID string, agentID string) {
	key := p.makeKey(sessionID, agentID)
	p.pools.Delete(key)
}

func (p *AgentPool) Clear() {
	p.pools.Range(func(key, value interface{}) bool {
		p.pools.Delete(key)
		return true
	})
}

func (p *AgentPool) Count() int {
	count := 0
	p.pools.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (p *AgentPool) Close() {
	close(p.done)
	p.cleanupTicker.Stop()
	p.Clear()
}

func (p *AgentPool) makeKey(sessionID string, agentID string) string {
	if agentID == "" {
		return sessionID
	}
	return fmt.Sprintf("%s:%s", sessionID, agentID)
}

func (p *AgentPool) cleanupLoop() {
	for {
		select {
		case <-p.done:
			return
		case <-p.cleanupTicker.C:
			p.cleanup()
		}
	}
}

func (p *AgentPool) cleanup() {
	now := time.Now()
	var deleted int

	p.pools.Range(func(key, value interface{}) bool {
		pe := value.(*poolEntry)
		if now.Sub(pe.lastActive) > p.maxIdleTime {
			p.pools.Delete(key)
			deleted++
		}
		return true
	})

	if deleted > 0 {
		slog.Debug("清理过期 Agent", "count", deleted, "remaining", p.Count())
	}
}

func (p *AgentPool) Stats() map[string]interface{} {
	return map[string]interface{}{
		"totalAgents": p.Count(),
		"maxIdleTime": p.maxIdleTime.String(),
	}
}
