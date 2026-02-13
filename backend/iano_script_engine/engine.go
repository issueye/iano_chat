// Package script - JavaScript 脚本引擎
// 基于 goja 实现安全的 JavaScript 执行环境

package iano_script_engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dop251/goja"

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
	Success bool              `json:"success"`
	Value   interface{}       `json:"value,omitempty"`
	Error   string            `json:"error,omitempty"`
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

// GojaEngine goja 脚本引擎实现
type GojaEngine struct {
	config  *Config
	globals map[string]interface{}
	funcs   map[string]interface{}
}

// NewEngine 创建新的脚本引擎
func NewEngine(config *Config) Engine {
	if config == nil {
		config = DefaultConfig()
	}

	return &GojaEngine{
		config:  config,
		globals: make(map[string]interface{}),
		funcs:   make(map[string]interface{}),
	}
}

// Execute 执行脚本
func (e *GojaEngine) Execute(ctx context.Context, script string, input map[string]interface{}) (*Result, error) {
	return e.executeWithContext(ctx, script, input)
}

// ExecuteWithTimeout 带超时的执行
func (e *GojaEngine) ExecuteWithTimeout(script string, input map[string]interface{}, timeout time.Duration) (*Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return e.executeWithContext(ctx, script, input)
}

// executeWithContext 内部执行实现
// 脚本必须定义 ScriptRun(input) 函数，引擎将调用该函数并返回结果
func (e *GojaEngine) executeWithContext(ctx context.Context, script string, input map[string]interface{}) (*Result, error) {
	result := &Result{
		Success: true,
		Logs:    make([]builtin.LogEntry, 0),
	}

	// 创建 goja runtime
	vm := goja.New()

	// 设置调用栈深度限制
	vm.SetMaxCallStackSize(e.config.MaxCallStackSize)

	// 设置超时检查
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt(ctx.Err())
		case <-done:
		}
	}()

	// 注入内置对象
	builtin.InjectBuiltins(vm, &result.Logs)

	// 注入全局变量
	for key, value := range e.globals {
		vm.Set(key, value)
	}

	// 注入全局函数
	for name, fn := range e.funcs {
		vm.Set(name, fn)
	}

	// 先执行脚本以定义 ScriptRun 函数
	_, err := vm.RunString(script)
	if err != nil {
		result.Success = false
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			result.Error = "script execution timeout"
		} else {
			result.Error = fmt.Sprintf("script error: %v", err)
		}
		return result, nil
	}

	// 检查 ScriptRun 函数是否存在
	scriptRunValue := vm.Get("ScriptRun")
	if scriptRunValue == nil || goja.IsUndefined(scriptRunValue) {
		result.Success = false
		result.Error = "script must define a ScriptRun function"
		return result, nil
	}

	// 调用 ScriptRun 函数
	scriptRun, ok := goja.AssertFunction(scriptRunValue)
	if !ok {
		result.Success = false
		result.Error = "ScriptRun must be a function"
		return result, nil
	}

	// 准备输入参数
	var inputArg interface{}
	if input != nil {
		inputArg = input
	} else {
		inputArg = make(map[string]interface{})
	}

	// 调用 ScriptRun(input)
	gojaValue, err := scriptRun(goja.Undefined(), vm.ToValue(inputArg))
	if err != nil {
		result.Success = false
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			result.Error = "script execution timeout"
		} else {
			result.Error = fmt.Sprintf("ScriptRun execution error: %v", err)
		}
		return result, nil
	}

	// 获取返回值
	result.Value = gojaValue.Export()

	return result, nil
}

// Validate 验证脚本语法和 ScriptRun 函数定义
func (e *GojaEngine) Validate(script string) error {
	// 编译检查语法
	program, err := goja.Compile("<validate>", script, false)
	if err != nil {
		return err
	}

	// 创建临时 VM 验证 ScriptRun 函数
	vm := goja.New()
	_, err = vm.RunProgram(program)
	if err != nil {
		return err
	}

	// 检查 ScriptRun 是否定义
	scriptRunValue := vm.Get("ScriptRun")
	if scriptRunValue == nil || goja.IsUndefined(scriptRunValue) {
		return fmt.Errorf("script must define a ScriptRun function")
	}

	// 检查 ScriptRun 是否为函数
	if _, ok := goja.AssertFunction(scriptRunValue); !ok {
		return fmt.Errorf("ScriptRun must be a function")
	}

	return nil
}

// SetGlobal 设置全局变量
func (e *GojaEngine) SetGlobal(key string, value interface{}) {
	e.globals[key] = value
}

// SetFunction 设置全局函数
func (e *GojaEngine) SetFunction(name string, fn interface{}) {
	e.funcs[name] = fn
}

// ToJSON 转换为 JSON 字符串
func (r *Result) ToJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}
