package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type GrepSearchTool struct {
	basePath string
}

func NewGrepSearchTool(basePath string) *GrepSearchTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &GrepSearchTool{basePath: basePath}
}

func (t *GrepSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "grep_search",
		Desc: "在文件中搜索文本模式，支持正则表达式",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"pattern": {
				Type:     schema.String,
				Desc:     "搜索模式（支持正则表达式）",
				Required: true,
			},
			"path": {
				Type:     schema.String,
				Desc:     "搜索路径（默认为当前目录）",
				Required: false,
			},
			"recursive": {
				Type:     schema.Boolean,
				Desc:     "是否递归搜索子目录",
				Required: false,
			},
			"file_pattern": {
				Type:     schema.String,
				Desc:     "文件匹配模式（如 *.go, *.txt）",
				Required: false,
			},
			"ignore_case": {
				Type:     schema.Boolean,
				Desc:     "忽略大小写",
				Required: false,
			},
			"line_numbers": {
				Type:     schema.Boolean,
				Desc:     "显示行号",
				Required: false,
			},
			"max_count": {
				Type:     schema.Number,
				Desc:     "最大匹配行数",
				Required: false,
			},
		}),
	}, nil
}

func (t *GrepSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Pattern     string `json:"pattern"`
		Path        string `json:"path"`
		Recursive   bool   `json:"recursive"`
		FilePattern string `json:"file_pattern"`
		IgnoreCase  bool   `json:"ignore_case"`
		LineNumbers bool   `json:"line_numbers"`
		MaxCount    int    `json:"max_count"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Pattern == "" {
		return "", fmt.Errorf("搜索模式不能为空")
	}

	searchPath := args.Path
	if searchPath == "" {
		searchPath = t.basePath
	}

	absPath, err := t.resolvePath(searchPath)
	if err != nil {
		return "", err
	}

	var re *regexp.Regexp
	pattern := args.Pattern
	if args.IgnoreCase {
		pattern = "(?i)" + pattern
	}
	re, err = regexp.CompilePOSIX(pattern)
	if err != nil {
		return "", fmt.Errorf("正则表达式无效: %w", err)
	}

	var results []string
	var totalMatches int

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() && !args.Recursive && path != absPath {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		if args.FilePattern != "" {
			matched, _ := filepath.Match(args.FilePattern, d.Name())
			if !matched {
				return nil
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if re.MatchString(line) {
				totalMatches++
				if args.MaxCount > 0 && totalMatches > args.MaxCount {
					return fmt.Errorf("达到最大匹配数")
				}
				if args.LineNumbers {
					relPath, _ := filepath.Rel(absPath, path)
					results = append(results, fmt.Sprintf("%s:%d:%s", relPath, i+1, line))
				} else {
					results = append(results, fmt.Sprintf("%s: %s", filepath.Base(path), line))
				}
			}
		}

		return nil
	}

	if args.Recursive {
		filepath.WalkDir(absPath, walkFn)
	} else {
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return "", fmt.Errorf("读取目录失败: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			walkFn(filepath.Join(absPath, entry.Name()), entry, nil)
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("未找到匹配结果 (路径: %s, 模式: %s)", args.Path, args.Pattern), nil
	}

	output := fmt.Sprintf("找到 %d 个匹配:\n\n%s", totalMatches, strings.Join(results, "\n"))
	return output, nil
}

func (t *GrepSearchTool) resolvePath(path string) (string, error) {
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

type GrepReplaceTool struct {
	basePath string
}

func NewGrepReplaceTool(basePath string) *GrepReplaceTool {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &GrepReplaceTool{basePath: basePath}
}

func (t *GrepReplaceTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "grep_replace",
		Desc: "在文件中搜索并替换文本",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"pattern": {
				Type:     schema.String,
				Desc:     "要搜索的模式",
				Required: true,
			},
			"replacement": {
				Type:     schema.String,
				Desc:     "替换为的文本",
				Required: true,
			},
			"path": {
				Type:     schema.String,
				Desc:     "文件路径",
				Required: true,
			},
			"ignore_case": {
				Type:     schema.Boolean,
				Desc:     "忽略大小写",
				Required: false,
			},
		}),
	}, nil
}

func (t *GrepReplaceTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Pattern     string `json:"pattern"`
		Replacement string `json:"replacement"`
		Path        string `json:"path"`
		IgnoreCase  bool   `json:"ignore_case"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	absPath, err := t.resolvePath(args.Path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	var newContent string
	pattern := regexp.QuoteMeta(args.Pattern)
	if args.IgnoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("正则表达式无效: %w", err)
	}

	newContent = re.ReplaceAllString(string(content), args.Replacement)

	if newContent == string(content) {
		return fmt.Sprintf("文件中未找到匹配的模式: %s", args.Pattern), nil
	}

	if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	replacedCount := strings.Count(string(content), args.Replacement) - strings.Count(newContent, args.Replacement)
	if replacedCount == 0 {
		replacedCount = strings.Count(string(content), args.Pattern) - strings.Count(newContent, args.Pattern)
	}

	return fmt.Sprintf("已替换: %s", args.Path), nil
}

func (t *GrepReplaceTool) resolvePath(path string) (string, error) {
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
