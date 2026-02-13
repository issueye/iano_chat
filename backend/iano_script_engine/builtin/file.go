// Package builtin - 文件访问模块
// 提供文件读写、目录操作等功能

package builtin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
)

// FileModule 文件访问模块
type FileModule struct {
	allowedDirs []string
	maxFileSize int64
	readOnly    bool
}

// FileModuleConfig 文件模块配置
type FileModuleConfig struct {
	// AllowedDirs 允许访问的目录列表（空表示允许所有）
	AllowedDirs []string
	// MaxFileSize 最大文件大小（字节）
	MaxFileSize int64
	// ReadOnly 是否只读模式
	ReadOnly bool
}

// NewFileModule 创建文件模块
func NewFileModule(config *FileModuleConfig) *FileModule {
	if config == nil {
		config = &FileModuleConfig{}
	}

	// 设置默认值
	maxFileSize := config.MaxFileSize
	if maxFileSize <= 0 {
		maxFileSize = 10 * 1024 * 1024 // 10MB
	}

	return &FileModule{
		allowedDirs: config.AllowedDirs,
		maxFileSize: maxFileSize,
		readOnly:    config.ReadOnly,
	}
}

// Name 返回模块名称
func (m *FileModule) Name() string {
	return "file"
}

// Register 注册模块到 VM
func (m *FileModule) Register(vm *goja.Runtime) error {
	module := vm.NewObject()

	// 设置模块方法
	_ = module.Set("read", m.makeReadFile(vm))
	_ = module.Set("write", m.makeWriteFile(vm))
	_ = module.Set("append", m.makeAppendFile(vm))
	_ = module.Set("exists", m.makeFileExists(vm))
	_ = module.Set("delete", m.makeDeleteFile(vm))
	_ = module.Set("rename", m.makeRenameFile(vm))
	_ = module.Set("copy", m.makeCopyFile(vm))
	_ = module.Set("mkdir", m.makeMkdir(vm))
	_ = module.Set("rmdir", m.makeRmdir(vm))
	_ = module.Set("list", m.makeListDir(vm))
	_ = module.Set("stat", m.makeFileStat(vm))
	_ = module.Set("readJSON", m.makeReadJSON(vm))
	_ = module.Set("writeJSON", m.makeWriteJSON(vm))

	// 注册到全局
	_ = vm.Set("file", module)

	return nil
}

// isPathAllowed 检查路径是否允许访问
func (m *FileModule) isPathAllowed(path string) error {
	// 获取绝对路径（统一使用正斜杠进行比较）
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %v", err)
	}
	// 统一路径分隔符为正斜杠，便于比较
	absPath = filepath.ToSlash(absPath)

	// 如果没有设置允许目录，则允许所有
	if len(m.allowedDirs) == 0 {
		return nil
	}

	// 检查是否在允许的目录内
	for _, dir := range m.allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		// 统一路径分隔符为正斜杠，便于比较
		absDir = filepath.ToSlash(absDir)
		if strings.HasPrefix(absPath, absDir) {
			return nil
		}
	}

	return fmt.Errorf("path not allowed: %s", path)
}

// makeReadFile 创建读取文件函数
func (m *FileModule) makeReadFile(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(path string) map[string]interface{} {
		// 检查路径权限
		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		// 读取文件
		content, err := os.ReadFile(path)
		if err != nil {
			return m.errorResult(err.Error())
		}

		// 检查文件大小
		if int64(len(content)) > m.maxFileSize {
			return m.errorResult("file too large")
		}

		return m.successResult(string(content))
	}
}

// makeWriteFile 创建写入文件函数
func (m *FileModule) makeWriteFile(vm *goja.Runtime) func(string, string) map[string]interface{} {
	return func(path string, content string) map[string]interface{} {
		// 检查只读模式
		if m.readOnly {
			return m.errorResult("file module is in read-only mode")
		}

		// 检查路径权限
		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		// 确保目录存在
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return m.errorResult(err.Error())
		}

		// 写入文件
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return m.errorResult(err.Error())
		}

		return m.successResult(true)
	}
}

// makeAppendFile 创建追加文件函数
func (m *FileModule) makeAppendFile(vm *goja.Runtime) func(string, string) map[string]interface{} {
	return func(path string, content string) map[string]interface{} {
		if m.readOnly {
			return m.errorResult("file module is in read-only mode")
		}

		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return m.errorResult(err.Error())
		}
		defer file.Close()

		if _, err := file.WriteString(content); err != nil {
			return m.errorResult(err.Error())
		}

		return m.successResult(true)
	}
}

// makeFileExists 创建检查文件存在函数
func (m *FileModule) makeFileExists(vm *goja.Runtime) func(string) bool {
	return func(path string) bool {
		if err := m.isPathAllowed(path); err != nil {
			return false
		}

		_, err := os.Stat(path)
		return err == nil
	}
}

// makeDeleteFile 创建删除文件函数
func (m *FileModule) makeDeleteFile(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(path string) map[string]interface{} {
		if m.readOnly {
			return m.errorResult("file module is in read-only mode")
		}

		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		err := os.Remove(path)
		if err != nil {
			return m.errorResult(err.Error())
		}

		return m.successResult(true)
	}
}

// makeRenameFile 创建重命名文件函数
func (m *FileModule) makeRenameFile(vm *goja.Runtime) func(string, string) map[string]interface{} {
	return func(oldPath string, newPath string) map[string]interface{} {
		if m.readOnly {
			return m.errorResult("file module is in read-only mode")
		}

		if err := m.isPathAllowed(oldPath); err != nil {
			return m.errorResult(err.Error())
		}
		if err := m.isPathAllowed(newPath); err != nil {
			return m.errorResult(err.Error())
		}

		err := os.Rename(oldPath, newPath)
		if err != nil {
			return m.errorResult(err.Error())
		}

		return m.successResult(true)
	}
}

// makeCopyFile 创建复制文件函数
func (m *FileModule) makeCopyFile(vm *goja.Runtime) func(string, string) map[string]interface{} {
	return func(src string, dst string) map[string]interface{} {
		if m.readOnly {
			return m.errorResult("file module is in read-only mode")
		}

		if err := m.isPathAllowed(src); err != nil {
			return m.errorResult(err.Error())
		}
		if err := m.isPathAllowed(dst); err != nil {
			return m.errorResult(err.Error())
		}

		// 读取源文件
		content, err := os.ReadFile(src)
		if err != nil {
			return m.errorResult(err.Error())
		}

		// 确保目标目录存在
		dir := filepath.Dir(dst)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return m.errorResult(err.Error())
		}

		// 写入目标文件
		err = os.WriteFile(dst, content, 0644)
		if err != nil {
			return m.errorResult(err.Error())
		}

		return m.successResult(true)
	}
}

// makeMkdir 创建目录函数
func (m *FileModule) makeMkdir(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(path string) map[string]interface{} {
		if m.readOnly {
			return m.errorResult("file module is in read-only mode")
		}

		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		err := os.MkdirAll(path, 0755)
		if err != nil {
			return m.errorResult(err.Error())
		}

		return m.successResult(true)
	}
}

// makeRmdir 创建删除目录函数
func (m *FileModule) makeRmdir(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(path string) map[string]interface{} {
		if m.readOnly {
			return m.errorResult("file module is in read-only mode")
		}

		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		err := os.RemoveAll(path)
		if err != nil {
			return m.errorResult(err.Error())
		}

		return m.successResult(true)
	}
}

// makeListDir 创建列出目录函数
func (m *FileModule) makeListDir(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(path string) map[string]interface{} {
		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return m.errorResult(err.Error())
		}

		// 构建结果数组
		result := make([]map[string]interface{}, 0, len(entries))
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			result = append(result, map[string]interface{}{
				"name":  entry.Name(),
				"isDir": entry.IsDir(),
				"size":  info.Size(),
				"mode":  info.Mode().String(),
			})
		}

		return m.successResult(result)
	}
}

// makeFileStat 创建获取文件信息函数
func (m *FileModule) makeFileStat(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(path string) map[string]interface{} {
		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		info, err := os.Stat(path)
		if err != nil {
			return m.errorResult(err.Error())
		}

		result := map[string]interface{}{
			"name":    info.Name(),
			"size":    info.Size(),
			"isDir":   info.IsDir(),
			"mode":    info.Mode().String(),
			"modTime": info.ModTime().Unix(),
		}

		return m.successResult(result)
	}
}

// makeReadJSON 创建读取 JSON 文件函数
func (m *FileModule) makeReadJSON(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(path string) map[string]interface{} {
		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return m.errorResult(err.Error())
		}

		var result interface{}
		if err := json.Unmarshal(content, &result); err != nil {
			return m.errorResult(fmt.Sprintf("invalid JSON: %v", err))
		}

		return m.successResult(result)
	}
}

// makeWriteJSON 创建写入 JSON 文件函数
func (m *FileModule) makeWriteJSON(vm *goja.Runtime) func(string, interface{}) map[string]interface{} {
	return func(path string, data interface{}) map[string]interface{} {
		if m.readOnly {
			return m.errorResult("file module is in read-only mode")
		}

		if err := m.isPathAllowed(path); err != nil {
			return m.errorResult(err.Error())
		}

		content, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return m.errorResult(err.Error())
		}

		// 确保目录存在
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return m.errorResult(err.Error())
		}

		err = os.WriteFile(path, content, 0644)
		if err != nil {
			return m.errorResult(err.Error())
		}

		return m.successResult(true)
	}
}

// successResult 返回成功结果
func (m *FileModule) successResult(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
	}
}

// errorResult 返回错误结果
func (m *FileModule) errorResult(errMsg string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   errMsg,
	}
}
