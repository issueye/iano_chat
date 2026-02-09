package agent

import (
	"errors"
	"fmt"
)

// ErrorCode 错误代码类型
type ErrorCode string

// 预定义错误代码
const (
	// 配置错误
	ErrCodeConfig ErrorCode = "CONFIG_ERROR"
	// 模型错误
	ErrCodeModel ErrorCode = "MODEL_ERROR"
	// 工具错误
	ErrCodeTool ErrorCode = "TOOL_ERROR"
	// 网络错误
	ErrCodeNetwork ErrorCode = "NETWORK_ERROR"
	// 限流错误
	ErrCodeRateLimit ErrorCode = "RATE_LIMIT_ERROR"
	// 验证错误
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// 对话错误
	ErrCodeConversation ErrorCode = "CONVERSATION_ERROR"
	// 内部错误
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	// 超时错误
	ErrCodeTimeout ErrorCode = "TIMEOUT_ERROR"
	// 未找到错误
	ErrCodeNotFound ErrorCode = "NOT_FOUND_ERROR"
)

// AgentError Agent 错误结构
type AgentError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Details map[string]interface{}
}

// Error 实现 error 接口
func (e *AgentError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误，支持 errors.Is 和 errors.As
func (e *AgentError) Unwrap() error {
	return e.Cause
}

// WithDetail 添加错误详情
func (e *AgentError) WithDetail(key string, value interface{}) *AgentError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// NewError 创建新的 AgentError
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

// IsErrorCode 检查错误是否匹配指定代码
func IsErrorCode(err error, code ErrorCode) bool {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr.Code == code
	}
	return false
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) ErrorCode {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr.Code
	}
	return ErrCodeInternal
}

// 预定义错误变量
var (
	// 配置错误
	ErrInvalidConfig = NewError(ErrCodeConfig, "配置无效")
	ErrMissingAPIKey = NewError(ErrCodeConfig, "缺少 API Key")
	ErrInvalidModel  = NewError(ErrCodeConfig, "模型配置无效")

	// 对话错误
	ErrConversationNotFound = NewError(ErrCodeConversation, "对话不存在")
	ErrMaxRoundsExceeded    = NewError(ErrCodeConversation, "超过最大对话轮数")
	ErrEmptyInput           = NewError(ErrCodeValidation, "输入不能为空")

	// 工具错误
	ErrToolNotFound      = NewError(ErrCodeTool, "工具不存在")
	ErrToolExecution     = NewError(ErrCodeTool, "工具执行失败")
	ErrToolAlreadyExists = NewError(ErrCodeTool, "工具已存在")

	// 限流错误
	ErrRateLimitExceeded = NewError(ErrCodeRateLimit, "请求过于频繁，请稍后再试")

	// 网络错误
	ErrNetworkTimeout = NewError(ErrCodeTimeout, "网络请求超时")
	ErrNetworkFailed  = NewError(ErrCodeNetwork, "网络请求失败")

	// 验证错误
	ErrInvalidURL     = NewError(ErrCodeValidation, "无效的 URL")
	ErrInvalidMethod  = NewError(ErrCodeValidation, "无效的 HTTP 方法")
	ErrInvalidParams  = NewError(ErrCodeValidation, "参数无效")
	ErrAccessDenied   = NewError(ErrCodeValidation, "访问被拒绝")
)

// WrapError 包装错误为 AgentError
func WrapError(code ErrorCode, message string, err error) *AgentError {
	// 如果已经是 AgentError，保留原错误代码
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return NewError(agentErr.Code, message, err)
	}
	return NewError(code, message, err)
}

// WrapModelError 包装模型相关错误
func WrapModelError(message string, err error) *AgentError {
	return WrapError(ErrCodeModel, message, err)
}

// WrapToolError 包装工具相关错误
func WrapToolError(message string, err error) *AgentError {
	return WrapError(ErrCodeTool, message, err)
}

// WrapNetworkError 包装网络相关错误
func WrapNetworkError(message string, err error) *AgentError {
	return WrapError(ErrCodeNetwork, message, err)
}

// WrapValidationError 包装验证错误
func WrapValidationError(message string, err error) *AgentError {
	return WrapError(ErrCodeValidation, message, err)
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	code := GetErrorCode(err)
	switch code {
	case ErrCodeNetwork, ErrCodeTimeout, ErrCodeRateLimit:
		return true
	default:
		return false
	}
}

// ErrorResponse 错误响应结构（用于 API 返回）
type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToErrorResponse 转换为错误响应
func (e *AgentError) ToErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Code:    string(e.Code),
		Message: e.Message,
		Details: e.Details,
	}
}

// NewErrorResponseFromError 从错误创建错误响应
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
