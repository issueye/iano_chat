package agent

import (
	"context"
	"fmt"
	"iano_chat/agent/tools"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// Config Agent 配置
type Config struct {
	Tools    []Tool
	Callback MessageCallback
	Summary  SummaryConfig
}

// ConversationLayer 对话分层存储
type ConversationLayer struct {
	// 完整对话历史（最近N轮）
	RecentRounds []*ConversationRound
	// 历史摘要内容
	SummaryContent string
	// 已摘要的对话轮数
	SummarizedRounds int
}

// Agent AI Agent 封装
type Agent struct {
	config           *Config
	ra               *react.Agent
	chatModel        model.ToolCallingChatModel
	conversation     *ConversationLayer
	mu               sync.RWMutex
	tokenUsage       *TokenUsage
	OutputTokenUsage int
	// 最大对话轮数
	maxRounds int
}

// NewAgent 创建新的 Agent 实例
func NewAgent(chatModel model.ToolCallingChatModel, opts ...Option) (*Agent, error) {
	// 默认配置
	cfg := &Config{
		Tools: make([]Tool, 0),
		Summary: SummaryConfig{
			KeepRecentRounds: 4,   // 默认保留最近4轮
			TriggerThreshold: 8,   // 达到8轮触发摘要
			MaxSummaryTokens: 500, // 摘要最多500 tokens
			Enabled:          true,
		},
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(cfg)
	}

	agent := &Agent{
		config: cfg,
		conversation: &ConversationLayer{
			RecentRounds:     make([]*ConversationRound, 0),
			SummaryContent:   "",
			SummarizedRounds: 0,
		},
		tokenUsage: &TokenUsage{
			LastUpdated: time.Now(),
		},
	}

	// 创建 Tools
	toolsConfig, err := agent.makeToolsConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to make tools config: %w", err)
	}

	// 创建 react agent
	ctx := context.Background()
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      toolsConfig,
		MessageModifier:  agent.messageModifier,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create react agent: %w", err)
	}

	agent.ra = ra
	agent.chatModel = chatModel

	return agent, nil
}

// messageModifier 消息修改器，组装上下文
func (a *Agent) messageModifier(ctx context.Context, input []*schema.Message) []*schema.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*schema.Message, 0)

	// 1. 添加系统提示（如果有摘要，加入摘要信息）
	systemPrompt := a.buildSystemPrompt()
	if systemPrompt != "" {
		result = append(result, schema.SystemMessage(systemPrompt))
	}

	// 2. 添加历史摘要（如果有）
	if a.conversation.SummaryContent != "" {
		summaryMsg := fmt.Sprintf("[历史对话摘要] %s", a.conversation.SummaryContent)
		result = append(result, schema.UserMessage(summaryMsg))
	}

	// 3. 添加最近的完整对话
	for _, round := range a.conversation.RecentRounds {
		if round.UserMessage != nil {
			result = append(result, round.UserMessage)
		}
		if round.AssistantMessage != nil {
			result = append(result, round.AssistantMessage)
		}
	}

	// 4. 添加当前输入
	result = append(result, input...)

	return result
}

// buildSystemPrompt 构建系统提示
func (a *Agent) buildSystemPrompt() string {
	var parts []string

	parts = append(parts, "你是一个智能助手。")

	if a.conversation.SummaryContent != "" {
		parts = append(parts, "以下是之前对话的摘要，供你参考上下文。")
	}

	return strings.Join(parts, "")
}

// makeToolsConfig 创建工具配置
func (a *Agent) makeToolsConfig() (compose.ToolsNodeConfig, error) {
	bts := make([]tool.BaseTool, 0)
	duckDuckGoTool, err := tools.NewDuckDuckGoTool()
	if err != nil {
		return compose.ToolsNodeConfig{}, fmt.Errorf("failed to create duckduckgo tool: %w", err)
	}
	bts = append(bts, duckDuckGoTool)
	// 添加 HTTP 客户端工具
	bts = append(bts, &tools.HTTPClientTool{})

	return compose.ToolsNodeConfig{
		Tools: bts,
	}, nil
}

// ClearHistory 清空对话历史
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

// GetHistory 获取对话历史（包含摘要信息）
func (a *Agent) GetHistory() []*schema.Message {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*schema.Message, 0)

	// 添加摘要标记
	if a.conversation.SummaryContent != "" {
		summaryInfo := fmt.Sprintf("[历史摘要: 已摘要%d轮对话]", a.conversation.SummarizedRounds)
		result = append(result, schema.SystemMessage(summaryInfo))
	}

	// 添加完整对话
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
