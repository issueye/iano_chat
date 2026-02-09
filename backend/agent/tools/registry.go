package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/tool"
)

// Registry 工具注册表接口
type Registry interface {
	// Register 注册工具
	Register(name string, t tool.BaseTool) error
	// Unregister 注销工具
	Unregister(name string) error
	// Get 获取工具
	Get(name string) (tool.BaseTool, bool)
	// List 列出所有工具
	List() []tool.BaseTool
	// Names 获取所有工具名称
	Names() []string
	// Clear 清空所有工具
	Clear()
}

// defaultRegistry 默认工具注册表实现
type defaultRegistry struct {
	tools map[string]tool.BaseTool
	mu    sync.RWMutex
}

// NewRegistry 创建新的工具注册表
func NewRegistry() Registry {
	return &defaultRegistry{
		tools: make(map[string]tool.BaseTool),
	}
}

// Register 注册工具
func (r *defaultRegistry) Register(name string, t tool.BaseTool) error {
	if name == "" {
		return fmt.Errorf("工具名称不能为空")
	}
	if t == nil {
		return fmt.Errorf("工具实例不能为空")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("工具 '%s' 已存在", name)
	}

	r.tools[name] = t
	return nil
}

// Unregister 注销工具
func (r *defaultRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("工具 '%s' 不存在", name)
	}

	delete(r.tools, name)
	return nil
}

// Get 获取工具
func (r *defaultRegistry) Get(name string) (tool.BaseTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tools[name]
	return t, exists
}

// List 列出所有工具
func (r *defaultRegistry) List() []tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]tool.BaseTool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

// Names 获取所有工具名称
func (r *defaultRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Clear 清空所有工具
func (r *defaultRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]tool.BaseTool)
}

// GlobalRegistry 全局工具注册表实例
var GlobalRegistry = NewRegistry()

// RegisterBuiltinTools 注册内置工具
func RegisterBuiltinTools(ctx context.Context) error {
	// 注册 DuckDuckGo 搜索工具
	ddgTool, err := NewDuckDuckGoTool()
	if err != nil {
		return fmt.Errorf("创建 DuckDuckGo 工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("web_search", ddgTool); err != nil {
		return fmt.Errorf("注册 DuckDuckGo 工具失败: %w", err)
	}

	// 注册 HTTP 客户端工具
	httpTool := &HTTPClientTool{}
	if err := GlobalRegistry.Register("http_request", httpTool); err != nil {
		return fmt.Errorf("注册 HTTP 工具失败: %w", err)
	}

	return nil
}
