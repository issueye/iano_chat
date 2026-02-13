// Package script - 脚本执行器
// 用于 Hook 和 Agent 的脚本执行

package iano_script_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"iano_script_engine/builtin"
)

// ScriptExecutor 脚本执行器实现
type ScriptExecutor struct {
	engine  Engine
	modules []Module
	config  *ExecutorConfig
}

// NewExecutor 创建脚本执行器
func NewExecutor(config *ExecutorConfig) Executor {
	if config == nil {
		config = DefaultExecutorConfig()
	}

	// 构建模块列表
	modules := make([]Module, 0)
	if config.EnableHTTP {
		modules = append(modules, builtin.NewHTTPModule(config.DefaultTimeout))
	}
	if config.EnableUtils {
		modules = append(modules, builtin.NewUtilsModule())
	}
	if config.EnableURL {
		modules = append(modules, builtin.NewURLModule())
	}
	if config.EnableFile {
		modules = append(modules, builtin.NewFileModule(nil))
	}
	if config.EnableCmd {
		modules = append(modules, builtin.NewCmdModule(nil))
	}

	engine := NewEngineWithModules(&Config{
		Timeout:          config.DefaultTimeout,
		MemoryLimit:      10 * 1024 * 1024,
		MaxCallStackSize: 1000,
	}, modules...)

	return &ScriptExecutor{
		engine:  engine,
		modules: modules,
		config:  config,
	}
}

// Execute 执行脚本
func (e *ScriptExecutor) Execute(ctx context.Context, script string, input map[string]interface{}) (*ExecutionResult, error) {
	start := time.Now()

	// 设置超时
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.config.DefaultTimeout)
		defer cancel()
	}

	// 执行脚本
	result, err := e.engine.Execute(ctx, script, input)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return &ExecutionResult{
			Success:  false,
			Error:    err.Error(),
			Duration: duration,
		}, nil
	}

	return &ExecutionResult{
		Success:  result.Success,
		Result:   result.Value,
		Error:    result.Error,
		Logs:     result.Logs,
		Duration: duration,
	}, nil
}

// ExecuteWithTimeout 带超时的执行
func (e *ScriptExecutor) ExecuteWithTimeout(script string, input map[string]interface{}, timeout time.Duration) (*ExecutionResult, error) {
	if timeout > e.config.MaxTimeout {
		timeout = e.config.MaxTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return e.Execute(ctx, script, input)
}

// Validate 验证脚本
func (e *ScriptExecutor) Validate(script string) error {
	return e.engine.Validate(script)
}

// SetGlobal 设置全局变量
func (e *ScriptExecutor) SetGlobal(key string, value interface{}) {
	if engine, ok := e.engine.(*GojaEngine); ok {
		engine.SetGlobal(key, value)
	}
}

// SetFunction 设置全局函数
func (e *ScriptExecutor) SetFunction(name string, fn interface{}) {
	if engine, ok := e.engine.(*GojaEngine); ok {
		engine.SetFunction(name, fn)
	}
}

// ToJSON 转换为 JSON
func (r *ExecutionResult) ToJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// HookScriptExecutor 专门用于 Hook 的脚本执行器
type HookScriptExecutor struct {
	*ScriptExecutor
}

// NewHookScriptExecutor 创建 Hook 脚本执行器
func NewHookScriptExecutor() HookExecutor {
	config := &ExecutorConfig{
		DefaultTimeout: 10 * time.Second,
		MaxTimeout:     60 * time.Second,
		EnableHTTP:     true,
		EnableUtils:    true,
		EnableURL:      false,
	}

	return &HookScriptExecutor{
		ScriptExecutor: NewExecutor(config).(*ScriptExecutor),
	}
}

// ExecuteHook 执行 Hook 脚本
func (e *HookScriptExecutor) ExecuteHook(ctx context.Context, script string, event string, data map[string]interface{}) (*ExecutionResult, error) {
	input := map[string]interface{}{
		"event":     event,
		"timestamp": time.Now().Unix(),
		"data":      data,
	}

	return e.Execute(ctx, script, input)
}

// AgentScriptExecutor 专门用于 Agent 的脚本执行器
type AgentScriptExecutor struct {
	*ScriptExecutor
}

// NewAgentScriptExecutor 创建 Agent 脚本执行器
func NewAgentScriptExecutor() AgentExecutor {
	config := &ExecutorConfig{
		DefaultTimeout: 30 * time.Second,
		MaxTimeout:     5 * time.Minute,
		EnableHTTP:     true,
		EnableUtils:    true,
		EnableURL:      true,
	}

	return &AgentScriptExecutor{
		ScriptExecutor: NewExecutor(config).(*ScriptExecutor),
	}
}

// ExecuteTool 执行工具脚本
func (e *AgentScriptExecutor) ExecuteTool(ctx context.Context, script string, toolName string, args map[string]interface{}) (*ExecutionResult, error) {
	input := map[string]interface{}{
		"tool":      toolName,
		"args":      args,
		"timestamp": time.Now().Unix(),
	}

	return e.Execute(ctx, script, input)
}

// ExecuteTransform 执行数据转换脚本
func (e *AgentScriptExecutor) ExecuteTransform(ctx context.Context, script string, data interface{}) (*ExecutionResult, error) {
	input := map[string]interface{}{
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	return e.Execute(ctx, script, input)
}

// ExecuteFilter 执行过滤脚本
func (e *AgentScriptExecutor) ExecuteFilter(ctx context.Context, script string, item interface{}) (bool, error) {
	input := map[string]interface{}{
		"item":      item,
		"timestamp": time.Now().Unix(),
	}

	result, err := e.Execute(ctx, script, input)
	if err != nil {
		return false, err
	}

	if !result.Success {
		return false, fmt.Errorf("script error: %s", result.Error)
	}

	// 检查结果是否为 true
	if v, ok := result.Result.(bool); ok {
		return v, nil
	}

	// 如果是 map，检查是否有 "match" 或 "result" 字段
	if m, ok := result.Result.(map[string]interface{}); ok {
		if v, ok := m["match"].(bool); ok {
			return v, nil
		}
		if v, ok := m["result"].(bool); ok {
			return v, nil
		}
	}

	return false, nil
}

// Sandbox 脚本沙箱（安全执行环境）
type Sandbox struct {
	executor *ScriptExecutor
	limits   *SandboxLimits
}

// NewSandbox 创建沙箱
func NewSandbox(limits *SandboxLimits) SandboxExecutor {
	if limits == nil {
		limits = DefaultSandboxLimits()
	}

	config := &ExecutorConfig{
		DefaultTimeout: limits.MaxExecutionTime,
		MaxTimeout:     limits.MaxExecutionTime,
		EnableHTTP:     contains(limits.AllowedModules, "http"),
		EnableUtils:    contains(limits.AllowedModules, "utils"),
		EnableURL:      contains(limits.AllowedModules, "url"),
	}

	return &Sandbox{
		executor: NewExecutor(config).(*ScriptExecutor),
		limits:   limits,
	}
}

// Run 在沙箱中运行脚本
func (s *Sandbox) Run(ctx context.Context, scriptCode string, input map[string]interface{}) (*ExecutionResult, error) {
	// 检查脚本是否包含危险函数
	for _, blocked := range s.limits.BlockedFunctions {
		if strContains(scriptCode, blocked) {
			return &ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("script contains blocked function: %s", blocked),
			}, nil
		}
	}

	return s.executor.Execute(ctx, scriptCode, input)
}

// contains 检查字符串切片是否包含某个字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// strContains 检查字符串是否包含子串
func strContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
