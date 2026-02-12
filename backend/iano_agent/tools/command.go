package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const (
	defaultTimeout     = 30 * time.Second
	maxOutputSize      = 100 * 1024
	allowedCommandsKey = "ALLOWED_COMMANDS"
)

type ShellType string

const (
	ShellCmd        ShellType = "cmd"
	ShellPowerShell ShellType = "powershell"
	ShellBash       ShellType = "bash"
)

var (
	defaultAllowedCommands = []string{
		"ls", "dir", "cat", "head", "tail", "wc", "grep", "find",
		"echo", "pwd", "whoami", "date", "uname",
		"git", "npm", "yarn", "pip", "python", "python3", "node",
		"go", "cargo", "rustc",
	}
)

type CommandExecuteTool struct {
	timeout         time.Duration
	allowedCommands map[string]bool
	workingDir      string
	shell           ShellType
}

func NewCommandExecuteTool() *CommandExecuteTool {
	t := &CommandExecuteTool{
		timeout:         defaultTimeout,
		allowedCommands: make(map[string]bool),
	}

	for _, cmd := range defaultAllowedCommands {
		t.allowedCommands[cmd] = true
	}

	if extra := os.Getenv(allowedCommandsKey); extra != "" {
		for _, cmd := range strings.Split(extra, ",") {
			t.allowedCommands[strings.TrimSpace(cmd)] = true
		}
	}

	if runtime.GOOS == "windows" {
		t.shell = ShellPowerShell
	} else {
		t.shell = ShellBash
	}

	return t
}

func (t *CommandExecuteTool) WithTimeout(timeout time.Duration) *CommandExecuteTool {
	t.timeout = timeout
	return t
}

func (t *CommandExecuteTool) WithWorkingDir(dir string) *CommandExecuteTool {
	t.workingDir = dir
	return t
}

func (t *CommandExecuteTool) WithAllowedCommands(commands []string) *CommandExecuteTool {
	for _, cmd := range commands {
		t.allowedCommands[cmd] = true
	}
	return t
}

// WithShell 设置执行命令的 Shell 类型
func (t *CommandExecuteTool) WithShell(shell ShellType) *CommandExecuteTool {
	t.shell = shell
	return t
}

func (t *CommandExecuteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "command_execute",
		Desc: "执行系统命令（仅允许安全的命令）",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"command": {
				Type:     schema.String,
				Desc:     "要执行的命令",
				Required: true,
			},
			"args": {
				Type:     schema.String,
				Desc:     "命令参数（空格分隔）",
				Required: false,
			},
			"timeout": {
				Type:     schema.Number,
				Desc:     "超时时间（秒），默认30秒",
				Required: false,
			},
		}),
	}, nil
}

func (t *CommandExecuteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Command string `json:"command"`
		Args    string `json:"args"`
		Timeout int    `json:"timeout"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Command == "" {
		return "", fmt.Errorf("命令不能为空")
	}

	if !t.isCommandAllowed(args.Command) {
		return "", fmt.Errorf("命令 '%s' 不在允许列表中", args.Command)
	}

	timeout := t.timeout
	if args.Timeout > 0 {
		timeout = time.Duration(args.Timeout) * time.Second
		if timeout > 5*time.Minute {
			timeout = 5 * time.Minute
		}
	}

	cmdArgs := []string{}
	if args.Args != "" {
		cmdArgs = strings.Fields(args.Args)
	}

	return t.executeCommand(args.Command, cmdArgs, timeout)
}

func (t *CommandExecuteTool) isCommandAllowed(command string) bool {
	if runtime.GOOS == "windows" {
		command = strings.ToLower(command)
		command = strings.TrimSuffix(command, ".exe")
	}

	_, allowed := t.allowedCommands[command]
	return allowed
}

func (t *CommandExecuteTool) executeCommand(name string, args []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	fullCommand := name + " " + strings.Join(args, " ")

	switch t.shell {
	case ShellPowerShell:
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", fullCommand)
	case ShellCmd:
		cmd = exec.CommandContext(ctx, "cmd", "/c", fullCommand)
	default:
		cmd = exec.CommandContext(ctx, "bash", "-c", fullCommand)
	}

	if t.workingDir != "" {
		cmd.Dir = t.workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	output := stdout.String()
	if len(output) > maxOutputSize {
		output = output[:maxOutputSize] + "\n... (输出被截断)"
	}

	errOutput := stderr.String()
	if len(errOutput) > maxOutputSize {
		errOutput = errOutput[:maxOutputSize] + "\n... (错误输出被截断)"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("命令: %s %s\n", name, strings.Join(args, " ")))
	result.WriteString(fmt.Sprintf("执行时间: %v\n", duration))
	result.WriteString(fmt.Sprintf("退出码: %d\n", cmd.ProcessState.ExitCode()))

	if output != "" {
		result.WriteString(fmt.Sprintf("\n--- 标准输出 ---\n%s", output))
	}

	if errOutput != "" {
		result.WriteString(fmt.Sprintf("\n--- 标准错误 ---\n%s", errOutput))
	}

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("命令执行超时 (%v)", timeout)
	}

	if err != nil {
		return result.String(), fmt.Errorf("命令执行失败: %w", err)
	}

	return result.String(), nil
}

type ShellExecuteTool struct {
	timeout    time.Duration
	workingDir string
	shell      ShellType
}

func NewShellExecuteTool() *ShellExecuteTool {
	t := &ShellExecuteTool{
		timeout: defaultTimeout,
	}
	if runtime.GOOS == "windows" {
		t.shell = ShellPowerShell
	} else {
		t.shell = ShellBash
	}
	return t
}

func (t *ShellExecuteTool) WithTimeout(timeout time.Duration) *ShellExecuteTool {
	t.timeout = timeout
	return t
}

func (t *ShellExecuteTool) WithWorkingDir(dir string) *ShellExecuteTool {
	t.workingDir = dir
	return t
}

// WithShell 设置执行命令的 Shell 类型
func (t *ShellExecuteTool) WithShell(shell ShellType) *ShellExecuteTool {
	t.shell = shell
	return t
}

func (t *ShellExecuteTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "shell_execute",
		Desc: "执行 Shell 命令（受限模式，仅支持简单命令）",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"command": {
				Type:     schema.String,
				Desc:     "Shell 命令",
				Required: true,
			},
			"timeout": {
				Type:     schema.Number,
				Desc:     "超时时间（秒），默认30秒",
				Required: false,
			},
		}),
	}, nil
}

func (t *ShellExecuteTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Command string `json:"command"`
		Timeout int    `json:"timeout"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Command == "" {
		return "", fmt.Errorf("命令不能为空")
	}

	if hasDangerousContent(args.Command) {
		return "", fmt.Errorf("命令包含危险内容")
	}

	timeout := t.timeout
	if args.Timeout > 0 {
		timeout = time.Duration(args.Timeout) * time.Second
		if timeout > 5*time.Minute {
			timeout = 5 * time.Minute
		}
	}

	return t.executeShell(args.Command, timeout)
}

func hasDangerousContent(s string) bool {
	dangerous := []string{
		"rm -rf", "mkfs", "dd if=", "> /dev/", ":(){ :|:& };:",
		"chmod 777", "chown -R", "wget", "curl -o",
		"eval", "exec", "/etc/passwd", "/etc/shadow",
		"nc -l", "ncat", "telnet", "ftp",
	}
	lowerCmd := strings.ToLower(s)
	for _, d := range dangerous {
		if strings.Contains(lowerCmd, strings.ToLower(d)) {
			return true
		}
	}
	return false
}

func (t *ShellExecuteTool) executeShell(command string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	switch t.shell {
	case ShellPowerShell:
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", command)
	case ShellCmd:
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	default:
		cmd = exec.CommandContext(ctx, "bash", "-c", command)
	}

	if t.workingDir != "" {
		cmd.Dir = t.workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	output := stdout.String()
	if len(output) > maxOutputSize {
		output = output[:maxOutputSize] + "\n... (输出被截断)"
	}

	errOutput := stderr.String()
	if len(errOutput) > maxOutputSize {
		errOutput = errOutput[:maxOutputSize] + "\n... (错误输出被截断)"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Shell: %s\n", command))
	result.WriteString(fmt.Sprintf("执行时间: %v\n", duration))

	if output != "" {
		result.WriteString(fmt.Sprintf("\n--- 输出 ---\n%s", output))
	}

	if errOutput != "" {
		result.WriteString(fmt.Sprintf("\n--- 错误 ---\n%s", errOutput))
	}

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("命令执行超时 (%v)", timeout)
	}

	if err != nil {
		return result.String(), fmt.Errorf("命令执行失败: %w", err)
	}

	return result.String(), nil
}

type ProcessListTool struct{}

func NewProcessListTool() *ProcessListTool {
	return &ProcessListTool{}
}

func (t *ProcessListTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "process_list",
		Desc: "列出当前运行的进程",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"filter": {
				Type:     schema.String,
				Desc:     "进程名过滤（可选）",
				Required: false,
			},
		}),
	}, nil
}

func (t *ProcessListTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Filter string `json:"filter"`
	}

	json.Unmarshal([]byte(argumentsInJSON), &args)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if args.Filter != "" {
			cmd = exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s*", args.Filter))
		} else {
			cmd = exec.Command("tasklist")
		}
	} else {
		if args.Filter != "" {
			cmd = exec.Command("sh", "-c", fmt.Sprintf("ps aux | grep -i %s | grep -v grep", args.Filter))
		} else {
			cmd = exec.Command("ps", "aux")
		}
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取进程列表失败: %w", err)
	}

	return string(output), nil
}
