package callback

// {"role":"assistant","content":"","response_meta":{"finish_reason":"stop","usage":{"prompt_tokens":2058,"prompt_token_details":{"cached_tokens":42},"completion_tokens":1165,"total_tokens":3223,"completion_token_details":{"reasoning_tokens":135}}},"extra":{"openai-request-id":"202602112043267e5892f57c654e0e"}}
type PromptTokenDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type CompletionTokenDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type Usage struct {
	PromptTokens           int                    `json:"prompt_tokens"`
	PromptTokenDetails     PromptTokenDetails     `json:"prompt_token_details"`
	CompletionTokens       int                    `json:"completion_tokens"`
	TotalTokens            int                    `json:"total_tokens"`
	CompletionTokenDetails CompletionTokenDetails `json:"completion_token_details"`
}

type ResponseMeta struct {
	FinishReason string `json:"finish_reason"`
	Usage        Usage  `json:"usage"`
}

type Extra struct {
	OpenAIRequestID string `json:"openai-request-id"`
}

type ReActAgentOut struct {
	Role         string       `json:"role"`
	Content      string       `json:"content"`
	ResponseMeta ResponseMeta `json:"response_meta"`
	Extra        Extra        `json:"extra"`
}
