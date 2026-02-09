package agent

import (
	"context"
	"errors"
	"fmt"
	"iano_chat/agent/tools"
	"io"
	"log/slog"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// MessageCallback 消息回调函数类型
type MessageCallback func(content string, isToolCall bool)

// Tool 工具定义
type Tool struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, params map[string]interface{}) (string, error)
}

// TokenUsage Token使用统计
type TokenUsage struct {
	TotalTokens      int64
	PromptTokens     int64
	CompletionTokens int64
	SummaryTokens    int64
	SavedTokens      int64
	LastUpdated      time.Time
}

// ConversationRound 对话轮次
type ConversationRound struct {
	UserMessage      *schema.Message
	AssistantMessage *schema.Message
	Timestamp        time.Time
	TokenCount       int
}

// Chat 执行对话（流式）
func (a *Agent) Chat(ctx context.Context, userInput string) (string, error) {
	// 第一阶段：准备数据（加锁）
	a.mu.Lock()

	// 检查是否需要摘要
	if err := a.checkAndSummarize(ctx); err != nil {
		slog.Error("摘要检查失败", slog.String("error", err.Error()))
	}

	// 检查是否超过最大对话轮数
	if len(a.conversation.RecentRounds) >= a.maxRounds {
		a.mu.Unlock()
		return "", fmt.Errorf("超过最大对话轮数 %d", a.maxRounds)
	}

	// 添加用户消息到当前轮
	userMsg := schema.UserMessage(userInput)
	currentRound := &ConversationRound{
		UserMessage: userMsg,
		Timestamp:   time.Now(),
		TokenCount:  estimateTokens(userInput),
	}

	opts := a.MakeStreamOpts()

	// 获取回调函数的副本，避免在流式处理期间持有锁
	callback := a.config.Callback

	a.mu.Unlock()

	// 第二阶段：执行流式对话（无锁）
	msgReader, err := a.ra.Stream(ctx, []*schema.Message{userMsg}, opts...)
	if err != nil {
		return "", fmt.Errorf("流式对话失败: %w", err)
	}

	var fullResponse string
	isToolCall := false

	// 读取流式响应
	for {
		msg, err := msgReader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			slog.Error("读取消息失败", slog.String("error", err.Error()))
			return "", fmt.Errorf("流式对话接收消息失败: %w", err)
		}

		// 检查是否是工具调用
		if len(msg.ToolCalls) > 0 {
			isToolCall = true
		}

		// 累积响应内容
		if msg.Content != "" {
			fullResponse += msg.Content

			// 调用回调函数（使用副本，无需锁）
			if callback != nil {
				callback(msg.Content, isToolCall)
			}
		}
	}

	// 第三阶段：更新状态（加锁）
	a.mu.Lock()
	defer a.mu.Unlock()

	// 完成当前轮
	currentRound.AssistantMessage = schema.AssistantMessage(fullResponse, nil)
	currentRound.TokenCount += estimateTokens(fullResponse)
	a.conversation.RecentRounds = append(a.conversation.RecentRounds, currentRound)

	// 更新token使用统计
	a.tokenUsage.TotalTokens += int64(currentRound.TokenCount)
	a.tokenUsage.CompletionTokens += int64(estimateTokens(fullResponse))
	a.tokenUsage.PromptTokens += int64(estimateTokens(userInput))

	return fullResponse, nil
}

// ChatSimple 简单对话（非流式）
func (a *Agent) ChatSimple(ctx context.Context, userInput string) (string, error) {
	return a.Chat(ctx, userInput)
}

// ChatWithHistory 带历史记录的对话
func (a *Agent) ChatWithHistory(ctx context.Context, messages []*schema.Message) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 配置回调函数
	opts := a.MakeStreamOpts()
	msgReader, err := a.ra.Stream(ctx, messages, opts...)
	if err != nil {
		return "", fmt.Errorf("流式对话失败: %w", err)
	}

	var fullResponse string
	for {
		msg, err := msgReader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("流式对话接收消息失败: %w", err)
		}

		if msg.Content != "" {
			fullResponse += msg.Content
			if a.config.Callback != nil {
				a.config.Callback(msg.Content, len(msg.ToolCalls) > 0)
			}
		}
	}

	return fullResponse, nil
}

func (a *Agent) MakeStreamOpts() []agent.AgentOption {
	return []agent.AgentOption{
		agent.WithComposeOptions(compose.WithCallbacks(&LogCallbackHandler{})),
	}
}

// AddTool 动态添加工具
func (a *Agent) AddTool(name string, t tool.BaseTool) error {
	return a.AddToolToRegistry(name, t)
}

// AddToolToRegistry 添加工具到注册表并刷新 Agent
func (a *Agent) AddToolToRegistry(name string, t tool.BaseTool) error {
	if name == "" {
		return fmt.Errorf("工具名称不能为空")
	}
	if t == nil {
		return fmt.Errorf("工具实例不能为空")
	}

	// 注册到全局注册表
	if err := tools.GlobalRegistry.Register(name, t); err != nil {
		return fmt.Errorf("注册工具失败: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 重新创建 agent 以应用新工具
	toolsConfig, err := a.makeToolsConfig()
	if err != nil {
		return fmt.Errorf("创建工具配置失败: %w", err)
	}

	ctx := context.Background()
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: a.chatModel,
		ToolsConfig:      toolsConfig,
		MessageModifier:  a.messageModifier,
	})
	if err != nil {
		return fmt.Errorf("创建代理失败: %w", err)
	}

	a.ra = ra
	return nil
}

// RemoveTool 从注册表移除工具
func (a *Agent) RemoveTool(name string) error {
	if err := tools.GlobalRegistry.Unregister(name); err != nil {
		return fmt.Errorf("注销工具失败: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 重新创建 agent 以应用变更
	toolsConfig, err := a.makeToolsConfig()
	if err != nil {
		return fmt.Errorf("创建工具配置失败: %w", err)
	}

	ctx := context.Background()
	ra, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: a.chatModel,
		ToolsConfig:      toolsConfig,
		MessageModifier:  a.messageModifier,
	})
	if err != nil {
		return fmt.Errorf("创建代理失败: %w", err)
	}

	a.ra = ra
	return nil
}

// ListTools 列出所有已注册的工具
func (a *Agent) ListTools() []string {
	return tools.GlobalRegistry.Names()
}

// GetConversationInfo 获取对话信息统计
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
	}
}

// func (a *Agent)

// SetCallback 设置消息回调
func (a *Agent) SetCallback(callback MessageCallback) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.Callback = callback
}

// GetConfig 获取当前配置
func (a *Agent) GetConfig() *Config {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cfg := *a.config
	return &cfg
}

// GetTokenUsage 获取Token使用统计
func (a *Agent) GetTokenUsage() *TokenUsage {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 返回副本
	usage := *a.tokenUsage
	return &usage
}
