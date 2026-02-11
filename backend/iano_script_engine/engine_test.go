package script_engine

import (
	"context"
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		engine := NewEngine(nil)
		if engine == nil {
			t.Fatal("engine should not be nil")
		}
		if engine.timeout != 30*time.Second {
			t.Errorf("expected timeout 30s, got %v", engine.timeout)
		}
	})

	t.Run("custom config", func(t *testing.T) {
		engine := NewEngine(&Config{
			Timeout:    10 * time.Second,
			MaxScripts: 50,
		})
		if engine.timeout != 10*time.Second {
			t.Errorf("expected timeout 10s, got %v", engine.timeout)
		}
	})
}

func TestEngine_Execute(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	tests := []struct {
		name      string
		script    string
		params    map[string]interface{}
		wantErr   bool
		wantValue interface{}
	}{
		{
			name:      "simple return",
			script:    `"hello world"`,
			params:    nil,
			wantErr:   false,
			wantValue: "hello world",
		},
		{
			name:      "use params",
			script:    `params.name + " is " + params.age + " years old"`,
			params:    map[string]interface{}{"name": "John", "age": 30},
			wantErr:   false,
			wantValue: "John is 30 years old",
		},
		{
			name:      "arithmetic",
			script:    `2 + 2`,
			params:    nil,
			wantErr:   false,
			wantValue: int64(4),
		},
		{
			name:    "object return",
			script:  `({result: params.a + params.b})`,
			params:  map[string]interface{}{"a": 1, "b": 2},
			wantErr: false,
		},
		{
			name:    "syntax error",
			script:  `invalid javascript {{`,
			params:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Execute(ctx, tt.script, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantValue != nil && result.Value != tt.wantValue {
				t.Errorf("Execute() value = %v, want %v", result.Value, tt.wantValue)
			}
		})
	}
}

func TestEngine_ExecuteWithStringResult(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		script  string
		params  map[string]interface{}
		wantErr bool
		wantStr string
	}{
		{
			name:    "string result",
			script:  `"hello"`,
			params:  nil,
			wantErr: false,
			wantStr: "hello",
		},
		{
			name:    "number result",
			script:  `42`,
			params:  nil,
			wantErr: false,
			wantStr: "42",
		},
		{
			name:    "object result as json",
			script:  `({a: 1, b: 2})`,
			params:  nil,
			wantErr: false,
			wantStr: `{"a":1,"b":2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.ExecuteWithStringResult(ctx, tt.script, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWithStringResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.wantStr {
				t.Errorf("ExecuteWithStringResult() = %v, want %v", result, tt.wantStr)
			}
		})
	}
}

func TestEngine_Timeout(t *testing.T) {
	engine := NewEngine(&Config{
		Timeout: 100 * time.Millisecond,
	})
	ctx := context.Background()

	_, err := engine.Execute(ctx, `while(true) {}`, nil)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestEngine_Console(t *testing.T) {
	engine := NewEngine(nil)
	ctx := context.Background()

	_, err := engine.Execute(ctx, `console.log("test"); "done"`, nil)
	if err != nil {
		t.Errorf("console.log caused error: %v", err)
	}
}

func TestEngine_ContextCancellation(t *testing.T) {
	engine := NewEngine(&Config{
		Timeout: 10 * time.Second,
	})
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := engine.Execute(ctx, `while(true) {}`, nil)
	if err == nil {
		t.Error("expected cancellation error")
	}
}
