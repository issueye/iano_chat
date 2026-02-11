package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
)

type AgentInstance struct {
	ID           string
	Name         string
	Agent        *Agent
	Model        model.ToolCallingChatModel
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
		Model:        m.chatModel,
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
}

func (m *Manager) Chat(ctx context.Context, agentID string, userInput string, callback MessageCallback) (string, error) {
	instance, exists := m.GetAgent(agentID)
	if !exists {
		return "", fmt.Errorf("agent with id %s not found", agentID)
	}

	if callback != nil {
		instance.Agent.SetCallback(callback)
	}

	response, err := instance.Agent.Chat(ctx, userInput)
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
