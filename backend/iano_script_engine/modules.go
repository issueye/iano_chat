// Package script - 脚本模块扩展
// 提供丰富的内置模块供 JavaScript 脚本使用

package iano_script_engine

import (
	"context"
	"errors"
	"fmt"

	"github.com/dop251/goja"

	"iano_script_engine/builtin"
)

// Module 脚本模块接口（别名，兼容旧代码）
type Module = builtin.Module

// EngineWithModules 带模块的引擎
type EngineWithModules struct {
	*GojaEngine
	modules []Module
}

// NewEngineWithModules 创建带模块的引擎
func NewEngineWithModules(config *Config, modules ...Module) *EngineWithModules {
	return &EngineWithModules{
		GojaEngine: NewEngine(config).(*GojaEngine),
		modules:    modules,
	}
}

// Execute 执行脚本（带模块）
func (e *EngineWithModules) Execute(ctx context.Context, script string, input map[string]interface{}) (*Result, error) {
	return e.executeWithModules(ctx, script, input)
}

// executeWithModules 内部执行实现
// 脚本必须定义 ScriptRun(input) 函数，引擎将调用该函数并返回结果
func (e *EngineWithModules) executeWithModules(ctx context.Context, script string, input map[string]interface{}) (*Result, error) {
	result := &Result{
		Success: true,
		Logs:    make([]builtin.LogEntry, 0),
	}

	// 创建 goja runtime
	vm := goja.New()
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

	// 注入全局变量和函数
	for key, value := range e.globals {
		vm.Set(key, value)
	}
	for name, fn := range e.funcs {
		vm.Set(name, fn)
	}

	// 注入模块
	for _, module := range e.modules {
		if err := module.Register(vm); err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to register module %s: %v", module.Name(), err)
			return result, nil
		}
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
