package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"iano_agent/store"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
)

type Manager struct {
	instances sync.Map
	chatModel model.ToolCallingChatModel
	mu        sync.RWMutex
	agents    map[string]*Agent
}

func NewManager(chatModel model.ToolCallingChatModel) *Manager {
	return &Manager{
		chatModel: chatModel,
		agents:    make(map[string]*Agent),
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

	opts := []Option{
		WithAgentID(cfg.ID),
	}
	if cfg.SystemPrompt != "" {
		opts = append(opts, WithSystemPrompt(cfg.SystemPrompt))
	}
	if len(cfg.AllowedTools) > 0 {
		opts = append(opts, WithAllowedTools(cfg.AllowedTools))
	}
	if cfg.MaxRounds > 0 {
		opts = append(opts, WithMaxRounds(cfg.MaxRounds))
	}

	agent, err := NewAgent(m.chatModel, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	instance := &AgentInstance{
		ID:           cfg.ID,
		Name:         cfg.Name,
		Agent:        agent,
		Tools:        cfg.AllowedTools,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	m.instances.Store(cfg.ID, instance)
	m.agents[cfg.ID] = agent

	slog.Info("Agent created", "id", cfg.ID, "name", cfg.Name)
	return instance, nil
}

type AgentInstance struct {
	ID           string
	Name         string
	Agent        *Agent
	Tools        []string
	CreatedAt    time.Time
	LastActiveAt time.Time
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
	delete(m.agents, id)
	slog.Info("Agent deleted", "id", id)
	return nil
}

func (m *Manager) AddToolToAgent(ctx context.Context, agentID string, toolName string, t tool.BaseTool) error {
	instance, exists := m.GetAgent(agentID)
	if !exists {
		return fmt.Errorf("agent with id %s not found", agentID)
	}
	if err := instance.Agent.AddTool(toolName, t); err != nil {
		return fmt.Errorf("failed to add tool to agent: %w", err)
	}
	instance.Tools = append(instance.Tools, toolName)
	slog.Info("Tool added to agent", "agentID", agentID, "toolName", toolName)
	return nil
}

func (m *Manager) RemoveToolFromAgent(agentID string, toolName string) error {
	instance, exists := m.GetAgent(agentID)
	if !exists {
		return fmt.Errorf("agent with id %s not found", agentID)
	}
	if err := instance.Agent.RemoveTool(toolName); err != nil {
		return fmt.Errorf("failed to remove tool from agent: %w", err)
	}
	newTools := make([]string, 0)
	for _, t := range instance.Tools {
		if t != toolName {
			newTools = append(newTools, t)
		}
	}
	instance.Tools = newTools
	slog.Info("Tool removed from agent", "agentID", agentID, "toolName", toolName)
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
	m.agents = make(map[string]*Agent)
}

func (m *Manager) Chat(ctx context.Context, agentID string, message string, callback MessageCallback) (string, error) {
	instance, exists := m.GetAgent(agentID)
	if !exists {
		return "", fmt.Errorf("agent with id %s not found", agentID)
	}
	if callback != nil {
		instance.Agent.SetCallback(callback)
	}
	response, err := instance.Agent.Chat(ctx, message)
	if err != nil {
		return "", err
	}
	instance.LastActiveAt = time.Now()
	return response, nil
}

func (m *Manager) GetAgentInfo(agentID string) (map[string]interface{}, error) {
	instance, exists := m.GetAgent(agentID)
	if !exists {
		return nil, fmt.Errorf("agent with id %s not found", agentID)
	}
	return map[string]interface{}{
		"id":           instance.ID,
		"name":         instance.Name,
		"tools":        instance.Tools,
		"createdAt":    instance.CreatedAt,
		"lastActiveAt": instance.LastActiveAt,
		"conversation": instance.Agent.GetConversationInfo(),
		"tokenUsage":   instance.Agent.GetTokenUsage(),
	}, nil
}

type TaskResult struct {
	AgentID   string
	Response  string
	TokenUsed int64
	Duration  time.Duration
	Error     error
}

type Coordinator struct {
	pool        *Pool
	store       store.ConversationStore
	mainAgentID string
	mu          sync.RWMutex
}

type Pool struct {
	instances sync.Map
	chatModel model.ToolCallingChatModel
	maxIdle   time.Duration
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

func NewCoordinator(chatModel model.ToolCallingChatModel, st store.ConversationStore, cfg *AgentPoolConfig) *Coordinator {
	return &Coordinator{
		pool: &Pool{
			chatModel: chatModel,
			maxIdle:   cfg.MaxIdleTime,
		},
		store: st,
	}
}

func (c *Coordinator) Chat(ctx context.Context, sessionID, agentID, message string, callback MessageCallback) (*TaskResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *Coordinator) ChatWithCallback(ctx context.Context, sessionID, agentID, message string, callback MessageCallback) (*TaskResult, error) {
	return c.Chat(ctx, sessionID, agentID, message, callback)
}

func (c *Coordinator) ClearSession(ctx context.Context, sessionID string) error {
	c.pool.instances.Delete(sessionID)
	if c.store != nil {
		return c.store.Delete(ctx, sessionID)
	}
	return nil
}

func (c *Coordinator) GetAgent(sessionID, agentID string, opts ...Option) (*Agent, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *Coordinator) GetPoolStats() map[string]interface{} {
	return map[string]interface{}{"totalAgents": 0}
}

func (c *Coordinator) Close() {}
