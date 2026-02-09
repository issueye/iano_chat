package services

import (
	"fmt"
)

// ErrorCode 错误码类型
type ErrorCode int

// 错误码定义
const (
	// 通用错误 (1000-1099)
	ErrCodeUnknown       ErrorCode = 1000 // 未知错误
	ErrCodeInvalidParam  ErrorCode = 1001 // 参数无效
	ErrCodeInternal      ErrorCode = 1002 // 内部错误
	ErrCodeUnauthorized  ErrorCode = 1003 // 未授权
	ErrCodeForbidden     ErrorCode = 1004 // 禁止访问
	ErrCodeNotFound      ErrorCode = 1005 // 资源不存在
	ErrCodeAlreadyExists ErrorCode = 1006 // 资源已存在
	ErrCodeDatabase      ErrorCode = 1007 // 数据库错误

	// 会话相关错误 (2000-2099)
	ErrCodeSessionNotFound  ErrorCode = 2001 // 会话不存在
	ErrCodeSessionArchived  ErrorCode = 2002 // 会话已归档
	ErrCodeSessionCompleted ErrorCode = 2003 // 会话已完成
	ErrCodeSessionLimit     ErrorCode = 2004 // 会话数量超限
	ErrCodeSessionConfig    ErrorCode = 2005 // 会话配置错误

	// 消息相关错误 (3000-3099)
	ErrCodeMessageNotFound    ErrorCode = 3001 // 消息不存在
	ErrCodeMessageEditFailed  ErrorCode = 3002 // 消息编辑失败
	ErrCodeMessageDelete      ErrorCode = 3003 // 消息删除失败
	ErrCodeInvalidMessageType ErrorCode = 3004 // 无效的消息类型

	// Agent 相关错误 (4000-4099)
	ErrCodeAgentNotFound     ErrorCode = 4001 // Agent 实例不存在
	ErrCodeAgentCreateFailed ErrorCode = 4002 // Agent 创建失败
	ErrCodeAgentBusy         ErrorCode = 4003 // Agent 正忙
	ErrCodeAgentConfig       ErrorCode = 4004 // Agent 配置错误
	ErrCodeAgentExecution    ErrorCode = 4005 // Agent 执行错误

	// 限流相关错误 (5000-5099)
	ErrCodeRateLimitUser    ErrorCode = 5001 // 用户级别限流
	ErrCodeRateLimitSession ErrorCode = 5002 // 会话级别限流
	ErrCodeRateLimitGlobal  ErrorCode = 5003 // 全局限流
	ErrCodeConcurrentLimit  ErrorCode = 5004 // 并发限制

	// 工具相关错误 (6000-6099)
	ErrCodeToolNotFound     ErrorCode = 6001 // 工具不存在
	ErrCodeToolDisabled     ErrorCode = 6002 // 工具已禁用
	ErrCodeToolExecution    ErrorCode = 6003 // 工具执行失败
	ErrCodeToolParamInvalid ErrorCode = 6004 // 工具参数无效

	// 模型相关错误 (7000-7099)
	ErrCodeModelNotFound    ErrorCode = 7001 // 模型不存在
	ErrCodeModelDisabled    ErrorCode = 7002 // 模型已禁用
	ErrCodeModelConfig      ErrorCode = 7003 // 模型配置错误
	ErrCodeModelHealthCheck ErrorCode = 7004 // 模型健康检查失败

	// 权限相关错误 (8000-8099)
	ErrCodePermissionDenied ErrorCode = 8001 // 权限不足
)

// ErrorCodeMessages 错误码对应的消息
var ErrorCodeMessages = map[ErrorCode]string{
	ErrCodeUnknown:       "未知错误",
	ErrCodeInvalidParam:  "参数无效",
	ErrCodeInternal:      "内部错误",
	ErrCodeUnauthorized:  "未授权",
	ErrCodeForbidden:     "禁止访问",
	ErrCodeNotFound:      "资源不存在",
	ErrCodeAlreadyExists: "资源已存在",
	ErrCodeDatabase:      "数据库错误",

	ErrCodeSessionNotFound:  "会话不存在",
	ErrCodeSessionArchived:  "会话已归档",
	ErrCodeSessionCompleted: "会话已完成",
	ErrCodeSessionLimit:     "会话数量超限",
	ErrCodeSessionConfig:    "会话配置错误",

	ErrCodeMessageNotFound:    "消息不存在",
	ErrCodeMessageEditFailed:  "消息编辑失败",
	ErrCodeMessageDelete:      "消息删除失败",
	ErrCodeInvalidMessageType: "无效的消息类型",

	ErrCodeAgentNotFound:     "Agent 实例不存在",
	ErrCodeAgentCreateFailed: "Agent 创建失败",
	ErrCodeAgentBusy:         "Agent 正忙",
	ErrCodeAgentConfig:       "Agent 配置错误",
	ErrCodeAgentExecution:    "Agent 执行错误",

	ErrCodeRateLimitUser:    "请求过于频繁，请稍后再试",
	ErrCodeRateLimitSession: "当前会话请求过于频繁",
	ErrCodeRateLimitGlobal:  "系统繁忙，请稍后再试",
	ErrCodeConcurrentLimit:  "当前有正在进行的请求",

	ErrCodeToolNotFound:     "工具不存在",
	ErrCodeToolDisabled:     "工具已禁用",
	ErrCodeToolExecution:    "工具执行失败",
	ErrCodeToolParamInvalid: "工具参数无效",

	ErrCodeModelNotFound:    "模型不存在",
	ErrCodeModelDisabled:    "模型已禁用",
	ErrCodeModelConfig:      "模型配置错误",
	ErrCodeModelHealthCheck: "模型健康检查失败",

	ErrCodePermissionDenied: "权限不足",
}

// ServiceError 服务错误
type ServiceError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Detail  string    `json:"detail,omitempty"`
	Cause   error     `json:"-"`
}

// Error 实现 error 接口
func (e *ServiceError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// WithDetail 添加详细错误信息
func (e *ServiceError) WithDetail(detail string) *ServiceError {
	return &ServiceError{
		Code:    e.Code,
		Message: e.Message,
		Detail:  detail,
		Cause:   e.Cause,
	}
}

// WithCause 添加原始错误
func (e *ServiceError) WithCause(cause error) *ServiceError {
	return &ServiceError{
		Code:    e.Code,
		Message: e.Message,
		Detail:  e.Detail,
		Cause:   cause,
	}
}

// Is 判断错误是否匹配
func (e *ServiceError) Is(target error) bool {
	if t, ok := target.(*ServiceError); ok {
		return e.Code == t.Code
	}
	return false
}

// NewError 创建新的服务错误
func NewError(code ErrorCode, detail string) *ServiceError {
	msg, ok := ErrorCodeMessages[code]
	if !ok {
		msg = "未知错误"
	}
	return &ServiceError{
		Code:    code,
		Message: msg,
		Detail:  detail,
	}
}

// NewErrorWithCause 创建带原始错误的服务错误
func NewErrorWithCause(code ErrorCode, cause error) *ServiceError {
	msg, ok := ErrorCodeMessages[code]
	if !ok {
		msg = "未知错误"
	}
	return &ServiceError{
		Code:    code,
		Message: msg,
		Cause:   cause,
		Detail:  cause.Error(),
	}
}

// 预定义错误实例（便于直接使用）
var (
	// 通用错误
	ErrUnknown       = &ServiceError{Code: ErrCodeUnknown, Message: ErrorCodeMessages[ErrCodeUnknown]}
	ErrInvalidParam  = &ServiceError{Code: ErrCodeInvalidParam, Message: ErrorCodeMessages[ErrCodeInvalidParam]}
	ErrInternal      = &ServiceError{Code: ErrCodeInternal, Message: ErrorCodeMessages[ErrCodeInternal]}
	ErrUnauthorized  = &ServiceError{Code: ErrCodeUnauthorized, Message: ErrorCodeMessages[ErrCodeUnauthorized]}
	ErrForbidden     = &ServiceError{Code: ErrCodeForbidden, Message: ErrorCodeMessages[ErrCodeForbidden]}
	ErrNotFound      = &ServiceError{Code: ErrCodeNotFound, Message: ErrorCodeMessages[ErrCodeNotFound]}
	ErrAlreadyExists = &ServiceError{Code: ErrCodeAlreadyExists, Message: ErrorCodeMessages[ErrCodeAlreadyExists]}
	ErrDatabase      = &ServiceError{Code: ErrCodeDatabase, Message: ErrorCodeMessages[ErrCodeDatabase]}

	// 会话错误
	ErrSessionNotFound  = &ServiceError{Code: ErrCodeSessionNotFound, Message: ErrorCodeMessages[ErrCodeSessionNotFound]}
	ErrSessionArchived  = &ServiceError{Code: ErrCodeSessionArchived, Message: ErrorCodeMessages[ErrCodeSessionArchived]}
	ErrSessionCompleted = &ServiceError{Code: ErrCodeSessionCompleted, Message: ErrorCodeMessages[ErrCodeSessionCompleted]}
	ErrSessionLimit     = &ServiceError{Code: ErrCodeSessionLimit, Message: ErrorCodeMessages[ErrCodeSessionLimit]}
	ErrSessionConfig    = &ServiceError{Code: ErrCodeSessionConfig, Message: ErrorCodeMessages[ErrCodeSessionConfig]}

	// 消息错误
	ErrMessageNotFound    = &ServiceError{Code: ErrCodeMessageNotFound, Message: ErrorCodeMessages[ErrCodeMessageNotFound]}
	ErrMessageEditFailed  = &ServiceError{Code: ErrCodeMessageEditFailed, Message: ErrorCodeMessages[ErrCodeMessageEditFailed]}
	ErrMessageDelete      = &ServiceError{Code: ErrCodeMessageDelete, Message: ErrorCodeMessages[ErrCodeMessageDelete]}
	ErrInvalidMessageType = &ServiceError{Code: ErrCodeInvalidMessageType, Message: ErrorCodeMessages[ErrCodeInvalidMessageType]}

	// Agent 错误
	ErrAgentNotFound     = &ServiceError{Code: ErrCodeAgentNotFound, Message: ErrorCodeMessages[ErrCodeAgentNotFound]}
	ErrAgentCreateFailed = &ServiceError{Code: ErrCodeAgentCreateFailed, Message: ErrorCodeMessages[ErrCodeAgentCreateFailed]}
	ErrAgentBusy         = &ServiceError{Code: ErrCodeAgentBusy, Message: ErrorCodeMessages[ErrCodeAgentBusy]}
	ErrAgentConfig       = &ServiceError{Code: ErrCodeAgentConfig, Message: ErrorCodeMessages[ErrCodeAgentConfig]}
	ErrAgentExecution    = &ServiceError{Code: ErrCodeAgentExecution, Message: ErrorCodeMessages[ErrCodeAgentExecution]}

	// 限流错误
	ErrRateLimitUser    = &ServiceError{Code: ErrCodeRateLimitUser, Message: ErrorCodeMessages[ErrCodeRateLimitUser]}
	ErrRateLimitSession = &ServiceError{Code: ErrCodeRateLimitSession, Message: ErrorCodeMessages[ErrCodeRateLimitSession]}
	ErrRateLimitGlobal  = &ServiceError{Code: ErrCodeRateLimitGlobal, Message: ErrorCodeMessages[ErrCodeRateLimitGlobal]}
	ErrConcurrentLimit  = &ServiceError{Code: ErrCodeConcurrentLimit, Message: ErrorCodeMessages[ErrCodeConcurrentLimit]}

	// 工具错误
	ErrToolNotFound     = &ServiceError{Code: ErrCodeToolNotFound, Message: ErrorCodeMessages[ErrCodeToolNotFound]}
	ErrToolDisabled     = &ServiceError{Code: ErrCodeToolDisabled, Message: ErrorCodeMessages[ErrCodeToolDisabled]}
	ErrToolExecution    = &ServiceError{Code: ErrCodeToolExecution, Message: ErrorCodeMessages[ErrCodeToolExecution]}
	ErrToolParamInvalid = &ServiceError{Code: ErrCodeToolParamInvalid, Message: ErrorCodeMessages[ErrCodeToolParamInvalid]}
)

// IsServiceError 判断是否为服务错误
func IsServiceError(err error) (*ServiceError, bool) {
	if err == nil {
		return nil, false
	}
	if se, ok := err.(*ServiceError); ok {
		return se, true
	}
	return nil, false
}

// IsErrorCode 判断错误是否匹配指定错误码
func IsErrorCode(err error, code ErrorCode) bool {
	if se, ok := IsServiceError(err); ok {
		return se.Code == code
	}
	return false
}

// GetHTTPStatus 获取错误对应的 HTTP 状态码
func (e *ServiceError) GetHTTPStatus() int {
	switch e.Code {
	case ErrCodeInvalidParam, ErrCodeSessionConfig, ErrCodeAgentConfig, ErrCodeToolParamInvalid:
		return 400
	case ErrCodeUnauthorized:
		return 401
	case ErrCodeForbidden:
		return 403
	case ErrCodeNotFound, ErrCodeSessionNotFound, ErrCodeMessageNotFound, ErrCodeAgentNotFound, ErrCodeToolNotFound:
		return 404
	case ErrCodeAlreadyExists:
		return 409
	case ErrCodeRateLimitUser, ErrCodeRateLimitSession, ErrCodeRateLimitGlobal, ErrCodeConcurrentLimit:
		return 429
	case ErrCodeInternal, ErrCodeDatabase, ErrCodeAgentCreateFailed, ErrCodeAgentExecution:
		return 500
	default:
		return 500
	}
}

// ToResponse 转换为响应格式
func (e *ServiceError) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"code":    e.Code,
		"message": e.Message,
		"detail":  e.Detail,
	}
}
