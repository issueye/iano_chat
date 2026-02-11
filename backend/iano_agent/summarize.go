package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

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
