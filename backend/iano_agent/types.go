package iano_agent

import (
	"context"
	"time"
)

type ToolCallInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type MessageCallback func(content string, isToolCall bool, toolCalls *ToolCallInfo)

type TokenUsage struct {
	TotalTokens      int64
	PromptTokens     int64
	CompletionTokens int64
	SummaryTokens    int64
	SavedTokens      int64
	LastUpdated      time.Time
}

type Tool struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, params map[string]interface{}) (string, error)
}
