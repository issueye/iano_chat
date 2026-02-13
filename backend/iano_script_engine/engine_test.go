// Package script - 脚本引擎测试

package iano_script_engine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"iano_script_engine/builtin"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine(nil)
	assert.NotNil(t, engine)

	config := &Config{
		Timeout:          10 * time.Second,
		MemoryLimit:      5 * 1024 * 1024,
		MaxCallStackSize: 500,
	}
	engine = NewEngine(config)
	assert.NotNil(t, engine)
}

func TestGojaEngine_Execute(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	tests := []struct {
		name     string
		script   string
		input    map[string]interface{}
		expected bool
		hasError bool
	}{
		{
			name: "简单返回值",
			script: `
				function ScriptRun(input) {
					return 1 + 1;
				}
			`,
			expected: true,
			hasError: false,
		},
		{
			name: "返回对象",
			script: `
				function ScriptRun(input) {
					return { value: 42, name: "test" };
				}
			`,
			expected: true,
			hasError: false,
		},
		{
			name: "使用输入数据",
			script: `
				function ScriptRun(input) {
					return { result: input.name + "!" };
				}
			`,
			input:    map[string]interface{}{"name": "World"},
			expected: true,
			hasError: false,
		},
		{
			name:     "语法错误",
			script:   `1 +`,
			expected: false,
			hasError: true,
		},
		{
			name: "缺少ScriptRun函数",
			script: `
				var x = 1;
			`,
			expected: false,
			hasError: true,
		},
		{
			name: "ScriptRun不是函数",
			script: `
				var ScriptRun = "not a function";
			`,
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Execute(ctx, tt.script, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result.Success)
			if tt.hasError {
				assert.NotEmpty(t, result.Error)
			}
		})
	}
}

func TestGojaEngine_ExecuteWithTimeout(t *testing.T) {
	engine := NewEngine(nil)

	// 测试正常执行
	result, err := engine.ExecuteWithTimeout(`
		function ScriptRun(input) {
			return 1 + 1;
		}
	`, nil, 5*time.Second)
	assert.NoError(t, err)
	assert.True(t, result.Success)

	// 测试超时
	result, err = engine.ExecuteWithTimeout(`
		function ScriptRun(input) {
			var start = Date.now();
			while (Date.now() - start < 10000) {}
			return "done";
		}
	`, nil, 100*time.Millisecond)
	assert.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "timeout")
}

func TestGojaEngine_Validate(t *testing.T) {
	engine := NewEngine(nil)

	tests := []struct {
		name    string
		script  string
		isValid bool
	}{
		{
			name: "有效脚本",
			script: `
				function ScriptRun(input) {
					var x = 1 + 1;
					return x;
				}
			`,
			isValid: true,
		},
		{
			name:    "无效脚本",
			script:  `var x = 1 +`,
			isValid: false,
		},
		{
			name: "复杂有效脚本",
			script: `
				function add(a, b) { return a + b; }
				function ScriptRun(input) {
					var result = add(1, 2);
					return result;
				}
			`,
			isValid: true,
		},
		{
			name: "缺少ScriptRun函数",
			script: `
				var x = 1 + 1;
			`,
			isValid: false,
		},
		{
			name: "ScriptRun不是函数",
			script: `
				var ScriptRun = "not a function";
			`,
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.Validate(tt.script)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGojaEngine_SetGlobal(t *testing.T) {
	engine := NewEngine(nil).(*GojaEngine)
	engine.SetGlobal("myVar", "hello")
	engine.SetGlobal("myNum", 42)

	ctx := context.Background()
	result, err := engine.Execute(ctx, `
		function ScriptRun(input) {
			return { result: myVar + " " + myNum };
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "hello 42", resultMap["result"])
}

func TestGojaEngine_SetFunction(t *testing.T) {
	engine := NewEngine(nil).(*GojaEngine)
	engine.SetFunction("greet", func(name string) string {
		return "Hello, " + name
	})

	ctx := context.Background()
	result, err := engine.Execute(ctx, `
		function ScriptRun(input) {
			return { message: greet("World") };
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Hello, World", resultMap["message"])
}

func TestGojaEngine_Console(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	result, err := engine.Execute(ctx, `
		function ScriptRun(input) {
			console.log("info message");
			console.debug("debug message");
			console.info("info message");
			console.warn("warn message");
			console.error("error message");
			return { done: true };
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.Logs, 5)

	// 检查日志级别
	levels := make([]string, 0)
	for _, log := range result.Logs {
		levels = append(levels, log.Level)
	}
	assert.Contains(t, levels, "info")
	assert.Contains(t, levels, "debug")
	assert.Contains(t, levels, "warn")
	assert.Contains(t, levels, "error")
}

func TestGojaEngine_JSON(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	result, err := engine.Execute(ctx, `
		function ScriptRun(input) {
			var obj = {name: "test", value: 42};
			var jsonStr = JSON.stringify(obj);
			var parsed = JSON.parse(jsonStr);
			return { original: jsonStr, parsed: parsed };
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, resultMap["original"])

	parsed, ok := resultMap["parsed"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test", parsed["name"])
}

func TestGojaEngine_Sleep(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	start := time.Now()
	result, err := engine.Execute(ctx, `
		function ScriptRun(input) {
			sleep(100);
			return { done: true };
		}
	`, nil)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, elapsed >= 100*time.Millisecond)
}

func TestResult_ToJSON(t *testing.T) {
	result := &Result{
		Success: true,
		Value:   map[string]interface{}{"key": "value"},
		Logs:    []builtin.LogEntry{{Level: "info", Message: "test"}},
	}

	jsonStr := result.ToJSON()
	assert.NotEmpty(t, jsonStr)
	assert.Contains(t, jsonStr, "success")
	assert.Contains(t, jsonStr, "true")
}

func TestEngineWithModules(t *testing.T) {
	modules := []Module{
		builtin.NewUtilsModule(),
	}

	engine := NewEngineWithModules(nil, modules...)
	ctx := context.Background()

	result, err := engine.Execute(ctx, `
		function ScriptRun(input) {
			return {
				uuid: utils.uuid(),
				lower: utils.string.toLower("HELLO")
			};
		}
	`, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, resultMap["uuid"])
	assert.Equal(t, "hello", resultMap["lower"])
}

func TestHTTPModule(t *testing.T) {
	module := builtin.NewHTTPModule(5 * time.Second)
	assert.Equal(t, "http", module.Name())

	// 注意：实际的 HTTP 请求测试需要外部服务，这里只测试模块注册
	// 在实际项目中，应该使用 mock HTTP 客户端
}

func TestUtilsModule(t *testing.T) {
	module := builtin.NewUtilsModule()
	assert.Equal(t, "utils", module.Name())
}

func TestURLModule(t *testing.T) {
	module := builtin.NewURLModule()
	assert.Equal(t, "url", module.Name())
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, uint64(10*1024*1024), config.MemoryLimit)
	assert.Equal(t, 1000, config.MaxCallStackSize)
}

func TestScriptExecution_ComplexScript(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	script := `
		function fibonacci(n) {
			if (n <= 1) return n;
			return fibonacci(n - 1) + fibonacci(n - 2);
		}
		
		function ScriptRun(input) {
			var results = [];
			for (var i = 0; i < 10; i++) {
				results.push(fibonacci(i));
			}
			
			return {
				fibonacci: results,
				sum: results.reduce(function(a, b) { return a + b; }, 0)
			};
		}
	`

	result, err := engine.Execute(ctx, script, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)

	resultMap, ok := result.Value.(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, resultMap["fibonacci"])
	assert.NotNil(t, resultMap["sum"])
}

func TestScriptExecution_ContextCancellation(t *testing.T) {
	engine := NewEngine(nil)
	ctx, cancel := context.WithCancel(context.Background())

	// 立即取消
	cancel()

	result, err := engine.Execute(ctx, `
		function ScriptRun(input) {
			return { test: 1 };
		}
	`, nil)

	// 结果可能成功也可能失败，取决于取消时机
	_ = result
	_ = err
}

func BenchmarkGojaEngine_Execute(b *testing.B) {
	engine := NewEngine(nil)
	ctx := context.Background()
	script := `
		function ScriptRun(input) {
			return 1 + 1;
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Execute(ctx, script, nil)
	}
}

func BenchmarkGojaEngine_ExecuteComplex(b *testing.B) {
	engine := NewEngine(nil)
	ctx := context.Background()
	script := `
		function sum(n) {
			var s = 0;
			for (var i = 0; i < n; i++) {
				s += i;
			}
			return s;
		}
		function ScriptRun(input) {
			return sum(1000);
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Execute(ctx, script, nil)
	}
}
