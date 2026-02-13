// Package builtin - 命令模块测试

package builtin

import (
	"runtime"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func TestCmdModule_Name(t *testing.T) {
	module := NewCmdModule(nil)
	assert.Equal(t, "cmd", module.Name())
}

func TestCmdModule_Register(t *testing.T) {
	vm := goja.New()
	module := NewCmdModule(nil)
	err := module.Register(vm)
	assert.NoError(t, err)

	// 检查 cmd 对象是否存在
	cmdObj := vm.Get("cmd")
	assert.NotNil(t, cmdObj)
	assert.False(t, goja.IsUndefined(cmdObj))
}

func TestCmdModule_Exec(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	// 根据操作系统选择命令
	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.exec("cmd", ["/c", "echo", "hello"])`
	} else {
		script = `cmd.exec("echo", ["hello"])`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result["stdout"].(string), "hello")
}

func TestCmdModule_ExecWithArray(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.exec(["cmd", "/c", "echo", "world"])`
	} else {
		script = `cmd.exec(["echo", "world"])`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result["stdout"].(string), "world")
}

func TestCmdModule_ExecWithTimeout(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.execWithTimeout("cmd", 5000, ["/c", "echo", "timeout test"])`
	} else {
		script = `cmd.execWithTimeout("echo", 5000, ["timeout test"])`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result["stdout"].(string), "timeout test")
}

func TestCmdModule_Which(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	// 查找命令
	var cmdName string
	if runtime.GOOS == "windows" {
		cmdName = "cmd"
	} else {
		cmdName = "sh"
	}

	script := `cmd.which("` + cmdName + `")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	assert.NotEmpty(t, result["data"])
}

func TestCmdModule_WhichNotFound(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	script := `cmd.which("nonexistent_command_12345")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "not found")
}

func TestCmdModule_Env(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	// 获取所有环境变量
	script := `cmd.env()`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	// envMap 可能是 map[string]interface{} 或 map[string]string
	envData := result["data"]
	assert.NotNil(t, envData)
}

func TestCmdModule_EnvSingle(t *testing.T) {
	module := NewCmdModule(&CmdModuleConfig{
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 获取指定环境变量
	script := `cmd.env("TEST_VAR")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "test_value", result["data"])
}

func TestCmdModule_BlockedCommand(t *testing.T) {
	module := NewCmdModule(&CmdModuleConfig{
		BlockedCmds: []string{"rm", "del", "format"},
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 尝试执行被禁止的命令
	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.exec("format", ["test"])`
	} else {
		script = `cmd.exec("rm", ["test.txt"])`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "blocked")
}

func TestCmdModule_AllowedCommand(t *testing.T) {
	// 在 Windows 上，命令可能是 cmd.exe，需要允许 cmd 和 cmd.exe
	// 在没有 AllowedCmds 限制时，所有命令都允许
	module := NewCmdModule(&CmdModuleConfig{
		AllowedCmds: nil, // nil 表示允许所有命令
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 执行允许的命令
	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.exec("cmd", ["/c", "echo", "allowed"])`
	} else {
		script = `cmd.exec("echo", ["allowed"])`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
}

func TestCmdModule_NotAllowedCommand(t *testing.T) {
	module := NewCmdModule(&CmdModuleConfig{
		AllowedCmds: []string{"echo"},
	})
	vm := goja.New()
	_ = module.Register(vm)

	// 尝试执行不在允许列表的命令
	script := `cmd.exec("ls", [])`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "not allowed")
}

func TestCmdModule_ShellDisabled(t *testing.T) {
	module := NewCmdModule(&CmdModuleConfig{
		EnableShell: false,
	})
	vm := goja.New()
	_ = module.Register(vm)

	script := `cmd.shell("echo test")`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "disabled")
}

func TestCmdModule_ShellEnabled(t *testing.T) {
	// 不设置 AllowedCmds，允许所有命令
	module := NewCmdModule(&CmdModuleConfig{
		EnableShell: true,
	})
	vm := goja.New()
	_ = module.Register(vm)

	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.shell("echo shell test")`
	} else {
		script = `cmd.shell("echo shell test")`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result["stdout"].(string), "shell test")
}

func TestCmdModule_ExitCode(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	// 执行一个会失败的命令
	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.exec("cmd", ["/c", "exit", "1"])`
	} else {
		script = `cmd.exec("sh", ["-c", "exit 1"])`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	// exitCode 可能是 int 或 int64
	exitCode := result["exitCode"]
	switch v := exitCode.(type) {
	case int:
		assert.Equal(t, 1, v)
	case int64:
		assert.Equal(t, int64(1), v)
	default:
		t.Fatalf("unexpected exitCode type: %T", exitCode)
	}
}

func TestCmdModule_Duration(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.exec("cmd", ["/c", "echo", "duration test"])`
	} else {
		script = `cmd.exec("echo", ["duration test"])`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	// duration 可能是 int 或 int64
	duration := result["duration"]
	switch v := duration.(type) {
	case int:
		assert.GreaterOrEqual(t, v, 0)
	case int64:
		assert.GreaterOrEqual(t, v, int64(0))
	default:
		t.Fatalf("unexpected duration type: %T", duration)
	}
}

func TestCmdModule_InvalidCommand(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	// 执行不存在的命令
	script := `cmd.exec("nonexistent_command_xyz", [])`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
}

func TestCmdModule_EmptyCommand(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	script := `cmd.exec([])`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "empty")
}

func TestCmdModule_InvalidCommandType(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	script := `cmd.exec(123)`
	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["error"], "invalid")
}

func TestCmdModule_ExecSync(t *testing.T) {
	module := NewCmdModule(nil)
	vm := goja.New()
	_ = module.Register(vm)

	var script string
	if runtime.GOOS == "windows" {
		script = `cmd.execSync("cmd", ["/c", "echo", "sync test"])`
	} else {
		script = `cmd.execSync("echo", ["sync test"])`
	}

	value, err := vm.RunString(script)
	assert.NoError(t, err)
	result := value.Export().(map[string]interface{})
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result["stdout"].(string), "sync test")
}
