package agent

import (
	"context"
	"errors"
	"fmt"
	"iano_chat/agent/tools"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
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

// SummaryConfig 摘要配置
type SummaryConfig struct {
	// 保留的最近对话轮数（一轮 = 用户 + 助手）
	KeepRecentRounds int
	// 触发摘要的对话轮数阈值
	TriggerThreshold int
	// 摘要的最大token数
	MaxSummaryTokens int
	// 是否启用摘要
	Enabled bool
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

// Config Agent 配置
type Config struct {
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float32
	MaxTokens   int
	Tools       []Tool
	Callback    MessageCallback
	Summary     SummaryConfig
}

// Option 配置选项函数类型
type Option func(*Config)

// WithBaseURL 设置 API 基础 URL
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithAPIKey 设置 API Key
func WithAPIKey(key string) Option {
	return func(c *Config) {
		c.APIKey = key
	}
}

// WithModel 设置模型名称
func WithModel(model string) Option {
	return func(c *Config) {
		c.Model = model
	}
}

// WithTemperature 设置温度参数
func WithTemperature(temp float32) Option {
	return func(c *Config) {
		c.Temperature = temp
	}
}

// WithMaxTokens 设置最大 token 数
func WithMaxTokens(tokens int) Option {
	return func(c *Config) {
		c.MaxTokens = tokens
	}
}

// WithTools 设置工具列表
func WithTools(tools []Tool) Option {
	return func(c *Config) {
		c.Tools = tools
	}
}

// WithCallback 设置消息回调
func WithCallback(callback MessageCallback) Option {
	return func(c *Config) {
		c.Callback = callback
	}
}

// WithSummaryConfig 设置摘要配置
func WithSummaryConfig(cfg SummaryConfig) Option {
	return func(c *Config) {
		c.Summary = cfg
	}
}

// ConversationRound 对话轮次
type ConversationRound struct {
	UserMessage      *schema.Message
	AssistantMessage *schema.Message
	Timestamp        time.Time
	TokenCount       int
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
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2048,
		Tools:       make([]Tool, 0),
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
	// 添加 HTTP 客户端工具
	bts = append(bts, &tools.HTTPClientTool{})

	return compose.ToolsNodeConfig{
		Tools: bts,
	}, nil
}

// estimateTokens 估算token数量（简单估算：中文字符算2个token，英文算1个）
func estimateTokens(text string) int {
	tokens := 0
	for _, r := range text {
		if r > 127 {
			tokens += 2
		} else {
			tokens += 1
		}
	}
	return tokens
}

// checkAndSummarize 检查并执行摘要
func (a *Agent) checkAndSummarize(ctx context.Context) error {
	if !a.config.Summary.Enabled {
		return nil
	}

	totalRounds := a.conversation.SummarizedRounds + len(a.conversation.RecentRounds)

	// 检查是否达到触发阈值
	if totalRounds < a.config.Summary.TriggerThreshold {
		return nil
	}

	// 计算需要摘要的轮数
	roundsToSummarize := len(a.conversation.RecentRounds) - a.config.Summary.KeepRecentRounds
	if roundsToSummarize <= 0 {
		return nil
	}

	slog.Info("触发对话摘要",
		slog.Int("totalRounds", totalRounds),
		slog.Int("roundsToSummarize", roundsToSummarize),
		slog.Int("keepRecent", a.config.Summary.KeepRecentRounds))

	return a.summarizeConversation(ctx, roundsToSummarize)
}

// summarizeConversation 执行对话摘要
func (a *Agent) summarizeConversation(ctx context.Context, roundsToSummarize int) error {
	if roundsToSummarize <= 0 || roundsToSummarize > len(a.conversation.RecentRounds) {
		return nil
	}

	// 准备需要摘要的对话内容
	var conversationText strings.Builder
	conversationText.WriteString("请将以下对话内容进行摘要，保留关键信息和上下文：\n\n")

	// 如果有之前的摘要，先包含
	if a.conversation.SummaryContent != "" {
		conversationText.WriteString("[之前的历史摘要]\n")
		conversationText.WriteString(a.conversation.SummaryContent)
		conversationText.WriteString("\n\n")
	}

	// 添加需要摘要的对话
	conversationText.WriteString("[需要摘要的新对话]\n")
	for i := 0; i < roundsToSummarize; i++ {
		round := a.conversation.RecentRounds[i]
		if round.UserMessage != nil {
			conversationText.WriteString(fmt.Sprintf("用户：%s\n", round.UserMessage.Content))
		}
		if round.AssistantMessage != nil {
			conversationText.WriteString(fmt.Sprintf("助手：%s\n", round.AssistantMessage.Content))
		}
		conversationText.WriteString("\n")
	}

	conversationText.WriteString(fmt.Sprintf("\n请生成一个简洁的摘要（不超过%d个token），保留关键信息：",
		a.config.Summary.MaxSummaryTokens))

	// 调用模型生成摘要
	summaryMsg, err := a.chatModel.Generate(ctx, []*schema.Message{
		schema.UserMessage(conversationText.String()),
	})
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	summary := summaryMsg.Content
	if len(summary) > a.config.Summary.MaxSummaryTokens*4 { // 粗略限制长度
		summary = summary[:a.config.Summary.MaxSummaryTokens*4] + "..."
	}

	// 计算节省的token
	oldTokens := 0
	for i := 0; i < roundsToSummarize; i++ {
		round := a.conversation.RecentRounds[i]
		if round.UserMessage != nil {
			oldTokens += estimateTokens(round.UserMessage.Content)
		}
		if round.AssistantMessage != nil {
			oldTokens += estimateTokens(round.AssistantMessage.Content)
		}
	}
	summaryTokens := estimateTokens(summary)
	savedTokens := oldTokens - summaryTokens

	// 更新对话层
	a.conversation.SummaryContent = summary
	a.conversation.SummarizedRounds += roundsToSummarize

	// 保留最近的对话
	newRecent := make([]*ConversationRound, 0)
	for i := roundsToSummarize; i < len(a.conversation.RecentRounds); i++ {
		newRecent = append(newRecent, a.conversation.RecentRounds[i])
	}
	a.conversation.RecentRounds = newRecent

	// 更新token统计
	a.tokenUsage.SummaryTokens += int64(summaryTokens)
	a.tokenUsage.SavedTokens += int64(savedTokens)
	a.tokenUsage.LastUpdated = time.Now()

	slog.Info("对话摘要完成",
		slog.Int("summarizedRounds", roundsToSummarize),
		slog.Int("oldTokens", oldTokens),
		slog.Int("summaryTokens", summaryTokens),
		slog.Int("savedTokens", savedTokens),
		slog.Int("remainingRounds", len(a.conversation.RecentRounds)))

	return nil
}

// Chat 执行对话（流式）
func (a *Agent) Chat(ctx context.Context, userInput string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 检查是否需要摘要
	if err := a.checkAndSummarize(ctx); err != nil {
		slog.Error("摘要检查失败", slog.String("error", err.Error()))
	}

	// 添加用户消息到当前轮
	userMsg := schema.UserMessage(userInput)
	currentRound := &ConversationRound{
		UserMessage: userMsg,
		Timestamp:   time.Now(),
		TokenCount:  estimateTokens(userInput),
	}

	opts := a.MakeStreamOpts()

	// 执行对话
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

			// 调用回调函数
			if a.config.Callback != nil {
				a.config.Callback(msg.Content, isToolCall)
			}
		}

		// 检查是否超过最大对话轮数
		if len(a.conversation.RecentRounds) > a.maxRounds {
			return "", fmt.Errorf("超过最大对话轮数 %d", a.maxRounds)
		}
	}

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
func (a *Agent) AddTool(t Tool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.Tools = append(a.config.Tools, t)

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

// ForceSummarize 强制触发摘要
func (a *Agent) ForceSummarize(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.conversation.RecentRounds) <= a.config.Summary.KeepRecentRounds {
		return fmt.Errorf("对话轮数不足，无法触发摘要，至少需要 %d 轮",
			a.config.Summary.KeepRecentRounds+1)
	}

	roundsToSummarize := len(a.conversation.RecentRounds) - a.config.Summary.KeepRecentRounds
	return a.summarizeConversation(ctx, roundsToSummarize)
}

// GetSummary 获取当前摘要内容
func (a *Agent) GetSummary() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.conversation.SummaryContent
}
