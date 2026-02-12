package iano_agent

import (
	"context"
	"errors"
	"fmt"
	"iano_agent/callback"
	"io"
	"log/slog"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

type ConversationRound struct {
	UserMessage      *schema.Message
	AssistantMessage *schema.Message
	Timestamp        time.Time
	TokenCount       int
}

func (a *Agent) Chat(ctx context.Context, userInput string) (string, error) {
	a.updateLastActive()

	a.mu.Lock()

	if err := a.checkAndSummarize(ctx); err != nil {
		slog.Error("摘要检查失败", slog.String("error", err.Error()))
	}

	if len(a.conversation.RecentRounds) >= a.maxRounds {
		a.mu.Unlock()
		return "", fmt.Errorf("超过最大对话轮数 %d", a.maxRounds)
	}

	userMsg := schema.UserMessage(userInput)
	currentRound := &ConversationRound{
		UserMessage: userMsg,
		Timestamp:   time.Now(),
		TokenCount:  estimateTokens(userInput),
	}

	opts := a.MakeStreamOpts()

	callback := a.config.Callback

	a.mu.Unlock()

	msgReader, err := a.ra.Stream(ctx, []*schema.Message{userMsg}, opts...)
	if err != nil {
		return "", fmt.Errorf("流式对话失败: %w", err)
	}

	var fullResponse string
	var allToolCalls []ToolCallInfo
	isToolCall := false

	for {
		msg, err := msgReader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			slog.Error("读取消息失败", slog.String("error", err.Error()))
			return "", fmt.Errorf("流式对话接收消息失败: %w", err)
		}

		if len(msg.ToolCalls) > 0 {
			isToolCall = true
			for _, tc := range msg.ToolCalls {
				allToolCalls = append(allToolCalls, ToolCallInfo{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				})
			}
		}

		if msg.Content != "" {
			fullResponse += msg.Content

			if callback != nil {
				callback(msg.Content, isToolCall, nil)
			}
		}
	}

	if callback != nil && len(allToolCalls) > 0 {
		callback("", true, allToolCalls)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 构建工具调用信息
	var toolCalls []schema.ToolCall
	for _, tc := range allToolCalls {
		toolCalls = append(toolCalls, schema.ToolCall{
			ID:   tc.ID,
			Type: "function",
			Function: schema.FunctionCall{
				Name:      tc.Name,
				Arguments: tc.Arguments,
			},
		})
	}

	currentRound.AssistantMessage = schema.AssistantMessage(fullResponse, toolCalls)
	currentRound.TokenCount += estimateTokens(fullResponse)
	a.conversation.RecentRounds = append(a.conversation.RecentRounds, currentRound)

	a.tokenUsage.TotalTokens += int64(currentRound.TokenCount)
	a.tokenUsage.CompletionTokens += int64(estimateTokens(fullResponse))
	a.tokenUsage.PromptTokens += int64(estimateTokens(userInput))
	a.lastActiveAt = time.Now()

	return fullResponse, nil
}

func (a *Agent) ChatSimple(ctx context.Context, userInput string) (string, error) {
	return a.Chat(ctx, userInput)
}

func (a *Agent) ChatWithHistory(ctx context.Context, messages []*schema.Message) (string, error) {
	a.updateLastActive()

	a.mu.Lock()
	defer a.mu.Unlock()

	opts := a.MakeStreamOpts()
	msgReader, err := a.ra.Stream(ctx, messages, opts...)
	if err != nil {
		return "", fmt.Errorf("流式对话失败: %w", err)
	}

	var fullResponse string
	var allToolCalls []ToolCallInfo
	for {
		msg, err := msgReader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("流式对话接收消息失败: %w", err)
		}

		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				allToolCalls = append(allToolCalls, ToolCallInfo{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				})
			}
		}

		if msg.Content != "" {
			fullResponse += msg.Content
			if a.config.Callback != nil {
				a.config.Callback(msg.Content, len(msg.ToolCalls) > 0, nil)
			}
		}
	}

	if a.config.Callback != nil && len(allToolCalls) > 0 {
		a.config.Callback("", true, allToolCalls)
	}

	a.lastActiveAt = time.Now()
	return fullResponse, nil
}

func (a *Agent) MakeStreamOpts() []agent.AgentOption {
	return []agent.AgentOption{
		agent.WithComposeOptions(compose.WithCallbacks(&callback.LogCallbackHandler{})),
	}
}

func (a *Agent) AddTool(name string, t tool.BaseTool) error {
	return a.AddToolToRegistry(name, t)
}

func (a *Agent) AddToolToRegistry(name string, t tool.BaseTool) error {
	if name == "" {
		return fmt.Errorf("工具名称不能为空")
	}
	if t == nil {
		return fmt.Errorf("工具实例不能为空")
	}

	if err := a.toolRegistry.Register(name, t); err != nil {
		return fmt.Errorf("注册工具失败: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	toolsConfig, err := a.makeToolsConfig()
	if err != nil {
		return fmt.Errorf("创建工具配置失败: %w", err)
	}

	ctx := context.Background()
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: a.chatModel,
		ToolsConfig:      toolsConfig,
		MessageModifier:  a.messageModifier,
		MaxStep:          30,
	})
	if err != nil {
		return fmt.Errorf("创建代理失败: %w", err)
	}

	a.ra = ra
	return nil
}

func (a *Agent) RemoveTool(name string) error {
	if err := a.toolRegistry.Unregister(name); err != nil {
		return fmt.Errorf("注销工具失败: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	toolsConfig, err := a.makeToolsConfig()
	if err != nil {
		return fmt.Errorf("创建工具配置失败: %w", err)
	}

	ctx := context.Background()
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: a.chatModel,
		ToolsConfig:      toolsConfig,
		MessageModifier:  a.messageModifier,
		MaxStep:          30,
	})
	if err != nil {
		return fmt.Errorf("创建代理失败: %w", err)
	}

	a.ra = ra
	return nil
}

func (a *Agent) ListTools() []string {
	return a.toolRegistry.Names()
}

func (a *Agent) GetConversationInfo() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	totalRounds := a.conversation.SummarizedRounds + len(a.conversation.RecentRounds)

	return map[string]interface{}{
		"totalRounds":      totalRounds,
		"summarizedRounds": a.conversation.SummarizedRounds,
		"recentRounds":     len(a.conversation.RecentRounds),
		"hasSummary":       a.conversation.SummaryContent != "",
		"summaryLength":    len(a.conversation.SummaryContent),
		"sessionId":        a.config.SessionID,
		"agentId":          a.config.AgentID,
	}
}

func (a *Agent) SetCallback(callback MessageCallback) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.Callback = callback
}

func (a *Agent) GetConfig() *Config {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cfg := *a.config
	return &cfg
}

func (a *Agent) GetTokenUsage() *TokenUsage {
	a.mu.RLock()
	defer a.mu.RUnlock()

	usage := *a.tokenUsage
	return &usage
}
