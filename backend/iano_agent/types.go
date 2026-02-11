package agent

import (
	"context"
	"time"
)

type MessageCallback func(content string, isToolCall bool)

type TokenUsage struct {
	TotalTokens      int64
	PromptTokens     int64
	CompletionTokens int64
	SummaryTokens    int64
	SavedTokens      int64
	LastUpdated      time.Time
}

type SummaryConfig struct {
	KeepRecentRounds int
	TriggerThreshold int
	MaxSummaryTokens int
	Enabled          bool
}

type Tool struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, params map[string]interface{}) (string, error)
}
