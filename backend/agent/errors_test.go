package agent

import (
	"errors"
	"testing"
)

func TestAgentError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AgentError
		expected string
	}{
		{
			name: "无原因错误",
			err: &AgentError{
				Code:    ErrCodeValidation,
				Message: "参数无效",
			},
			expected: "[VALIDATION_ERROR] 参数无效",
		},
		{
			name: "有原因错误",
			err: &AgentError{
				Code:    ErrCodeNetwork,
				Message: "请求失败",
				Cause:   errors.New("connection refused"),
			},
			expected: "[NETWORK_ERROR] 请求失败: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAgentError_Unwrap(t *testing.T) {
	cause := errors.New("原始错误")
	err := &AgentError{
		Code:    ErrCodeInternal,
		Message: "包装错误",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap() should return cause error")
	}

	// 测试 errors.Is
	if !errors.Is(err, cause) {
		t.Error("errors.Is() should return true for cause")
	}
}

func TestAgentError_WithDetail(t *testing.T) {
	err := NewError(ErrCodeValidation, "验证失败").
		WithDetail("field", "username").
		WithDetail("reason", "too short")

	if err.Details == nil {
		t.Fatal("Details should not be nil")
	}

	if err.Details["field"] != "username" {
		t.Error("Details[field] should be 'username'")
	}

	if err.Details["reason"] != "too short" {
		t.Error("Details[reason] should be 'too short'")
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name      string
		code      ErrorCode
		message   string
		cause     error
		wantCause bool
	}{
		{
			name:      "无原因",
			code:      ErrCodeConfig,
			message:   "配置错误",
			cause:     nil,
			wantCause: false,
		},
		{
			name:      "有原因",
			code:      ErrCodeModel,
			message:   "模型错误",
			cause:     errors.New("API 错误"),
			wantCause: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err *AgentError
			if tt.cause != nil {
				err = NewError(tt.code, tt.message, tt.cause)
			} else {
				err = NewError(tt.code, tt.message)
			}

			if err.Code != tt.code {
				t.Errorf("Code = %v, want %v", err.Code, tt.code)
			}

			if err.Message != tt.message {
				t.Errorf("Message = %v, want %v", err.Message, tt.message)
			}

			if tt.wantCause && err.Cause == nil {
				t.Error("Cause should not be nil")
			}

			if !tt.wantCause && err.Cause != nil {
				t.Error("Cause should be nil")
			}
		})
	}
}

func TestIsErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     ErrorCode
		expected bool
	}{
		{
			name:     "匹配",
			err:      NewError(ErrCodeValidation, "验证失败"),
			code:     ErrCodeValidation,
			expected: true,
		},
		{
			name:     "不匹配",
			err:      NewError(ErrCodeValidation, "验证失败"),
			code:     ErrCodeNetwork,
			expected: false,
		},
		{
			name:     "非 AgentError",
			err:      errors.New("普通错误"),
			code:     ErrCodeValidation,
			expected: false,
		},
		{
			name:     "nil 错误",
			err:      nil,
			code:     ErrCodeValidation,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsErrorCode(tt.err, tt.code)
			if got != tt.expected {
				t.Errorf("IsErrorCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCode
	}{
		{
			name:     "AgentError",
			err:      NewError(ErrCodeTool, "工具错误"),
			expected: ErrCodeTool,
		},
		{
			name:     "普通错误",
			err:      errors.New("普通错误"),
			expected: ErrCodeInternal,
		},
		{
			name:     "nil",
			err:      nil,
			expected: ErrCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorCode(tt.err)
			if got != tt.expected {
				t.Errorf("GetErrorCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := NewError(ErrCodeNetwork, "网络错误")
	wrappedErr := WrapError(ErrCodeInternal, "内部错误", originalErr)

	// 验证保留了原始错误代码
	if !IsErrorCode(wrappedErr, ErrCodeNetwork) {
		t.Error("WrapError should preserve original error code")
	}

	// 验证消息已更新
	if wrappedErr.Message != "内部错误" {
		t.Errorf("Message = %v, want '内部错误'", wrappedErr.Message)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "网络错误可重试",
			err:      NewError(ErrCodeNetwork, "网络错误"),
			expected: true,
		},
		{
			name:     "超时错误可重试",
			err:      NewError(ErrCodeTimeout, "超时"),
			expected: true,
		},
		{
			name:     "限流错误可重试",
			err:      NewError(ErrCodeRateLimit, "限流"),
			expected: true,
		},
		{
			name:     "验证错误不可重试",
			err:      NewError(ErrCodeValidation, "验证失败"),
			expected: false,
		},
		{
			name:     "nil 不可重试",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableError(tt.err)
			if got != tt.expected {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	err := NewError(ErrCodeValidation, "验证失败").
		WithDetail("field", "username").
		WithDetail("value", "ab")

	resp := err.ToErrorResponse()

	if resp.Code != string(ErrCodeValidation) {
		t.Errorf("Code = %v, want %v", resp.Code, ErrCodeValidation)
	}

	if resp.Message != "验证失败" {
		t.Errorf("Message = %v, want '验证失败'", resp.Message)
	}

	if resp.Details == nil {
		t.Fatal("Details should not be nil")
	}

	if resp.Details["field"] != "username" {
		t.Error("Details[field] should be 'username'")
	}
}

func TestNewErrorResponseFromError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "AgentError",
			err:          NewError(ErrCodeTool, "工具错误"),
			expectedCode: "TOOL_ERROR",
		},
		{
			name:         "普通错误",
			err:          errors.New("普通错误"),
			expectedCode: "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewErrorResponseFromError(tt.err)
			if resp.Code != tt.expectedCode {
				t.Errorf("Code = %v, want %v", resp.Code, tt.expectedCode)
			}
		})
	}
}
