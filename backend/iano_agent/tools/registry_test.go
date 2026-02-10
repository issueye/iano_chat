package tools

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// mockTool 模拟工具
type mockTool struct {
	name string
}

func (m *mockTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: m.name,
		Desc: "Mock tool for testing",
	}, nil
}

func (m *mockTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	return "mock result", nil
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	t.Run("注册成功", func(t *testing.T) {
		mock := &mockTool{name: "test_tool"}
		err := registry.Register("test_tool", mock)
		if err != nil {
			t.Errorf("Register() error = %v", err)
		}

		// 验证注册成功
		got, exists := registry.Get("test_tool")
		if !exists {
			t.Error("Get() should return exists = true")
		}
		if got == nil {
			t.Error("Get() should not return nil")
		}

		// 验证 Info
		info, err := got.Info(ctx)
		if err != nil {
			t.Errorf("Info() error = %v", err)
		}
		if info.Name != "test_tool" {
			t.Errorf("Info().Name = %v, want test_tool", info.Name)
		}
	})

	t.Run("重复注册", func(t *testing.T) {
		mock := &mockTool{name: "duplicate_tool"}
		registry.Register("duplicate_tool", mock)

		err := registry.Register("duplicate_tool", mock)
		if err == nil {
			t.Error("Register() should return error for duplicate")
		}
	})

	t.Run("空名称", func(t *testing.T) {
		mock := &mockTool{name: ""}
		err := registry.Register("", mock)
		if err == nil {
			t.Error("Register() should return error for empty name")
		}
	})

	t.Run("空工具", func(t *testing.T) {
		err := registry.Register("nil_tool", nil)
		if err == nil {
			t.Error("Register() should return error for nil tool")
		}
	})
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	t.Run("注销成功", func(t *testing.T) {
		mock := &mockTool{name: "tool_to_remove"}
		registry.Register("tool_to_remove", mock)

		err := registry.Unregister("tool_to_remove")
		if err != nil {
			t.Errorf("Unregister() error = %v", err)
		}

		// 验证已注销
		_, exists := registry.Get("tool_to_remove")
		if exists {
			t.Error("Get() should return exists = false after unregister")
		}
	})

	t.Run("注销不存在", func(t *testing.T) {
		err := registry.Unregister("non_existent")
		if err == nil {
			t.Error("Unregister() should return error for non-existent tool")
		}
	})
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// 清空并添加测试工具
	registry.Clear()

	tools := []string{"tool1", "tool2", "tool3"}
	for _, name := range tools {
		registry.Register(name, &mockTool{name: name})
	}

	list := registry.List()
	if len(list) != len(tools) {
		t.Errorf("List() returned %d tools, want %d", len(list), len(tools))
	}
}

func TestRegistry_Names(t *testing.T) {
	registry := NewRegistry()
	registry.Clear()

	tools := []string{"alpha", "beta", "gamma"}
	for _, name := range tools {
		registry.Register(name, &mockTool{name: name})
	}

	names := registry.Names()
	if len(names) != len(tools) {
		t.Errorf("Names() returned %d names, want %d", len(names), len(tools))
	}

	// 验证名称存在
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, expected := range tools {
		if !nameMap[expected] {
			t.Errorf("Names() missing %s", expected)
		}
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	registry.Register("tool_a", &mockTool{name: "tool_a"})
	registry.Register("tool_b", &mockTool{name: "tool_b"})

	registry.Clear()

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("List() should return empty after Clear(), got %d", len(list))
	}
}

func TestRegisterBuiltinTools(t *testing.T) {
	// 清理全局注册表
	GlobalRegistry.Clear()

	ctx := context.Background()
	err := RegisterBuiltinTools(ctx)
	if err != nil {
		t.Errorf("RegisterBuiltinTools() error = %v", err)
	}

	// 验证内置工具已注册
	tools := []string{"web_search", "http_request"}
	for _, name := range tools {
		_, exists := GlobalRegistry.Get(name)
		if !exists {
			t.Errorf("Builtin tool %s should be registered", name)
		}
	}
}

func TestGlobalRegistry(t *testing.T) {
	// 测试全局注册表是单例
	reg1 := GlobalRegistry
	reg2 := GlobalRegistry

	// 添加工具到 reg1
	reg1.Register("global_test", &mockTool{name: "global_test"})

	// 从 reg2 获取
	_, exists := reg2.Get("global_test")
	if !exists {
		t.Error("GlobalRegistry should be singleton")
	}

	// 清理
	reg1.Unregister("global_test")
}
