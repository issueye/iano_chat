package tools

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/cloudwego/eino/components/tool"
)

type Registry interface {
	Register(name string, t tool.InvokableTool) error
	Unregister(name string) error
	Get(name string) (tool.InvokableTool, bool)
	List() []tool.InvokableTool
	Names() []string
	Clear()
	Clone() Registry
	Merge(other Registry) error
}

type defaultRegistry struct {
	tools map[string]tool.InvokableTool
	mu    sync.RWMutex
}

func NewRegistry() Registry {
	return &defaultRegistry{
		tools: make(map[string]tool.InvokableTool),
	}
}

func NewRegistryWithTools(toolsMap map[string]tool.InvokableTool) Registry {
	r := &defaultRegistry{
		tools: make(map[string]tool.InvokableTool),
	}
	for name, t := range toolsMap {
		r.tools[name] = t
	}
	return r
}

func (r *defaultRegistry) Register(name string, t tool.InvokableTool) error {
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

func (r *defaultRegistry) Get(name string) (tool.InvokableTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tools[name]
	return t, exists
}

func (r *defaultRegistry) List() []tool.InvokableTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]tool.InvokableTool, 0, len(r.tools))
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

	r.tools = make(map[string]tool.InvokableTool)
}

func (r *defaultRegistry) Clone() Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	newRegistry := &defaultRegistry{
		tools: make(map[string]tool.InvokableTool),
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

	ddgTool, err := NewDuckDuckGoTool(30)
	if err != nil {
		return nil, fmt.Errorf("创建 DuckDuckGo 工具失败: %w", err)
	}
	toolsMap["web_search"] = ddgTool.tool

	httpTool := &HTTPClientTool{}
	toolsMap["http_request"] = httpTool

	basePath, _ := os.Getwd()
	toolsMap["file_read"] = NewFileReadTool(basePath)
	toolsMap["file_write"] = NewFileWriteTool(basePath)
	toolsMap["file_create"] = NewFileCreateTool(basePath)
	toolsMap["file_list"] = NewFileListTool(basePath)
	toolsMap["file_delete"] = NewFileDeleteTool(basePath)
	toolsMap["file_info"] = NewFileInfoTool(basePath)

	toolsMap["grep_search"] = NewGrepSearchTool(basePath)
	toolsMap["grep_replace"] = NewGrepReplaceTool(basePath)

	toolsMap["archive_create"] = NewArchiveCreateTool(basePath)
	toolsMap["archive_extract"] = NewArchiveExtractTool(basePath)

	cmdTool := NewCommandExecuteTool()
	toolsMap["command_execute"] = cmdTool
	toolsMap["shell_execute"] = NewShellExecuteTool()
	toolsMap["process_list"] = NewProcessListTool()

	toolsMap["env_get"] = NewEnvironmentGetTool()
	toolsMap["env_set"] = NewEnvironmentSetTool()
	toolsMap["system_info"] = NewSystemInfoTool()

	toolsMap["ping"] = NewPingTool()
	toolsMap["dns_lookup"] = NewDNSLookupTool()
	toolsMap["http_headers"] = NewHTTPHeadersTool()

	return toolsMap, nil
}

func RegisterBuiltinTools(ctx context.Context, workDir string, timeout int) error {
	ddgTool, err := NewDuckDuckGoTool(timeout)
	if err != nil {
		return fmt.Errorf("创建 DuckDuckGo 工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("web_search", ddgTool.tool); err != nil {
		return fmt.Errorf("注册 DuckDuckGo 工具失败: %w", err)
	}

	httpTool := &HTTPClientTool{}
	if err := GlobalRegistry.Register("http_request", httpTool); err != nil {
		return fmt.Errorf("注册 HTTP 工具失败: %w", err)
	}

	basePath := workDir
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	slog.Info("当前工作区", slog.String("workDir", basePath))

	if err := GlobalRegistry.Register("file_read", NewFileReadTool(basePath)); err != nil {
		return fmt.Errorf("注册文件读取工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("file_write", NewFileWriteTool(basePath)); err != nil {
		return fmt.Errorf("注册文件写入工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("file_create", NewFileCreateTool(basePath)); err != nil {
		return fmt.Errorf("注册文件创建工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("file_list", NewFileListTool(basePath)); err != nil {
		return fmt.Errorf("注册文件列表工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("file_delete", NewFileDeleteTool(basePath)); err != nil {
		return fmt.Errorf("注册文件删除工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("file_info", NewFileInfoTool(basePath)); err != nil {
		return fmt.Errorf("注册文件信息工具失败: %w", err)
	}

	if err := GlobalRegistry.Register("grep_search", NewGrepSearchTool(basePath)); err != nil {
		return fmt.Errorf("注册搜索工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("grep_replace", NewGrepReplaceTool(basePath)); err != nil {
		return fmt.Errorf("注册替换工具失败: %w", err)
	}

	if err := GlobalRegistry.Register("archive_create", NewArchiveCreateTool(basePath)); err != nil {
		return fmt.Errorf("注册压缩工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("archive_extract", NewArchiveExtractTool(basePath)); err != nil {
		return fmt.Errorf("注册解压工具失败: %w", err)
	}

	cmdTool := NewCommandExecuteTool()
	if err := GlobalRegistry.Register("command_execute", cmdTool); err != nil {
		return fmt.Errorf("注册命令执行工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("shell_execute", NewShellExecuteTool()); err != nil {
		return fmt.Errorf("注册 Shell 执行工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("process_list", NewProcessListTool()); err != nil {
		return fmt.Errorf("注册进程列表工具失败: %w", err)
	}

	if err := GlobalRegistry.Register("env_get", NewEnvironmentGetTool()); err != nil {
		return fmt.Errorf("注册环境变量获取工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("env_set", NewEnvironmentSetTool()); err != nil {
		return fmt.Errorf("注册环境变量设置工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("system_info", NewSystemInfoTool()); err != nil {
		return fmt.Errorf("注册系统信息工具失败: %w", err)
	}

	if err := GlobalRegistry.Register("ping", NewPingTool()); err != nil {
		return fmt.Errorf("注册 Ping 工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("dns_lookup", NewDNSLookupTool()); err != nil {
		return fmt.Errorf("注册 DNS 查询工具失败: %w", err)
	}
	if err := GlobalRegistry.Register("http_headers", NewHTTPHeadersTool()); err != nil {
		return fmt.Errorf("注册 HTTP 头工具失败: %w", err)
	}

	return nil
}

type ScopedRegistry struct {
	parent  Registry
	tools   map[string]tool.InvokableTool
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
		tools:   make(map[string]tool.InvokableTool),
		allowed: allowed,
	}
}

func (r *ScopedRegistry) Register(name string, t tool.InvokableTool) error {
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

func (r *ScopedRegistry) Get(name string) (tool.InvokableTool, bool) {
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

func (r *ScopedRegistry) List() []tool.InvokableTool {
	r.mu.RLock()
	seen := make(map[string]bool)
	list := make([]tool.InvokableTool, 0)

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
	r.tools = make(map[string]tool.InvokableTool)
}

func (r *ScopedRegistry) Clone() Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	newRegistry := &ScopedRegistry{
		parent:  r.parent,
		tools:   make(map[string]tool.InvokableTool),
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

// CreateToolsWithBasePath 根据指定的基础路径创建工具实例
// 用于限制文件操作工具的操作范围
func CreateToolsWithBasePath(basePath string, toolNames []string) []tool.InvokableTool {
	toolsList := make([]tool.InvokableTool, 0)
	allowedMap := make(map[string]bool)
	for _, name := range toolNames {
		allowedMap[name] = true
	}

	for _, name := range toolNames {
		var t tool.InvokableTool
		switch name {
		case "file_read":
			t = NewFileReadTool(basePath)
		case "file_write":
			t = NewFileWriteTool(basePath)
		case "file_create":
			t = NewFileCreateTool(basePath)
		case "file_list":
			t = NewFileListTool(basePath)
		case "file_delete":
			t = NewFileDeleteTool(basePath)
		case "file_info":
			t = NewFileInfoTool(basePath)
		case "grep_search":
			t = NewGrepSearchTool(basePath)
		case "grep_replace":
			t = NewGrepReplaceTool(basePath)
		case "archive_create":
			t = NewArchiveCreateTool(basePath)
		case "archive_extract":
			t = NewArchiveExtractTool(basePath)
		case "command_execute":
			t = NewCommandExecuteTool().WithWorkingDir(basePath)
		case "shell_execute":
			t = NewShellExecuteTool().WithWorkingDir(basePath)
		default:
			if globalT, ok := GlobalRegistry.Get(name); ok {
				t = globalT
			}
		}
		if t != nil {
			toolsList = append(toolsList, t)
		}
	}

	return toolsList
}
