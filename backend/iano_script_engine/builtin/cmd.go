// Package builtin - 命令执行模块
// 提供命令行执行功能

package builtin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// CmdModule 命令执行模块
type CmdModule struct {
	timeout       time.Duration
	allowedCmds   []string
	blockedCmds   []string
	allowedDirs   []string
	env           map[string]string
	maxOutputSize int64
	enableShell   bool
}

// CmdModuleConfig 命令模块配置
type CmdModuleConfig struct {
	// Timeout 默认命令超时时间
	Timeout time.Duration
	// AllowedCmds 允许执行的命令列表（空表示允许所有）
	AllowedCmds []string
	// BlockedCmds 禁止执行的命令列表
	BlockedCmds []string
	// AllowedDirs 允许执行命令的目录
	AllowedDirs []string
	// Env 额外的环境变量
	Env map[string]string
	// MaxOutputSize 最大输出大小（字节）
	MaxOutputSize int64
	// EnableShell 是否允许 shell 模式
	EnableShell bool
}

// NewCmdModule 创建命令模块
func NewCmdModule(config *CmdModuleConfig) *CmdModule {
	if config == nil {
		config = &CmdModuleConfig{}
	}

	// 设置默认值
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	maxOutputSize := config.MaxOutputSize
	if maxOutputSize <= 0 {
		maxOutputSize = 1024 * 1024 // 1MB
	}

	return &CmdModule{
		timeout:       timeout,
		allowedCmds:   config.AllowedCmds,
		blockedCmds:   config.BlockedCmds,
		allowedDirs:   config.AllowedDirs,
		env:           config.Env,
		maxOutputSize: maxOutputSize,
		enableShell:   config.EnableShell,
	}
}

// Name 返回模块名称
func (m *CmdModule) Name() string {
	return "cmd"
}

// Register 注册模块到 VM
func (m *CmdModule) Register(vm *goja.Runtime) error {
	module := vm.NewObject()

	// 设置模块方法
	_ = module.Set("exec", m.makeExec(vm))
	_ = module.Set("execSync", m.makeExecSync(vm))
	_ = module.Set("execWithTimeout", m.makeExecWithTimeout(vm))
	_ = module.Set("shell", m.makeShell(vm))
	_ = module.Set("which", m.makeWhich(vm))
	_ = module.Set("env", m.makeEnv(vm))

	// 注册到全局
	_ = vm.Set("cmd", module)

	return nil
}

// isCommandAllowed 检查命令是否允许执行
func (m *CmdModule) isCommandAllowed(cmd string) error {
	// 提取命令名称（不含路径）
	cmdName := cmd
	if idx := strings.LastIndex(cmd, "/"); idx != -1 {
		cmdName = cmd[idx+1:]
	} else if idx := strings.LastIndex(cmd, "\\"); idx != -1 {
		cmdName = cmd[idx+1:]
	}

	// 检查禁止列表
	for _, blocked := range m.blockedCmds {
		if strings.EqualFold(cmdName, blocked) || strings.EqualFold(cmd, blocked) {
			return fmt.Errorf("command is blocked: %s", cmd)
		}
	}

	// 检查允许列表
	if len(m.allowedCmds) > 0 {
		allowed := false
		for _, a := range m.allowedCmds {
			if strings.EqualFold(cmdName, a) || strings.EqualFold(cmd, a) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("command not allowed: %s", cmd)
		}
	}

	return nil
}

// makeExec 创建执行命令函数
func (m *CmdModule) makeExec(vm *goja.Runtime) func(interface{}, ...interface{}) map[string]interface{} {
	return func(cmdArg interface{}, args ...interface{}) map[string]interface{} {
		var cmdStr string
		var cmdArgs []string

		switch v := cmdArg.(type) {
		case string:
			cmdStr = v
		case []interface{}:
			if len(v) == 0 {
				return m.errorResult("empty command")
			}
			cmdStr = fmt.Sprintf("%v", v[0])
			for i := 1; i < len(v); i++ {
				cmdArgs = append(cmdArgs, fmt.Sprintf("%v", v[i]))
			}
		default:
			return m.errorResult("invalid command type")
		}

		// 处理额外参数
		for _, arg := range args {
			switch v := arg.(type) {
			case []interface{}:
				for _, a := range v {
					cmdArgs = append(cmdArgs, fmt.Sprintf("%v", a))
				}
			case string:
				cmdArgs = append(cmdArgs, v)
			default:
				cmdArgs = append(cmdArgs, fmt.Sprintf("%v", v))
			}
		}

		// 检查命令权限
		if err := m.isCommandAllowed(cmdStr); err != nil {
			return m.errorResult(err.Error())
		}

		// 执行命令
		return m.runCommand(cmdStr, cmdArgs, m.timeout, nil, "")
	}
}

// makeExecSync 创建同步执行命令函数
func (m *CmdModule) makeExecSync(vm *goja.Runtime) func(interface{}, ...interface{}) map[string]interface{} {
	return m.makeExec(vm)
}

// makeExecWithTimeout 创建带超时的命令执行函数
func (m *CmdModule) makeExecWithTimeout(vm *goja.Runtime) func(interface{}, int64, ...interface{}) map[string]interface{} {
	return func(cmdArg interface{}, timeoutMs int64, args ...interface{}) map[string]interface{} {
		var cmdStr string
		var cmdArgs []string

		switch v := cmdArg.(type) {
		case string:
			cmdStr = v
		case []interface{}:
			if len(v) == 0 {
				return m.errorResult("empty command")
			}
			cmdStr = fmt.Sprintf("%v", v[0])
			for i := 1; i < len(v); i++ {
				cmdArgs = append(cmdArgs, fmt.Sprintf("%v", v[i]))
			}
		}

		// 处理额外参数
		for _, arg := range args {
			switch v := arg.(type) {
			case []interface{}:
				for _, a := range v {
					cmdArgs = append(cmdArgs, fmt.Sprintf("%v", a))
				}
			case string:
				cmdArgs = append(cmdArgs, v)
			}
		}

		// 解析超时时间
		timeout := time.Duration(timeoutMs) * time.Millisecond
		if timeout <= 0 {
			timeout = m.timeout
		}

		// 检查命令权限
		if err := m.isCommandAllowed(cmdStr); err != nil {
			return m.errorResult(err.Error())
		}

		return m.runCommand(cmdStr, cmdArgs, timeout, nil, "")
	}
}

// makeShell 创建 shell 命令执行函数
func (m *CmdModule) makeShell(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(cmdStr string) map[string]interface{} {
		if !m.enableShell {
			return m.errorResult("shell mode is disabled")
		}

		// 根据操作系统选择 shell
		var shell string
		var flag string
		if runtime.GOOS == "windows" {
			shell = "cmd"
			flag = "/c"
		} else {
			shell = "/bin/sh"
			flag = "-c"
		}

		// 检查 shell 权限
		if err := m.isCommandAllowed(shell); err != nil {
			return m.errorResult(err.Error())
		}

		return m.runCommand(shell, []string{flag, cmdStr}, m.timeout, nil, "")
	}
}

// makeWhich 创建查找命令路径函数
func (m *CmdModule) makeWhich(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(cmdName string) map[string]interface{} {
		path, err := exec.LookPath(cmdName)
		if err != nil {
			return m.errorResult("command not found")
		}
		return m.successResult(path)
	}
}

// makeEnv 创建获取环境变量函数
func (m *CmdModule) makeEnv(vm *goja.Runtime) func(...string) map[string]interface{} {
	return func(keys ...string) map[string]interface{} {
		// 如果没有参数，返回所有环境变量
		if len(keys) == 0 {
			envMap := make(map[string]string)
			for _, env := range m.getEnvList() {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}
			return m.successResult(envMap)
		}

		// 获取指定环境变量
		if len(keys) == 1 {
			for _, env := range m.getEnvList() {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 && parts[0] == keys[0] {
					return m.successResult(parts[1])
				}
			}
			return m.successResult("")
		}

		// 获取多个环境变量
		envMap := make(map[string]string)
		for _, key := range keys {
			for _, env := range m.getEnvList() {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 && parts[0] == key {
					envMap[key] = parts[1]
					break
				}
			}
		}
		return m.successResult(envMap)
	}
}

// getEnvList 获取环境变量列表
func (m *CmdModule) getEnvList() []string {
	env := exec.Command("").Env
	if env == nil {
		env = []string{}
	}

	// 添加额外环境变量
	for k, v := range m.env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

// runCommand 执行命令
func (m *CmdModule) runCommand(cmdStr string, args []string, timeout time.Duration, env []string, dir string) map[string]interface{} {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 创建命令
	cmd := exec.CommandContext(ctx, cmdStr, args...)

	// 设置工作目录
	if dir != "" {
		cmd.Dir = dir
	} else if len(m.allowedDirs) > 0 {
		cmd.Dir = m.allowedDirs[0]
	}

	// 设置环境变量
	if env != nil {
		cmd.Env = env
	} else {
		cmd.Env = m.getEnvList()
	}

	// 捕获输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime).Milliseconds()

	// 检查是否超时
	if ctx.Err() == context.DeadlineExceeded {
		return map[string]interface{}{
			"success":  false,
			"error":    "command timeout",
			"stdout":   stdout.String(),
			"stderr":   stderr.String(),
			"duration": duration,
			"exitCode": -1,
		}
	}

	// 构建结果
	result := map[string]interface{}{
		"success":  err == nil,
		"stdout":   stdout.String(),
		"stderr":   stderr.String(),
		"duration": duration,
	}

	if err != nil {
		result["error"] = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exitCode"] = exitErr.ExitCode()
		} else {
			result["exitCode"] = -1
		}
	} else {
		result["exitCode"] = 0
	}

	return result
}

// successResult 返回成功结果
func (m *CmdModule) successResult(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
	}
}

// errorResult 返回错误结果
func (m *CmdModule) errorResult(errMsg string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   errMsg,
	}
}
