package pool

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"iano_agent/store"

	"github.com/cloudwego/eino/components/model"
)

type AgentInstance struct {
	ID           string
	Name         string
	Agent        interface{}
	Tools        []string
	CreatedAt    time.Time
	LastActiveAt time.Time
}

type Manager struct {
	instances sync.Map
	chatModel model.ToolCallingChatModel
	mu        sync.RWMutex
}

func NewManager(chatModel model.ToolCallingChatModel) *Manager {
	return &Manager{
		chatModel: chatModel,
	}
}

type CreateAgentConfig struct {
	ID           string
	Name         string
	SystemPrompt string
	AllowedTools []string
	MaxRounds    int
}

func (m *Manager) CreateAgent(ctx context.Context, cfg *CreateAgentConfig) (*AgentInstance, error) {
	if cfg.ID == "" {
		return nil, fmt.Errorf("agent id is required")
	}

	if _, exists := m.instances.Load(cfg.ID); exists {
		return nil, fmt.Errorf("agent with id %s already exists", cfg.ID)
	}

	instance := &AgentInstance{
		ID:           cfg.ID,
		Name:         cfg.Name,
		Tools:        cfg.AllowedTools,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	m.instances.Store(cfg.ID, instance)
	slog.Info("Agent created", "id", cfg.ID, "name", cfg.Name)
	return instance, nil
}

func (m *Manager) GetAgent(id string) (*AgentInstance, bool) {
	if v, ok := m.instances.Load(id); ok {
		instance := v.(*AgentInstance)
		instance.LastActiveAt = time.Now()
		return instance, true
	}
	return nil, false
}

func (m *Manager) DeleteAgent(id string) error {
	if _, exists := m.instances.Load(id); !exists {
		return fmt.Errorf("agent with id %s not found", id)
	}
	m.instances.Delete(id)
	slog.Info("Agent deleted", "id", id)
	return nil
}

func (m *Manager) ListAgents() []*AgentInstance {
	var instances []*AgentInstance
	m.instances.Range(func(key, value interface{}) bool {
		instance := value.(*AgentInstance)
		instances = append(instances, instance)
		return true
	})
	return instances
}

func (m *Manager) Count() int {
	count := 0
	m.instances.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (m *Manager) Clear() {
	m.instances.Range(func(key, value interface{}) bool {
		m.instances.Delete(key)
		return true
	})
}

func (m *Manager) GetChatModel() model.ToolCallingChatModel {
	return m.chatModel
}

type Coordinator struct {
	pool        *Pool
	store       store.ConversationStore
	mainAgentID string
	mu          sync.RWMutex
}

func NewCoordinator(chatModel model.ToolCallingChatModel, st store.ConversationStore, poolConfig *PoolConfig) *Coordinator {
	p := NewPool(chatModel, poolConfig)
	return &Coordinator{
		pool:  p,
		store: st,
	}
}

type TaskResult struct {
	AgentID   string
	Response  string
	TokenUsed int64
	Duration  time.Duration
	Error     error
}

func (c *Coordinator) GetPool() *Pool {
	return c.pool
}

func (c *Coordinator) GetStore() store.ConversationStore {
	return c.store
}

func (c *Coordinator) ClearSession(ctx context.Context, sessionID string) error {
	c.pool.Delete(sessionID)
	if c.store != nil {
		return c.store.Delete(ctx, sessionID)
	}
	return nil
}

func (c *Coordinator) GetPoolStats() map[string]interface{} {
	return c.pool.Stats()
}

func (c *Coordinator) Close() {
	c.pool.Close()
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
