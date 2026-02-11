package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type EnvironmentGetTool struct{}

func NewEnvironmentGetTool() *EnvironmentGetTool {
	return &EnvironmentGetTool{}
}

func (t *EnvironmentGetTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "env_get",
		Desc: "获取环境变量",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"name": {
				Type:     schema.String,
				Desc:     "环境变量名称",
				Required: false,
			},
		}),
	}, nil
}

func (t *EnvironmentGetTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Name != "" {
		value := os.Getenv(args.Name)
		if value == "" {
			return fmt.Sprintf("环境变量 '%s' 未设置或为空", args.Name), nil
		}
		return fmt.Sprintf("%s=%s", args.Name, value), nil
	}

	var result strings.Builder
	result.WriteString("环境变量:\n")
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			if isSensitiveEnvVar(parts[0]) {
				result.WriteString(fmt.Sprintf("  %s=*******\n", parts[0]))
			} else {
				result.WriteString(fmt.Sprintf("  %s=%s\n", parts[0], parts[1]))
			}
		}
	}

	return result.String(), nil
}

func isSensitiveEnvVar(name string) bool {
	sensitive := []string{
		"PASSWORD", "PASSWD", "SECRET", "TOKEN", "KEY", "API_KEY",
		"AUTH", "CREDENTIAL", "PRIVATE", "AWS_ACCESS",
	}
	nameUpper := strings.ToUpper(name)
	for _, s := range sensitive {
		if strings.Contains(nameUpper, s) {
			return true
		}
	}
	return false
}

type EnvironmentSetTool struct{}

func NewEnvironmentSetTool() *EnvironmentSetTool {
	return &EnvironmentSetTool{}
}

func (t *EnvironmentSetTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "env_set",
		Desc: "设置环境变量（仅当前进程有效）",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"name": {
				Type:     schema.String,
				Desc:     "环境变量名称",
				Required: true,
			},
			"value": {
				Type:     schema.String,
				Desc:     "环境变量值",
				Required: true,
			},
		}),
	}, nil
}

func (t *EnvironmentSetTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Name == "" {
		return "", fmt.Errorf("环境变量名称不能为空")
	}

	if isSensitiveEnvVar(args.Name) {
		return "", fmt.Errorf("不能设置敏感环境变量: %s", args.Name)
	}

	os.Setenv(args.Name, args.Value)
	return fmt.Sprintf("已设置: %s=%s", args.Name, args.Value), nil
}

type SystemInfoTool struct{}

func NewSystemInfoTool() *SystemInfoTool {
	return &SystemInfoTool{}
}

func (t *SystemInfoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "system_info",
		Desc: "获取系统信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"info_type": {
				Type:     schema.String,
				Desc:     "信息类型: os, cpu, memory, all",
				Required: false,
			},
		}),
	}, nil
}

func (t *SystemInfoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		InfoType string `json:"info_type"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	infoType := strings.ToLower(args.InfoType)
	if infoType == "" {
		infoType = "all"
	}

	var result strings.Builder

	if infoType == "os" || infoType == "all" {
		result.WriteString(fmt.Sprintf("操作系统: %s\n", runtime.GOOS))
		result.WriteString(fmt.Sprintf("架构: %s\n", runtime.GOARCH))
		result.WriteString(fmt.Sprintf("编译器: %s\n", runtime.Compiler))
		result.WriteString(fmt.Sprintf("版本: %s\n", runtime.Version()))
	}

	if infoType == "cpu" || infoType == "all" {
		result.WriteString(fmt.Sprintf("CPU 核心数: %d\n", runtime.NumCPU()))
	}

	if infoType == "memory" || infoType == "all" {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		result.WriteString(fmt.Sprintf("堆内存分配: %.2f MB\n", float64(m.Alloc)/1024/1024))
		result.WriteString(fmt.Sprintf("总堆内存: %.2f MB\n", float64(m.TotalAlloc)/1024/1024))
		result.WriteString(fmt.Sprintf("系统内存: %.2f MB\n", float64(m.Sys)/1024/1024))
	}

	return result.String(), nil
}
