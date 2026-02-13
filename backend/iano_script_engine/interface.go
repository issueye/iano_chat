// Package script - 脚本引擎接口定义
// 定义核心接口、配置和结果类型

package iano_script_engine

import (
	"context"
	"time"

	"iano_script_engine/builtin"
)

// Engine 脚本引擎接口
type Engine interface {
	// Execute 执行脚本
	Execute(ctx context.Context, script string, input map[string]interface{}) (*Result, error)
	// ExecuteWithTimeout 带超时的执行
	ExecuteWithTimeout(script string, input map[string]interface{}, timeout time.Duration) (*Result, error)
	// Validate 验证脚本语法
	Validate(script string) error
	// SetGlobal 设置全局变量
	SetGlobal(key string, value interface{})
	// SetFunction 设置全局函数
	SetFunction(name string, fn interface{})
}

// Result 脚本执行结果
type Result struct {
	Success bool               `json:"success"`
	Value   interface{}        `json:"value,omitempty"`
	Error   string             `json:"error,omitempty"`
	Logs    []builtin.LogEntry `json:"logs,omitempty"`
}

// Config 脚本引擎配置
type Config struct {
	// Timeout 默认执行超时
	Timeout time.Duration
	// MemoryLimit 内存限制 (字节)
	MemoryLimit uint64
	// MaxCallStackSize 最大调用栈深度
	MaxCallStackSize int
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:          30 * time.Second,
		MemoryLimit:      10 * 1024 * 1024, // 10MB
		MaxCallStackSize: 1000,
	}
}

// Executor 脚本执行器接口
type Executor interface {
	// Execute 执行脚本
	Execute(ctx context.Context, script string, input map[string]interface{}) (*ExecutionResult, error)
	// ExecuteWithTimeout 带超时的执行
	ExecuteWithTimeout(script string, input map[string]interface{}, timeout time.Duration) (*ExecutionResult, error)
	// Validate 验证脚本
	Validate(script string) error
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success  bool               `json:"success"`
	Result   interface{}        `json:"result,omitempty"`
	Error    string             `json:"error,omitempty"`
	Logs     []builtin.LogEntry `json:"logs,omitempty"`
	Duration int64              `json:"duration_ms"`
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	DefaultTimeout time.Duration
	MaxTimeout     time.Duration
	EnableHTTP     bool
	EnableUtils    bool
	EnableURL      bool
	EnableFile     bool
	EnableCmd      bool
}

// DefaultExecutorConfig 默认配置
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		DefaultTimeout: 30 * time.Second,
		MaxTimeout:     5 * time.Minute,
		EnableHTTP:     true,
		EnableUtils:    true,
		EnableURL:      true,
		EnableFile:     false,
		EnableCmd:      false,
	}
}

// HookExecutor Hook 执行器接口
type HookExecutor interface {
	Executor
	// ExecuteHook 执行 Hook 脚本
	ExecuteHook(ctx context.Context, script string, event string, data map[string]interface{}) (*ExecutionResult, error)
}

// AgentExecutor Agent 执行器接口
type AgentExecutor interface {
	Executor
	// ExecuteTool 执行工具脚本
	ExecuteTool(ctx context.Context, script string, toolName string, args map[string]interface{}) (*ExecutionResult, error)
	// ExecuteTransform 执行数据转换脚本
	ExecuteTransform(ctx context.Context, script string, data interface{}) (*ExecutionResult, error)
	// ExecuteFilter 执行过滤脚本
	ExecuteFilter(ctx context.Context, script string, item interface{}) (bool, error)
}

// SandboxExecutor 沙箱执行器接口
type SandboxExecutor interface {
	// Run 在沙箱中运行脚本
	Run(ctx context.Context, scriptCode string, input map[string]interface{}) (*ExecutionResult, error)
}

// SandboxLimits 沙箱限制
type SandboxLimits struct {
	MaxExecutionTime time.Duration
	MaxMemoryMB      int64
	MaxOutputSize    int
	AllowedModules   []string
	BlockedFunctions []string
}

// DefaultSandboxLimits 默认沙箱限制
func DefaultSandboxLimits() *SandboxLimits {
	return &SandboxLimits{
		MaxExecutionTime: 5 * time.Second,
		MaxMemoryMB:      50,
		MaxOutputSize:    1024 * 1024, // 1MB
		AllowedModules:   []string{"utils", "url"},
		BlockedFunctions: []string{"http", "eval", "Function", "file", "cmd"},
	}
}
