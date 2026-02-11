package errors

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrCodeConfig       ErrorCode = "CONFIG_ERROR"
	ErrCodeModel        ErrorCode = "MODEL_ERROR"
	ErrCodeTool         ErrorCode = "TOOL_ERROR"
	ErrCodeNetwork      ErrorCode = "NETWORK_ERROR"
	ErrCodeRateLimit    ErrorCode = "RATE_LIMIT_ERROR"
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeConversation ErrorCode = "CONVERSATION_ERROR"
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeTimeout      ErrorCode = "TIMEOUT_ERROR"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND_ERROR"
)

type AgentError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]interface{}
}

func (e *AgentError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AgentError) Unwrap() error {
	return e.Cause
}

func (e *AgentError) WithDetail(key string, value interface{}) *AgentError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

func NewError(code ErrorCode, message string, cause ...error) *AgentError {
	err := &AgentError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
	if len(cause) > 0 && cause[0] != nil {
		err.Cause = cause[0]
	}
	return err
}

func IsErrorCode(err error, code ErrorCode) bool {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr.Code == code
	}
	return false
}

func GetErrorCode(err error) ErrorCode {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr.Code
	}
	return ErrCodeInternal
}

var (
	ErrInvalidConfig        = NewError(ErrCodeConfig, "配置无效")
	ErrMissingAPIKey        = NewError(ErrCodeConfig, "缺少 API Key")
	ErrInvalidModel         = NewError(ErrCodeConfig, "模型配置无效")
	ErrConversationNotFound = NewError(ErrCodeConversation, "对话不存在")
	ErrMaxRoundsExceeded    = NewError(ErrCodeConversation, "超过最大对话轮数")
	ErrEmptyInput           = NewError(ErrCodeValidation, "输入不能为空")
	ErrToolNotFound         = NewError(ErrCodeTool, "工具不存在")
	ErrToolExecution        = NewError(ErrCodeTool, "工具执行失败")
	ErrToolAlreadyExists    = NewError(ErrCodeTool, "工具已存在")
	ErrRateLimitExceeded    = NewError(ErrCodeRateLimit, "请求过于频繁，请稍后再试")
	ErrNetworkTimeout       = NewError(ErrCodeTimeout, "网络请求超时")
	ErrNetworkFailed        = NewError(ErrCodeNetwork, "网络请求失败")
	ErrInvalidURL           = NewError(ErrCodeValidation, "无效的 URL")
	ErrInvalidMethod        = NewError(ErrCodeValidation, "无效的 HTTP 方法")
	ErrInvalidParams        = NewError(ErrCodeValidation, "参数无效")
	ErrAccessDenied         = NewError(ErrCodeValidation, "访问被拒绝")
	ErrAgentNotFound        = NewError(ErrCodeNotFound, "Agent 不存在")
	ErrAgentAlreadyExists   = NewError(ErrCodeNotFound, "Agent 已存在")
	ErrSessionNotFound      = NewError(ErrCodeNotFound, "会话不存在")
)

func WrapError(code ErrorCode, message string, err error) *AgentError {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return NewError(agentErr.Code, message, err)
	}
	return NewError(code, message, err)
}

func WrapModelError(message string, err error) *AgentError {
	return WrapError(ErrCodeModel, message, err)
}

func WrapToolError(message string, err error) *AgentError {
	return WrapError(ErrCodeTool, message, err)
}

func WrapNetworkError(message string, err error) *AgentError {
	return WrapError(ErrCodeNetwork, message, err)
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	code := GetErrorCode(err)
	return code == ErrCodeNetwork || code == ErrCodeTimeout || code == ErrCodeRateLimit
}

type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *AgentError) ToErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Code:    string(e.Code),
		Message: e.Message,
		Details: e.Details,
	}
}

func NewErrorResponseFromError(err error) *ErrorResponse {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr.ToErrorResponse()
	}
	return &ErrorResponse{
		Code:    string(ErrCodeInternal),
		Message: err.Error(),
	}
}
