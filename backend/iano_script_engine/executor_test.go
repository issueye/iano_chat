// Package script - 脚本执行器测试

package iano_script_engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"iano_script_engine/builtin"
)

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor(nil)
	assert.NotNil(t, executor)

	config := &ExecutorConfig{
		DefaultTimeout: 60 * time.Second,
		MaxTimeout:     10 * time.Minute,
		EnableHTTP:     true,
		EnableUtils:    true,
		EnableURL:      true,
	}
	executor = NewExecutor(config)
	assert.NotNil(t, executor)
}

func TestDefaultExecutorConfig(t *testing.T) {
	config := DefaultExecutorConfig()
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 5*time.Minute, config.MaxTimeout)
	assert.True(t, config.EnableHTTP)
	assert.True(t, config.EnableUtils)
	assert.True(t, config.EnableURL)
}

func TestScriptExecutor_Execute(t *testing.T) {
	executor := NewExecutor(nil)
	ctx := context.Background()

	result, err := executor.Execute(ctx, `
		function ScriptRun(input) {
			return { value: 42 };
		}
	`, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)
	assert.GreaterOrEqual(t, result.Duration, int64(0))

	resultMap, ok := result.Result.(map[string]interface{})
	assert.True(t, ok)
	// goja 导出数字可能是 int64 或 float64
	value := resultMap["value"]
	assert.True(t, value == float64(42) || value == int64(42))
}

func TestScriptExecutor_ExecuteWithTimeout(t *testing.T) {
	executor := NewExecutor(nil)

	result, err := executor.ExecuteWithTimeout(`
		function ScriptRun(input) {
			return { value: 42 };
		}
	`, nil, 5*time.Second)

	assert.NoError(t, err)
	assert.True(t, result.Success)
}

func TestScriptExecutor_Validate(t *testing.T) {
	executor := NewExecutor(nil)

	err := executor.Validate(`
		function ScriptRun(input) {
			var x = 1 + 1;
			return x;
		}
	`)
	assert.NoError(t, err)

	err = executor.Validate(`var x = 1 +`)
	assert.Error(t, err)
}

func TestScriptExecutor_SetGlobal(t *testing.T) {
	executor := NewExecutor(nil).(*ScriptExecutor)
	executor.SetGlobal("globalVar", "test_value")

	ctx := context.Background()
	result, err := executor.Execute(ctx, `
		function ScriptRun(input) {
			return { result: globalVar };
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test_value", resultMap["result"])
}

func TestScriptExecutor_SetFunction(t *testing.T) {
	executor := NewExecutor(nil).(*ScriptExecutor)
	executor.SetFunction("double", func(x int) int {
		return x * 2
	})

	ctx := context.Background()
	result, err := executor.Execute(ctx, `
		function ScriptRun(input) {
			return { result: double(21) };
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, int64(42), resultMap["result"])
}

func TestExecutionResult_ToJSON(t *testing.T) {
	result := &ExecutionResult{
		Success:  true,
		Result:   map[string]interface{}{"key": "value"},
		Error:    "",
		Logs:     []builtin.LogEntry{{Level: "info", Message: "test"}},
		Duration: 100,
	}

	jsonStr := result.ToJSON()
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "success")
	assert.Contains(t, jsonStr, "true")
	assert.Contains(t, jsonStr, "key")
}

func TestNewHookScriptExecutor(t *testing.T) {
	executor := NewHookScriptExecutor()
	assert.NotNil(t, executor)
}

func TestHookScriptExecutor_ExecuteHook(t *testing.T) {
	executor := NewHookScriptExecutor()
	ctx := context.Background()

	result, err := executor.ExecuteHook(ctx, `
		function ScriptRun(input) {
			console.log("Event:", input.event);
			return { processed: true };
		}
	`, "test.event", map[string]interface{}{
		"key": "value",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestNewAgentScriptExecutor(t *testing.T) {
	executor := NewAgentScriptExecutor()
	assert.NotNil(t, executor)
}

func TestAgentScriptExecutor_ExecuteTool(t *testing.T) {
	executor := NewAgentScriptExecutor()
	ctx := context.Background()

	result, err := executor.ExecuteTool(ctx, `
		function ScriptRun(input) {
			return { sum: input.args.a + input.args.b };
		}
	`, "calculator", map[string]interface{}{
		"a": 10,
		"b": 20,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	resultMap, ok := result.Result.(map[string]interface{})
	assert.True(t, ok)
	// goja 导出数字可能是 int64 或 float64
	sum := resultMap["sum"]
	assert.True(t, sum == float64(30) || sum == int64(30))
}

func TestAgentScriptExecutor_ExecuteTransform(t *testing.T) {
	executor := NewAgentScriptExecutor()
	ctx := context.Background()

	result, err := executor.ExecuteTransform(ctx, `
		function ScriptRun(input) {
			return { upper: input.data.toUpperCase() };
		}
	`, "hello")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)

	resultMap, ok := result.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "HELLO", resultMap["upper"])
}

func TestAgentScriptExecutor_ExecuteFilter(t *testing.T) {
	executor := NewAgentScriptExecutor()
	ctx := context.Background()

	tests := []struct {
		script   string
		expected bool
	}{
		{
			script: `
				function ScriptRun(input) {
					return { match: input.item.length > 5 };
				}
			`,
			expected: true,
		},
		{
			script: `
				function ScriptRun(input) {
					return { match: false };
				}
			`,
			expected: false,
		},
		{
			// 直接返回布尔值
			script: `
				function ScriptRun(input) {
					return true;
				}
			`,
			expected: true,
		},
	}

	for _, tt := range tests {
		match, err := executor.ExecuteFilter(ctx, tt.script, "hello world")
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, match)
	}
}

func TestDefaultSandboxLimits(t *testing.T) {
	limits := DefaultSandboxLimits()
	assert.Equal(t, 5*time.Second, limits.MaxExecutionTime)
	assert.Equal(t, int64(50), limits.MaxMemoryMB)
	assert.Equal(t, 1024*1024, limits.MaxOutputSize)
	assert.NotEmpty(t, limits.AllowedModules)
	assert.NotEmpty(t, limits.BlockedFunctions)
}

func TestNewSandbox(t *testing.T) {
	sandbox := NewSandbox(nil)
	assert.NotNil(t, sandbox)

	limits := &SandboxLimits{
		MaxExecutionTime: 3 * time.Second,
		MaxMemoryMB:      30,
		AllowedModules:   []string{"utils"},
		BlockedFunctions: []string{"http"},
	}
	sandbox = NewSandbox(limits)
	assert.NotNil(t, sandbox)
}

func TestSandbox_Run(t *testing.T) {
	sandbox := NewSandbox(nil)
	ctx := context.Background()

	result, err := sandbox.Run(ctx, `
		function ScriptRun(input) {
			return { value: 42 };
		}
	`, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestSandbox_RunBlockedFunction(t *testing.T) {
	limits := &SandboxLimits{
		MaxExecutionTime: 5 * time.Second,
		BlockedFunctions: []string{"http"},
	}
	sandbox := NewSandbox(limits)
	ctx := context.Background()

	result, err := sandbox.Run(ctx, `
		function ScriptRun(input) {
			http.get("http://example.com");
			return { done: true };
		}
	`, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "blocked")
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.True(t, contains(slice, "a"))
	assert.True(t, contains(slice, "b"))
	assert.True(t, contains(slice, "c"))
	assert.False(t, contains(slice, "d"))
	assert.False(t, contains([]string{}, "a"))
}

func TestExecutorWithModules_HTTP(t *testing.T) {
	config := &ExecutorConfig{
		DefaultTimeout: 30 * time.Second,
		EnableHTTP:     true,
		EnableUtils:    true,
		EnableURL:      true,
	}
	executor := NewExecutor(config)
	ctx := context.Background()

	// 测试 HTTP 模块函数存在（不实际发送请求）
	result, err := executor.Execute(ctx, `
		function ScriptRun(input) {
			return {
				hasGet: typeof http.get === 'function',
				hasPost: typeof http.post === 'function'
			};
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, resultMap["hasGet"])
	assert.Equal(t, true, resultMap["hasPost"])
}

func TestExecutorWithModules_Utils(t *testing.T) {
	config := &ExecutorConfig{
		DefaultTimeout: 30 * time.Second,
		EnableHTTP:     false,
		EnableUtils:    true,
		EnableURL:      false,
	}
	executor := NewExecutor(config)
	ctx := context.Background()

	result, err := executor.Execute(ctx, `
		function ScriptRun(input) {
			return {
				uuid: utils.uuid(),
				lower: utils.string.toLower("HELLO"),
				upper: utils.string.toUpper("hello")
			};
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, resultMap["uuid"])
	assert.Equal(t, "hello", resultMap["lower"])
	assert.Equal(t, "HELLO", resultMap["upper"])
}

func TestExecutorWithModules_URL(t *testing.T) {
	config := &ExecutorConfig{
		DefaultTimeout: 30 * time.Second,
		EnableHTTP:     false,
		EnableUtils:    false,
		EnableURL:      true,
	}
	executor := NewExecutor(config)
	ctx := context.Background()

	result, err := executor.Execute(ctx, `
		function ScriptRun(input) {
			var parsed = url.parse("https://example.com/path?query=value");
			return {
				scheme: parsed.scheme,
				host: parsed.host,
				path: parsed.path
			};
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "https", resultMap["scheme"])
	assert.Equal(t, "example.com", resultMap["host"])
	assert.Equal(t, "/path", resultMap["path"])
}

func TestExecutor_ErrorHandling(t *testing.T) {
	executor := NewExecutor(nil)
	ctx := context.Background()

	// 测试脚本错误
	result, err := executor.Execute(ctx, `
		function ScriptRun(input) {
			throw new Error("test error");
		}
	`, nil)
	assert.NoError(t, err) // 执行器本身不返回错误
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)
}

func TestExecutor_Timeout(t *testing.T) {
	executor := NewExecutor(nil)
	ctx := context.Background()

	// 使用超时的 context
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	result, err := executor.Execute(ctx, `
		function ScriptRun(input) {
			var start = Date.now();
			while (Date.now() - start < 10000) {}
			return { done: true };
		}
	`, nil)

	assert.NoError(t, err)
	assert.False(t, result.Success)
	// 错误消息可能是 "timeout" 或 "context deadline exceeded"
	assert.True(t,
		result.Error != "" &&
			(result.Error == "script execution timeout" ||
				len(result.Error) > 0))
}
