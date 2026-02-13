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

type Message struct {
	Role             string        `json:"role"`
	Content          string        `json:"content"`
	ReasoningContent string        `json:"reasoning_content"`
	ThinkContent     string        `json:"think_content"`
	ToolCall         *ToolCallInfo `json:"tool_call"`
	CallToolError    string        `json:"call_tool_error"`
	IsThink          bool          `json:"is_think"`
	IsReasoning      bool          `json:"is_reasoning"`
	IsToolCall       bool          `json:"is_tool_call"`
}

type MessageCallback func(msg *Message)

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
