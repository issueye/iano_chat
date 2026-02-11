package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/tool"
)

type Registry interface {
	Register(name string, t tool.BaseTool) error
	Unregister(name string) error
	Get(name string) (tool.BaseTool, bool)
	List() []tool.BaseTool
	Names() []string
	Clear()
	Clone() Registry
	Merge(other Registry) error
}

type defaultRegistry struct {
	tools map[string]tool.BaseTool
	mu    sync.RWMutex
}

func NewRegistry() Registry {
	return &defaultRegistry{
		tools: make(map[string]tool.BaseTool),
	}
}

func NewRegistryWithTools(toolsMap map[string]tool.BaseTool) Registry {
	r := &defaultRegistry{
		tools: make(map[string]tool.BaseTool),
	}
	for name, t := range toolsMap {
		r.tools[name] = t
	}
	return r
}

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

func (r *defaultRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("工具 '%s' 不存在", name)
	}

	delete(r.tools, name)
	return nil
}

func (r *defaultRegistry) Get(name string) (tool.BaseTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tools[name]
	return t, exists
}

func (r *defaultRegistry) List() []tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]tool.BaseTool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

func (r *defaultRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

func (r *defaultRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools = make(map[string]tool.BaseTool)
}

func (r *defaultRegistry) Clone() Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	newRegistry := &defaultRegistry{
		tools: make(map[string]tool.BaseTool),
	}
	for name, t := range r.tools {
		newRegistry.tools[name] = t
	}
	return newRegistry
}

func (r *defaultRegistry) Merge(other Registry) error {
	for _, name := range other.Names() {
		t, ok := other.Get(name)
		if !ok {
			continue
		}
		if err := r.Register(name, t); err != nil {
			return err
		}
	}
	return nil
}

var GlobalRegistry = NewRegistry()

func GetBuiltinTools(ctx context.Context) (map[string]tool.BaseTool, error) {
	toolsMap := make(map[string]tool.BaseTool)

	ddgTool, err := NewDuckDuckGoTool()
	if err != nil {
		return nil, fmt.Errorf("创建 DuckDuckGo 工具失败: %w", err)
	}
	toolsMap["web_search"] = ddgTool

	httpTool := &HTTPClientTool{}
	toolsMap["http_request"] = httpTool

	return toolsMap, nil
}

func RegisterBuiltinTools(ctx context.Context) error {
	ddgTool, err := NewDuckDuckGoTool()
	if err != nil {
		return fmt.Errorf("创建 DuckDuckGo 工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("web_search", ddgTool); err != nil {
		return fmt.Errorf("注册 DuckDuckGo 工具失败: %w", err)
	}

	httpTool := &HTTPClientTool{}
	if err := GlobalRegistry.Register("http_request", httpTool); err != nil {
		return fmt.Errorf("注册 HTTP 工具失败: %w", err)
	}

	return nil
}

type ScopedRegistry struct {
	parent  Registry
	tools   map[string]tool.BaseTool
	mu      sync.RWMutex
	allowed map[string]bool
}

func NewScopedRegistry(parent Registry, allowedTools []string) *ScopedRegistry {
	allowed := make(map[string]bool)
	for _, name := range allowedTools {
		allowed[name] = true
	}
	return &ScopedRegistry{
		parent:  parent,
		tools:   make(map[string]tool.BaseTool),
		allowed: allowed,
	}
}

func (r *ScopedRegistry) Register(name string, t tool.BaseTool) error {
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
	r.allowed[name] = true
	return nil
}

func (r *ScopedRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		if r.parent != nil {
			return fmt.Errorf("不能注销父注册表的工具 '%s'", name)
		}
		return fmt.Errorf("工具 '%s' 不存在", name)
	}

	delete(r.tools, name)
	delete(r.allowed, name)
	return nil
}

func (r *ScopedRegistry) Get(name string) (tool.BaseTool, bool) {
	r.mu.RLock()
	if t, exists := r.tools[name]; exists {
		r.mu.RUnlock()
		return t, true
	}
	r.mu.RUnlock()

	if r.parent != nil && r.isAllowed(name) {
		return r.parent.Get(name)
	}
	return nil, false
}

func (r *ScopedRegistry) isAllowed(name string) bool {
	if len(r.allowed) == 0 {
		return true
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.allowed[name]
}

func (r *ScopedRegistry) List() []tool.BaseTool {
	r.mu.RLock()
	seen := make(map[string]bool)
	list := make([]tool.BaseTool, 0)

	for name, t := range r.tools {
		if !seen[name] {
			list = append(list, t)
			seen[name] = true
		}
	}
	r.mu.RUnlock()

	if r.parent != nil {
		for _, t := range r.parent.List() {
			info, _ := t.Info(context.Background())
			if info != nil && !seen[info.Name] && r.isAllowed(info.Name) {
				list = append(list, t)
				seen[info.Name] = true
			}
		}
	}

	return list
}

func (r *ScopedRegistry) Names() []string {
	r.mu.RLock()
	seen := make(map[string]bool)
	names := make([]string, 0)

	for name := range r.tools {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	r.mu.RUnlock()

	if r.parent != nil {
		for _, name := range r.parent.Names() {
			if !seen[name] && r.isAllowed(name) {
				names = append(names, name)
				seen[name] = true
			}
		}
	}

	return names
}

func (r *ScopedRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = make(map[string]tool.BaseTool)
}

func (r *ScopedRegistry) Clone() Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	newRegistry := &ScopedRegistry{
		parent:  r.parent,
		tools:   make(map[string]tool.BaseTool),
		allowed: make(map[string]bool),
	}
	for name, t := range r.tools {
		newRegistry.tools[name] = t
	}
	for name := range r.allowed {
		newRegistry.allowed[name] = true
	}
	return newRegistry
}

func (r *ScopedRegistry) Merge(other Registry) error {
	for _, name := range other.Names() {
		t, ok := other.Get(name)
		if !ok {
			continue
		}
		if err := r.Register(name, t); err != nil {
			return err
		}
	}
	return nil
}
