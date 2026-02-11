package script_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type Engine struct {
	mu         sync.RWMutex
	timeout    time.Duration
	maxScripts int
}

type Config struct {
	Timeout    time.Duration
	MaxScripts int
}

func NewEngine(cfg *Config) *Engine {
	if cfg == nil {
		cfg = &Config{
			Timeout:    30 * time.Second,
			MaxScripts: 100,
		}
	}

	return &Engine{
		timeout:    cfg.Timeout,
		maxScripts: cfg.MaxScripts,
	}
}

type Result struct {
	Value interface{}
	Error string
}

func (e *Engine) Execute(ctx context.Context, script string, params map[string]interface{}) (*Result, error) {
	vm := goja.New()

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	console := vm.NewObject()
	_ = vm.Set("console", console)
	_ = console.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Printf("[Script Log] %v\n", args)
		return goja.Undefined()
	})
	_ = console.Set("error", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Printf("[Script Error] %v\n", args)
		return goja.Undefined()
	})

	_, err = vm.RunString("var params = " + string(paramsJSON) + ";")
	if err != nil {
		return nil, fmt.Errorf("failed to set params: %w", err)
	}

	result := &Result{}

	type scriptDone struct {
		value interface{}
		err   error
	}
	done := make(chan scriptDone, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- scriptDone{nil, fmt.Errorf("script panic: %v", r)}
			}
		}()

		value, err := vm.RunString(script)
		if err != nil {
			done <- scriptDone{nil, err}
			return
		}

		done <- scriptDone{value.Export(), nil}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("script execution cancelled")
	case <-time.After(e.timeout):
		vm.Interrupt("execution timeout")
		return nil, fmt.Errorf("script execution timeout after %v", e.timeout)
	case d := <-done:
		if d.err != nil {
			result.Error = d.err.Error()
			return result, d.err
		}
		result.Value = d.value
	}

	return result, nil
}

func (e *Engine) ExecuteWithStringResult(ctx context.Context, script string, params map[string]interface{}) (string, error) {
	result, err := e.Execute(ctx, script, params)
	if err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("script error: %s", result.Error)
	}

	switch v := result.Value.(type) {
	case string:
		return v, nil
	case goja.Value:
		return v.String(), nil
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(jsonBytes), nil
	}
}
