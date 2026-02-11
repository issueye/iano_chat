package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const (
	maxFileSize     = 10 * 1024 * 1024
	allowedBasePath = ""
)

type FileReadTool struct {
	basePath string
}

func NewFileReadTool(basePath string) *FileReadTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &FileReadTool{basePath: basePath}
}

func (t *FileReadTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_read",
		Desc: "读取文件内容，支持文本文件",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "文件路径（相对或绝对路径）",
				Required: true,
			},
			"offset": {
				Type:     schema.Number,
				Desc:     "起始行号（从0开始，可选）",
				Required: false,
			},
			"limit": {
				Type:     schema.Number,
				Desc:     "读取行数限制（可选，默认1000行）",
				Required: false,
			},
		}),
	}, nil
}

func (t *FileReadTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Path   string `json:"path"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Limit == 0 {
		args.Limit = 1000
	}

	absPath, err := t.resolvePath(args.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("文件不存在: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("路径是目录，不是文件")
	}

	if info.Size() > maxFileSize {
		return "", fmt.Errorf("文件大小超过限制 (%d MB)", maxFileSize/1024/1024)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	start := args.Offset
	if start < 0 {
		start = 0
	}
	if start >= totalLines {
		return fmt.Sprintf("文件共 %d 行，起始行 %d 超出范围", totalLines, start), nil
	}

	end := start + args.Limit
	if end > totalLines {
		end = totalLines
	}

	result := strings.Join(lines[start:end], "\n")

	return fmt.Sprintf("文件: %s\n行数: %d-%d / %d\n\n%s",
		args.Path, start+1, end, totalLines, result), nil
}

func (t *FileReadTool) resolvePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(t.basePath, path)
	}

	absPath = filepath.Clean(absPath)

	if t.basePath != "" {
		rel, err := filepath.Rel(t.basePath, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("路径超出允许范围")
		}
	}

	return absPath, nil
}

type FileWriteTool struct {
	basePath string
}

func NewFileWriteTool(basePath string) *FileWriteTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &FileWriteTool{basePath: basePath}
}

func (t *FileWriteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_write",
		Desc: "写入文件内容，如果文件不存在则创建",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "文件路径",
				Required: true,
			},
			"content": {
				Type:     schema.String,
				Desc:     "要写入的内容",
				Required: true,
			},
			"mode": {
				Type:     schema.String,
				Desc:     "写入模式: write(覆盖) 或 append(追加)，默认 write",
				Required: false,
			},
		}),
	}, nil
}

func (t *FileWriteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Path    string `json:"path"`
		Content string `json:"content"`
		Mode    string `json:"mode"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Path == "" {
		return "", fmt.Errorf("文件路径不能为空")
	}

	absPath, err := t.resolvePath(args.Path)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	var flag int = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if args.Mode == "append" {
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}

	file, err := os.OpenFile(absPath, flag, 0644)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(args.Content)
	if err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return fmt.Sprintf("成功写入文件: %s (%d 字节)", args.Path, len(args.Content)), nil
}

func (t *FileWriteTool) resolvePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(t.basePath, path)
	}
	absPath = filepath.Clean(absPath)

	if t.basePath != "" {
		rel, err := filepath.Rel(t.basePath, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("路径超出允许范围")
		}
	}

	return absPath, nil
}

type FileListTool struct {
	basePath string
}

func NewFileListTool(basePath string) *FileListTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &FileListTool{basePath: basePath}
}

func (t *FileListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_list",
		Desc: "列出目录内容",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "目录路径（默认为当前目录）",
				Required: false,
			},
			"recursive": {
				Type:     schema.Boolean,
				Desc:     "是否递归列出子目录",
				Required: false,
			},
			"pattern": {
				Type:     schema.String,
				Desc:     "文件名匹配模式（如 *.go）",
				Required: false,
			},
		}),
	}, nil
}

func (t *FileListTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Path      string `json:"path"`
		Recursive bool   `json:"recursive"`
		Pattern   string `json:"pattern"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Path == "" {
		args.Path = "."
	}

	absPath, err := t.resolvePath(args.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("路径不存在: %w", err)
	}

	if !info.IsDir() {
		return fmt.Sprintf("文件: %s (%d 字节, %s)",
			args.Path, info.Size(), info.Mode().String()), nil
	}

	var entries []fs.DirEntry
	if args.Recursive {
		return t.listRecursive(absPath, args.Pattern)
	}

	entries, err = os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("读取目录失败: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("目录: %s\n\n", args.Path))

	dirs := []string{}
	files := []string{}

	for _, entry := range entries {
		name := entry.Name()
		if args.Pattern != "" {
			matched, _ := filepath.Match(args.Pattern, name)
			if !matched {
				continue
			}
		}

		if entry.IsDir() {
			dirs = append(dirs, name+"/")
		} else {
			info, _ := entry.Info()
			files = append(files, fmt.Sprintf("%s (%s)", name, formatSize(info.Size())))
		}
	}

	if len(dirs) > 0 {
		result.WriteString("📁 目录:\n")
		for _, d := range dirs {
			result.WriteString(fmt.Sprintf("  %s\n", d))
		}
	}

	if len(files) > 0 {
		result.WriteString("\n📄 文件:\n")
		for _, f := range files {
			result.WriteString(fmt.Sprintf("  %s\n", f))
		}
	}

	result.WriteString(fmt.Sprintf("\n共 %d 个目录, %d 个文件", len(dirs), len(files)))

	return result.String(), nil
}

func (t *FileListTool) listRecursive(rootPath, pattern string) (string, error) {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("目录: %s (递归)\n\n", rootPath))

	count := 0
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(rootPath, path)
		if relPath == "." {
			return nil
		}

		if pattern != "" && !d.IsDir() {
			matched, _ := filepath.Match(pattern, d.Name())
			if !matched {
				return nil
			}
		}

		prefix := ""
		if d.IsDir() {
			prefix = "📁 "
		} else {
			prefix = "📄 "
		}

		result.WriteString(fmt.Sprintf("%s%s\n", prefix, relPath))
		count++

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("遍历目录失败: %w", err)
	}

	result.WriteString(fmt.Sprintf("\n共 %d 项", count))
	return result.String(), nil
}

func (t *FileListTool) resolvePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(t.basePath, path)
	}
	absPath = filepath.Clean(absPath)

	if t.basePath != "" {
		rel, err := filepath.Rel(t.basePath, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("路径超出允许范围")
		}
	}

	return absPath, nil
}

type FileDeleteTool struct {
	basePath string
}

func NewFileDeleteTool(basePath string) *FileDeleteTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &FileDeleteTool{basePath: basePath}
}

func (t *FileDeleteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_delete",
		Desc: "删除文件或目录",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "要删除的文件或目录路径",
				Required: true,
			},
			"recursive": {
				Type:     schema.Boolean,
				Desc:     "是否递归删除目录（谨慎使用）",
				Required: false,
			},
		}),
	}, nil
}

func (t *FileDeleteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Path      string `json:"path"`
		Recursive bool   `json:"recursive"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Path == "" {
		return "", fmt.Errorf("路径不能为空")
	}

	absPath, err := t.resolvePath(args.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("路径不存在: %w", err)
	}

	if info.IsDir() {
		if args.Recursive {
			if err := os.RemoveAll(absPath); err != nil {
				return "", fmt.Errorf("删除目录失败: %w", err)
			}
			return fmt.Sprintf("已递归删除目录: %s", args.Path), nil
		}

		entries, err := os.ReadDir(absPath)
		if err != nil {
			return "", fmt.Errorf("读取目录失败: %w", err)
		}
		if len(entries) > 0 {
			return "", fmt.Errorf("目录不为空，需要设置 recursive=true 才能删除")
		}

		if err := os.Remove(absPath); err != nil {
			return "", fmt.Errorf("删除目录失败: %w", err)
		}
		return fmt.Sprintf("已删除空目录: %s", args.Path), nil
	}

	if err := os.Remove(absPath); err != nil {
		return "", fmt.Errorf("删除文件失败: %w", err)
	}

	return fmt.Sprintf("已删除文件: %s", args.Path), nil
}

func (t *FileDeleteTool) resolvePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(t.basePath, path)
	}
	absPath = filepath.Clean(absPath)

	if t.basePath != "" {
		rel, err := filepath.Rel(t.basePath, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("路径超出允许范围")
		}
	}

	return absPath, nil
}

type FileInfoTool struct {
	basePath string
}

func NewFileInfoTool(basePath string) *FileInfoTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &FileInfoTool{basePath: basePath}
}

func (t *FileInfoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file_info",
		Desc: "获取文件或目录的详细信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path": {
				Type:     schema.String,
				Desc:     "文件或目录路径",
				Required: true,
			},
		}),
	}, nil
}

func (t *FileInfoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	absPath, err := t.resolvePath(args.Path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("路径不存在: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("路径: %s\n", args.Path))
	result.WriteString(fmt.Sprintf("类型: %s\n", map[bool]string{true: "目录", false: "文件"}[info.IsDir()]))
	result.WriteString(fmt.Sprintf("大小: %s\n", formatSize(info.Size())))
	result.WriteString(fmt.Sprintf("权限: %s\n", info.Mode().String()))
	result.WriteString(fmt.Sprintf("修改时间: %s\n", info.ModTime().Format(time.RFC3339)))

	if !info.IsDir() {
		ext := filepath.Ext(info.Name())
		if ext != "" {
			result.WriteString(fmt.Sprintf("扩展名: %s\n", ext))
		}
	}

	return result.String(), nil
}

func (t *FileInfoTool) resolvePath(path string) (string, error) {
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(t.basePath, path)
	}
	absPath = filepath.Clean(absPath)

	if t.basePath != "" {
		rel, err := filepath.Rel(t.basePath, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("路径超出允许范围")
		}
	}

	return absPath, nil
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
