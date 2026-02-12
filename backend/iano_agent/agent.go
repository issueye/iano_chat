package iano_agent

import (
	"context"
	"fmt"
	"iano_agent/tools"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

type Config struct {
	Tools        []Tool
	Callback     MessageCallback
	MaxRounds    int
	AllowedTools []string
	SystemPrompt string
}

func DefaultConfig() *Config {
	return &Config{
		Tools:        make([]Tool, 0),
		MaxRounds:    50,
		SystemPrompt: "你是一个智能助手。",
	}
}

type Agent struct {
	config       *Config
	ra           *react.Agent
	chatModel    model.ToolCallingChatModel
	mu           sync.RWMutex
	tokenUsage   *TokenUsage
	maxRounds    int
	toolRegistry tools.Registry
}

func NewAgent(chatModel model.ToolCallingChatModel, opts ...Option) (*Agent, error) {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	agent := &Agent{
		config:     cfg,
		maxRounds:  cfg.MaxRounds,
		tokenUsage: &TokenUsage{LastUpdated: time.Now()},
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
		MaxStep:          30, // 增加最大步数，支持更多工具调用和推理轮次
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create react agent: %w", err)
	}

	agent.ra = ra
	agent.chatModel = chatModel

	return agent, nil
}

func (a *Agent) buildSystemPrompt() string {
	var parts []string

	if a.config.SystemPrompt != "" {
		parts = append(parts, a.config.SystemPrompt)
	} else {
		parts = append(parts, "你是一个智能助手。")
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

	// 转换为 BaseTool 类型
	tools := make([]tool.BaseTool, 0, len(bts))
	for _, t := range bts {
		tools = append(tools, t)
	}

	return compose.ToolsNodeConfig{
		Tools: tools,
	}, nil
}

func (a *Agent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.tokenUsage = &TokenUsage{
		LastUpdated: time.Now(),
	}
}

func (a *Agent) GetToolRegistry() tools.Registry {
	return a.toolRegistry
}
