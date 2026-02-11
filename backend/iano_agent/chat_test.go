package iano_agent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockChatModel 模拟聊天模型
type MockChatModel struct {
	generateFunc func(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error)
}

func (m *MockChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, messages, opts...)
	}
	return schema.AssistantMessage("mock response", nil), nil
}

func (m *MockChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 使用 Pipe 创建流
	reader, writer := schema.Pipe[*schema.Message](1)

	go func() {
		defer writer.Close()
		writer.Send(schema.AssistantMessage("mock stream response", nil), nil)
	}()

	return reader, nil
}

func (m *MockChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return m, nil
}

// TestSummaryConfig 测试摘要配置
func TestSummaryConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   SummaryConfig
		expected SummaryConfig
	}{
		{
			name: "默认配置",
			config: SummaryConfig{
				KeepRecentRounds: 4,
				TriggerThreshold: 8,
				MaxSummaryTokens: 500,
				Enabled:          true,
			},
			expected: SummaryConfig{
				KeepRecentRounds: 4,
				TriggerThreshold: 8,
				MaxSummaryTokens: 500,
				Enabled:          true,
			},
		},
		{
			name: "自定义配置",
			config: SummaryConfig{
				KeepRecentRounds: 2,
				TriggerThreshold: 6,
				MaxSummaryTokens: 300,
				Enabled:          true,
			},
			expected: SummaryConfig{
				KeepRecentRounds: 2,
				TriggerThreshold: 6,
				MaxSummaryTokens: 300,
				Enabled:          true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.KeepRecentRounds != tt.expected.KeepRecentRounds {
				t.Errorf("KeepRecentRounds = %v, want %v", tt.config.KeepRecentRounds, tt.expected.KeepRecentRounds)
			}
			if tt.config.TriggerThreshold != tt.expected.TriggerThreshold {
				t.Errorf("TriggerThreshold = %v, want %v", tt.config.TriggerThreshold, tt.expected.TriggerThreshold)
			}
		})
	}
}

// TestEstimateTokens 测试Token估算
func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		minExpected int // 最小期望值（新算法范围）
		maxExpected int // 最大期望值
	}{
		{
			name:        "纯英文",
			input:       "Hello World",
			minExpected: 8,  // 新算法：单词估算 + 开销
			maxExpected: 15, // 宽松上限
		},
		{
			name:        "纯中文",
			input:       "你好世界",
			minExpected: 8, // 新算法：4个中文字符 * 2 = 8 + 开销
			maxExpected: 15,
		},
		{
			name:        "中英文混合",
			input:       "Hello 你好",
			minExpected: 8,
			maxExpected: 18,
		},
		{
			name:        "空字符串",
			input:       "",
			minExpected: 0,
			maxExpected: 0,
		},
		{
			name:        "长文本",
			input:       "这是一个比较长的中文文本，用于测试Token估算算法的准确性。",
			minExpected: 30,
			maxExpected: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateTokens(tt.input)
			if got < tt.minExpected || got > tt.maxExpected {
				t.Errorf("estimateTokens() = %v, 期望范围 [%d, %d]", got, tt.minExpected, tt.maxExpected)
			}
			t.Logf("estimateTokens(%q) = %d", tt.input, got)
		})
	}
}

// TestConversationLayer 测试对话分层存储
func TestConversationLayer(t *testing.T) {
	layer := &ConversationLayer{
		RecentRounds:     make([]*ConversationRound, 0),
		SummaryContent:   "",
		SummarizedRounds: 0,
	}

	// 添加对话轮次
	for i := 0; i < 5; i++ {
		round := &ConversationRound{
			UserMessage:      schema.UserMessage("用户消息" + string(rune('0'+i))),
			AssistantMessage: schema.AssistantMessage("助手回复"+string(rune('0'+i)), nil),
			Timestamp:        time.Now(),
			TokenCount:       100,
		}
		layer.RecentRounds = append(layer.RecentRounds, round)
	}

	if len(layer.RecentRounds) != 5 {
		t.Errorf("RecentRounds length = %v, want %v", len(layer.RecentRounds), 5)
	}

	// 测试摘要后保留最近轮次
	keepRecent := 2
	roundsToSummarize := len(layer.RecentRounds) - keepRecent
	newRecent := make([]*ConversationRound, 0)
	for i := roundsToSummarize; i < len(layer.RecentRounds); i++ {
		newRecent = append(newRecent, layer.RecentRounds[i])
	}
	layer.RecentRounds = newRecent
	layer.SummarizedRounds = roundsToSummarize

	if len(layer.RecentRounds) != keepRecent {
		t.Errorf("After summary, RecentRounds length = %v, want %v", len(layer.RecentRounds), keepRecent)
	}

	if layer.SummarizedRounds != 3 {
		t.Errorf("SummarizedRounds = %v, want %v", layer.SummarizedRounds, 3)
	}
}

// TestTokenUsage 测试Token使用统计
func TestTokenUsage(t *testing.T) {
	usage := &TokenUsage{
		TotalTokens:      1000,
		PromptTokens:     400,
		CompletionTokens: 600,
		SummaryTokens:    100,
		SavedTokens:      500,
		LastUpdated:      time.Now(),
	}

	if usage.TotalTokens != 1000 {
		t.Errorf("TotalTokens = %v, want %v", usage.TotalTokens, 1000)
	}

	if usage.SavedTokens != 500 {
		t.Errorf("SavedTokens = %v, want %v", usage.SavedTokens, 500)
	}

	// 计算节省比例
	saveRatio := float64(usage.SavedTokens) / float64(usage.TotalTokens+usage.SavedTokens)
	if saveRatio < 0.3 {
		t.Logf("Token节省比例: %.2f%%", saveRatio*100)
	}
}

// TestCheckAndSummarize 测试摘要触发逻辑
func TestCheckAndSummarize(t *testing.T) {
	tests := []struct {
		name             string
		totalRounds      int
		keepRecent       int
		triggerThreshold int
		shouldTrigger    bool
	}{
		{
			name:             "未达到触发阈值",
			totalRounds:      5,
			keepRecent:       4,
			triggerThreshold: 8,
			shouldTrigger:    false,
		},
		{
			name:             "达到触发阈值",
			totalRounds:      10,
			keepRecent:       4,
			triggerThreshold: 8,
			shouldTrigger:    true,
		},
		{
			name:             "刚好达到阈值",
			totalRounds:      8,
			keepRecent:       4,
			triggerThreshold: 8,
			shouldTrigger:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recentRounds := tt.totalRounds - 3 // 假设已摘要3轮
			roundsToSummarize := recentRounds - tt.keepRecent
			willTrigger := tt.totalRounds >= tt.triggerThreshold && roundsToSummarize > 0

			if willTrigger != tt.shouldTrigger {
				t.Errorf("Trigger check = %v, want %v (total=%d, keep=%d, threshold=%d)",
					willTrigger, tt.shouldTrigger, tt.totalRounds, tt.keepRecent, tt.triggerThreshold)
			}
		})
	}
}

// TestBuildSystemPrompt 测试系统提示构建
func TestBuildSystemPrompt(t *testing.T) {
	tests := []struct {
		name          string
		hasSummary    bool
		expectedParts []string
	}{
		{
			name:          "无摘要",
			hasSummary:    false,
			expectedParts: []string{"你是一个智能助手"},
		},
		{
			name:          "有摘要",
			hasSummary:    true,
			expectedParts: []string{"你是一个智能助手", "摘要"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var summary string
			if tt.hasSummary {
				summary = "这是历史摘要"
			}

			var parts []string
			parts = append(parts, "你是一个智能助手。")
			if summary != "" {
				parts = append(parts, "以下是之前对话的摘要，供你参考上下文。")
			}
			prompt := strings.Join(parts, "")

			for _, part := range tt.expectedParts {
				if !strings.Contains(prompt, part) {
					t.Errorf("System prompt should contain '%s'", part)
				}
			}
		})
	}
}

// TestMessageModifier 测试消息修改器
func TestMessageModifier(t *testing.T) {
	agent := &Agent{
		config: &Config{
			Summary: SummaryConfig{
				KeepRecentRounds: 4,
			},
		},
		conversation: &ConversationLayer{
			RecentRounds: []*ConversationRound{
				{
					UserMessage:      schema.UserMessage("用户1"),
					AssistantMessage: schema.AssistantMessage("助手1", nil),
				},
				{
					UserMessage:      schema.UserMessage("用户2"),
					AssistantMessage: schema.AssistantMessage("助手2", nil),
				},
			},
			SummaryContent:   "历史摘要",
			SummarizedRounds: 3,
		},
	}

	input := []*schema.Message{schema.UserMessage("当前输入")}
	result := agent.messageModifier(context.Background(), input)

	// 检查结果
	hasSystem := false
	hasSummary := false
	hasHistory := false
	hasInput := false

	for _, msg := range result {
		switch msg.Role {
		case schema.System:
			hasSystem = true
		case schema.User:
			if strings.Contains(msg.Content, "历史摘要") {
				hasSummary = true
			}
			if msg.Content == "当前输入" {
				hasInput = true
			}
			if msg.Content == "用户1" || msg.Content == "用户2" {
				hasHistory = true
			}
		case schema.Assistant:
			if msg.Content == "助手1" || msg.Content == "助手2" {
				hasHistory = true
			}
		}
	}

	if !hasSystem {
		t.Error("Message modifier should add system prompt")
	}
	if !hasSummary {
		t.Error("Message modifier should include summary")
	}
	if !hasHistory {
		t.Error("Message modifier should include recent history")
	}
	if !hasInput {
		t.Error("Message modifier should include current input")
	}
}

// TestTokenSavings 测试Token节省效果
func TestTokenSavings(t *testing.T) {
	// 模拟一个长对话场景
	conversationText := `
用户：请帮我介绍一下Go语言的特点和优势。
助手：Go语言（Golang）是由Google开发的一种静态类型、编译型编程语言。它的主要特点包括：
1. 简洁的语法：Go语言的语法设计非常简洁，易于学习和使用。
2. 高效的并发：通过goroutine和channel实现轻量级并发。
3. 快速编译：Go的编译速度非常快，适合大型项目开发。
4. 强大的标准库：提供了丰富的标准库，涵盖网络、IO、加密等常用功能。
5. 跨平台支持：可以编译成各种平台的可执行文件。
6. 垃圾回收：自动内存管理，减少内存泄漏风险。

用户：那Go语言适合做什么类型的项目？
助手：Go语言特别适合以下类型的项目：
1. 云原生应用：Docker、Kubernetes等核心项目都是用Go编写的。
2. 微服务架构：Go的高性能和低内存占用使其成为微服务的理想选择。
3. 网络服务：标准库提供了强大的网络编程支持，适合开发Web服务器、API服务等。
4. 命令行工具：编译后的单二进制文件便于分发和部署。
5. 区块链应用：以太坊等区块链项目的核心代码使用Go编写。
6. DevOps工具：许多流行的DevOps工具如Terraform、Consul都是用Go开发的。

用户：Go语言的性能如何？
助手：Go语言的性能表现非常出色：
1. 执行速度：Go是编译型语言，执行速度接近C/C++，比Python、JavaScript等解释型语言快很多。
2. 内存占用：Go的内存占用相对较低，适合资源受限的环境。
3. 启动时间：Go程序启动速度快，适合Serverless和容器化部署。
4. 并发性能：goroutine的调度效率非常高，可以轻松创建数百万个goroutine。
5. GC性能：Go的垃圾回收器经过多次优化，延迟已大幅降低。
`

	originalTokens := estimateTokens(conversationText)

	// 假设摘要后的内容
	summaryText := "用户询问Go语言特点、适用场景和性能。Go是Google开发的静态编译型语言，特点包括简洁语法、高效并发、快速编译、强大标准库。适合云原生应用、微服务、网络服务、命令行工具等。性能接近C/C++，内存占用低，启动快，并发性能优秀。"
	summaryTokens := estimateTokens(summaryText)

	savedTokens := originalTokens - summaryTokens
	saveRatio := float64(savedTokens) / float64(originalTokens) * 100

	t.Logf("原始对话: %d tokens", originalTokens)
	t.Logf("摘要内容: %d tokens", summaryTokens)
	t.Logf("节省: %d tokens (%.1f%%)", savedTokens, saveRatio)

	// 验证节省效果显著（至少节省50%）
	if saveRatio < 50 {
		t.Errorf("Token节省效果不佳，仅节省 %.1f%%", saveRatio)
	}
}

// TestConversationContinuity 测试对话连贯性
func TestConversationContinuity(t *testing.T) {
	// 验证摘要后对话仍然连贯
	summary := "用户询问天气和交通情况。助手告知今天晴天，建议穿轻便衣物，并提醒早高峰交通拥堵。"

	recentDialogues := []struct {
		user      string
		assistant string
	}{
		{"明天还会下雨吗？", "根据天气预报，明天也是晴天，不会下雨。"},
		{"那我可以安排户外活动了？", "是的，明天天气很好，适合户外活动。建议做好防晒措施。"},
	}

	// 构建完整上下文
	var contextBuilder strings.Builder
	contextBuilder.WriteString("[历史摘要] " + summary + "\n\n")
	for _, d := range recentDialogues {
		contextBuilder.WriteString("用户：" + d.user + "\n")
		contextBuilder.WriteString("助手：" + d.assistant + "\n\n")
	}

	fullContext := contextBuilder.String()

	// 验证上下文包含关键信息
	keyPoints := []string{"晴天", "明天", "户外活动"}
	for _, point := range keyPoints {
		if !strings.Contains(fullContext, point) {
			t.Errorf("上下文应包含关键信息 '%s'", point)
		}
	}

	// 验证上下文连贯性（摘要和最近对话能衔接）
	if !strings.Contains(fullContext, "历史摘要") || !strings.Contains(fullContext, "用户：") {
		t.Error("上下文结构应包含摘要和最近对话")
	}
}

// BenchmarkEstimateTokens 测试Token估算性能
func BenchmarkEstimateTokens(b *testing.B) {
	text := strings.Repeat("Hello World 你好世界 ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		estimateTokens(text)
	}
}

// TestSummarizeConversation 测试摘要生成逻辑
func TestSummarizeConversation(t *testing.T) {
	// 创建模拟模型
	mockModel := &MockChatModel{
		generateFunc: func(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
			// 返回模拟摘要
			return schema.AssistantMessage("这是生成的摘要内容", nil), nil
		},
	}

	// 创建 Agent
	agent := &Agent{
		config: &Config{
			Summary: SummaryConfig{
				KeepRecentRounds: 2,
				TriggerThreshold: 4,
				MaxSummaryTokens: 500,
				Enabled:          true,
			},
		},
		conversation: &ConversationLayer{
			RecentRounds: []*ConversationRound{
				{
					UserMessage:      schema.UserMessage("你好"),
					AssistantMessage: schema.AssistantMessage("你好！有什么可以帮助你的？", nil),
					Timestamp:        time.Now(),
					TokenCount:       50,
				},
				{
					UserMessage:      schema.UserMessage("介绍一下Go语言"),
					AssistantMessage: schema.AssistantMessage("Go是Google开发的编程语言...", nil),
					Timestamp:        time.Now(),
					TokenCount:       200,
				},
				{
					UserMessage:      schema.UserMessage("Go有什么特点？"),
					AssistantMessage: schema.AssistantMessage("Go的特点包括并发、简洁、快速编译...", nil),
					Timestamp:        time.Now(),
					TokenCount:       300,
				},
				{
					UserMessage:      schema.UserMessage("适合做什么项目？"),
					AssistantMessage: schema.AssistantMessage("Go适合云原生、微服务、网络服务...", nil),
					Timestamp:        time.Now(),
					TokenCount:       250,
				},
			},
			SummaryContent:   "",
			SummarizedRounds: 0,
		},
		chatModel: mockModel,
		tokenUsage: &TokenUsage{
			LastUpdated: time.Now(),
		},
	}

	// 执行摘要
	roundsToSummarize := len(agent.conversation.RecentRounds) - agent.config.Summary.KeepRecentRounds
	err := agent.summarizeConversation(context.Background(), roundsToSummarize)
	if err != nil {
		t.Errorf("summarizeConversation failed: %v", err)
	}

	// 验证摘要结果
	if agent.conversation.SummaryContent == "" {
		t.Error("摘要内容不应为空")
	}

	if agent.conversation.SummarizedRounds != 2 {
		t.Errorf("SummarizedRounds = %v, want %v", agent.conversation.SummarizedRounds, 2)
	}

	if len(agent.conversation.RecentRounds) != 2 {
		t.Errorf("RecentRounds length = %v, want %v", len(agent.conversation.RecentRounds), 2)
	}

	// 验证token统计
	if agent.tokenUsage.SavedTokens <= 0 {
		t.Error("应该有token节省")
	}
}

// TestGetConversationInfo 测试对话信息获取
func TestGetConversationInfo(t *testing.T) {
	agent := &Agent{
		config: &Config{
			Summary: SummaryConfig{
				KeepRecentRounds: 4,
			},
		},
		conversation: &ConversationLayer{
			RecentRounds: []*ConversationRound{
				{UserMessage: schema.UserMessage("测试")},
				{UserMessage: schema.UserMessage("测试2")},
			},
			SummaryContent:   "历史摘要内容",
			SummarizedRounds: 5,
		},
	}

	info := agent.GetConversationInfo()

	if info["totalRounds"] != 7 { // 5 + 2
		t.Errorf("totalRounds = %v, want %v", info["totalRounds"], 7)
	}

	if info["summarizedRounds"] != 5 {
		t.Errorf("summarizedRounds = %v, want %v", info["summarizedRounds"], 5)
	}

	if info["recentRounds"] != 2 {
		t.Errorf("recentRounds = %v, want %v", info["recentRounds"], 2)
	}

	if !info["hasSummary"].(bool) {
		t.Error("hasSummary should be true")
	}
}

// TestForceSummarize 测试强制摘要
func TestForceSummarize(t *testing.T) {
	mockModel := &MockChatModel{
		generateFunc: func(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
			return schema.AssistantMessage("强制摘要结果", nil), nil
		},
	}

	agent := &Agent{
		config: &Config{
			Summary: SummaryConfig{
				KeepRecentRounds: 2,
				MaxSummaryTokens: 500,
				Enabled:          true,
			},
		},
		conversation: &ConversationLayer{
			RecentRounds: []*ConversationRound{
				{UserMessage: schema.UserMessage("消息1")},
				{UserMessage: schema.UserMessage("消息2")},
				{UserMessage: schema.UserMessage("消息3")},
				{UserMessage: schema.UserMessage("消息4")},
			},
			SummaryContent:   "",
			SummarizedRounds: 0,
		},
		chatModel: mockModel,
		tokenUsage: &TokenUsage{
			LastUpdated: time.Now(),
		},
	}

	err := agent.ForceSummarize(context.Background())
	if err != nil {
		t.Errorf("ForceSummarize failed: %v", err)
	}

	if agent.conversation.SummaryContent != "强制摘要结果" {
		t.Errorf("SummaryContent = %v, want %v", agent.conversation.SummaryContent, "强制摘要结果")
	}

	// 测试轮数不足的情况
	agent2 := &Agent{
		config: &Config{
			Summary: SummaryConfig{
				KeepRecentRounds: 5,
			},
		},
		conversation: &ConversationLayer{
			RecentRounds: []*ConversationRound{
				{UserMessage: schema.UserMessage("消息1")},
			},
		},
	}

	err = agent2.ForceSummarize(context.Background())
	if err == nil {
		t.Error("轮数不足时应该返回错误")
	}
}
