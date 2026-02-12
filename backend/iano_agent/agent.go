package iano_agent

import (
	"context"
	"fmt"
	"iano_agent/tools"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type Config struct {
	Tools        []Tool
	Callback     MessageCallback
	Summary      SummaryConfig
	MaxRounds    int
	AllowedTools []string
	SessionID    string
	AgentID      string
	SystemPrompt string
}

func DefaultConfig() *Config {
	return &Config{
		Tools: make([]Tool, 0),
		Summary: SummaryConfig{
			KeepRecentRounds: 4,
			TriggerThreshold: 8,
			MaxSummaryTokens: 500,
			Enabled:          true,
		},
		MaxRounds:    50,
		SystemPrompt: "你是一个智能助手。",
	}
}

type ConversationLayer struct {
	RecentRounds     []*ConversationRound
	SummaryContent   string
	SummarizedRounds int
}

type Agent struct {
	config       *Config
	ra           *react.Agent
	chatModel    model.ToolCallingChatModel
	conversation *ConversationLayer
	mu           sync.RWMutex
	tokenUsage   *TokenUsage
	maxRounds    int
	toolRegistry tools.Registry
	createdAt    time.Time
	lastActiveAt time.Time
}

func NewAgent(chatModel model.ToolCallingChatModel, opts ...Option) (*Agent, error) {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	agent := &Agent{
		config:    cfg,
		maxRounds: cfg.MaxRounds,
		conversation: &ConversationLayer{
			RecentRounds:     make([]*ConversationRound, 0),
			SummaryContent:   "",
			SummarizedRounds: 0,
		},
		tokenUsage: &TokenUsage{
			LastUpdated: time.Now(),
		},
		createdAt:    time.Now(),
		lastActiveAt: time.Now(),
	}

	agent.toolRegistry = tools.NewScopedRegistry(tools.GlobalRegistry, cfg.AllowedTools)

	toolsConfig, err := agent.makeToolsConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to make tools config: %w", err)
	}

	ctx := context.Background()
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      toolsConfig,
		MessageModifier:  agent.messageModifier,
		MaxStep:          30, // 增加最大步数，支持更多工具调用和推理轮次
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create react agent: %w", err)
	}

	agent.ra = ra
	agent.chatModel = chatModel

	return agent, nil
}

func (a *Agent) messageModifier(ctx context.Context, input []*schema.Message) []*schema.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*schema.Message, 0)

	systemPrompt := a.buildSystemPrompt()
	if systemPrompt != "" {
		result = append(result, schema.SystemMessage(systemPrompt))
	}

	if a.conversation.SummaryContent != "" {
		summaryMsg := fmt.Sprintf("[历史对话摘要] %s", a.conversation.SummaryContent)
		result = append(result, schema.UserMessage(summaryMsg))
	}

	for _, round := range a.conversation.RecentRounds {
		if round.UserMessage != nil {
			result = append(result, round.UserMessage)
		}
		if round.AssistantMessage != nil {
			result = append(result, round.AssistantMessage)
		}
	}

	result = append(result, input...)

	return result
}

func (a *Agent) buildSystemPrompt() string {
	var parts []string

	if a.config.SystemPrompt != "" {
		parts = append(parts, a.config.SystemPrompt)
	} else {
		parts = append(parts, "你是一个智能助手。")
	}

	if a.conversation.SummaryContent != "" {
		parts = append(parts, "以下是之前对话的摘要，供你参考上下文。")
	}

	return strings.Join(parts, "")
}

func (a *Agent) makeToolsConfig() (compose.ToolsNodeConfig, error) {
	bts := a.toolRegistry.List()

	if len(bts) == 0 {
		ctx := context.Background()
		if err := tools.RegisterBuiltinTools(ctx); err != nil {
			return compose.ToolsNodeConfig{}, fmt.Errorf("注册内置工具失败: %w", err)
		}
		bts = a.toolRegistry.List()
	}

	return compose.ToolsNodeConfig{
		Tools: bts,
	}, nil
}

func (a *Agent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.conversation = &ConversationLayer{
		RecentRounds:     make([]*ConversationRound, 0),
		SummaryContent:   "",
		SummarizedRounds: 0,
	}
	a.tokenUsage = &TokenUsage{
		LastUpdated: time.Now(),
	}
}

func (a *Agent) GetHistory() []*schema.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*schema.Message, 0)

	if a.conversation.SummaryContent != "" {
		summaryInfo := fmt.Sprintf("[历史摘要: 已摘要%d轮对话]", a.conversation.SummarizedRounds)
		result = append(result, schema.SystemMessage(summaryInfo))
	}

	for _, round := range a.conversation.RecentRounds {
		if round.UserMessage != nil {
			result = append(result, round.UserMessage)
		}
		if round.AssistantMessage != nil {
			result = append(result, round.AssistantMessage)
		}
	}

	return result
}

func (a *Agent) GetToolRegistry() tools.Registry {
	return a.toolRegistry
}

func (a *Agent) GetSessionID() string {
	return a.config.SessionID
}

func (a *Agent) GetAgentID() string {
	return a.config.AgentID
}

func (a *Agent) GetCreatedAt() time.Time {
	return a.createdAt
}

func (a *Agent) GetLastActiveAt() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastActiveAt
}

func (a *Agent) updateLastActive() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastActiveAt = time.Now()
}

func (a *Agent) RestoreConversation(layer *ConversationLayer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.conversation = layer
}

func (a *Agent) GetConversationLayer() *ConversationLayer {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.conversation
}

// LoadConversationHistory 加载对话历史
func (a *Agent) LoadConversationHistory(ctx context.Context, rounds []*ConversationRound) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 清空现有历史
	a.conversation.RecentRounds = make([]*ConversationRound, 0)
	a.conversation.SummaryContent = ""
	a.conversation.SummarizedRounds = 0

	// 加载新的历史记录
	for _, round := range rounds {
		if round != nil && round.UserMessage != nil {
			a.conversation.RecentRounds = append(a.conversation.RecentRounds, round)
		}
	}
}
