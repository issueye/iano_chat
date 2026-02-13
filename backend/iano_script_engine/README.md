# Iano Script Engine

基于 [goja](https://github.com/dop251/goja) 实现的安全 JavaScript 脚本执行引擎，用于 Hook 和 Agent 的脚本执行。

## 特性

- 安全的 JavaScript 执行环境
- 支持超时控制
- 支持内存限制
- 内置模块支持（HTTP、Utils、URL）
- 完整的控制台日志支持
- JSON 处理支持

## 安装

```bash
go get github.com/issueye/iano_chat/backend/iano_script_engine
```

## 快速开始

### 脚本格式

脚本必须定义 `ScriptRun(input)` 函数作为入口点：

```javascript
function ScriptRun(input) {
    // input 是传入的参数对象
    return {
        result: input.name + "!",
        value: 42
    };
}
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "time"

    engine "github.com/issueye/iano_chat/backend/iano_script_engine"
)

func main() {
    // 创建引擎
    e := engine.NewEngine(nil)
    ctx := context.Background()

    // 定义脚本
    script := `
        function ScriptRun(input) {
            return {
                greeting: "Hello, " + input.name + "!",
                timestamp: Date.now()
            };
        }
    `

    // 执行脚本
    result, err := e.Execute(ctx, script, map[string]interface{}{
        "name": "World",
    })
    if err != nil {
        panic(err)
    }

    if result.Success {
        fmt.Printf("Result: %v\n", result.Value)
    } else {
        fmt.Printf("Error: %s\n", result.Error)
    }
}
```

### 带超时执行

```go
result, err := e.ExecuteWithTimeout(script, input, 5*time.Second)
```

### 验证脚本

```go
err := e.Validate(script)
if err != nil {
    fmt.Printf("Invalid script: %v\n", err)
}
```

## 内置对象

### console

```javascript
function ScriptRun(input) {
    console.log("info message");
    console.debug("debug message");
    console.info("info message");
    console.warn("warn message");
    console.error("error message");
    return { done: true };
}
```

### JSON

```javascript
function ScriptRun(input) {
    var obj = { name: "test", value: 42 };
    var jsonStr = JSON.stringify(obj);
    var parsed = JSON.parse(jsonStr);
    return { original: jsonStr, parsed: parsed };
}
```

### Date

```javascript
function ScriptRun(input) {
    return {
        now: Date.now(),
        timestamp: new Date().toISOString()
    };
}
```

### sleep

```javascript
function ScriptRun(input) {
    sleep(1000); // 暂停 1 秒
    return { done: true };
}
```

## 内置模块

### HTTP 模块

```javascript
function ScriptRun(input) {
    // GET 请求
    var response = http.get("https://api.example.com/data", {
        params: { page: 1, limit: 10 }
    });
    
    // POST 请求
    var postResponse = http.post("https://api.example.com/create", {
        json: { name: "test", value: 42 }
    });
    
    return {
        getStatus: response.status,
        postStatus: postResponse.status
    };
}
```

### Utils 模块

```javascript
function ScriptRun(input) {
    return {
        uuid: utils.uuid(),
        lower: utils.string.toLower("HELLO"),
        upper: utils.string.toUpper("hello"),
        contains: utils.string.contains("hello world", "world"),
        random: utils.random.int(100)
    };
}
```

### URL 模块

```javascript
function ScriptRun(input) {
    var parsed = url.parse("https://example.com/path?query=value");
    return {
        scheme: parsed.scheme,
        host: parsed.host,
        path: parsed.path,
        encoded: url.encode("hello world"),
        decoded: url.decode("hello%20world")
    };
}
```

## 执行器

### ScriptExecutor

通用脚本执行器：

```go
config := &engine.ExecutorConfig{
    DefaultTimeout: 30 * time.Second,
    MaxTimeout:     5 * time.Minute,
    EnableHTTP:     true,
    EnableUtils:    true,
    EnableURL:      true,
}
executor := engine.NewExecutor(config)
```

### HookScriptExecutor

专门用于 Hook 的脚本执行器：

```go
executor := engine.NewHookScriptExecutor()
result, err := executor.ExecuteHook(ctx, script, "user.created", map[string]interface{}{
    "userId": 123,
    "name": "John",
})
```

### AgentScriptExecutor

专门用于 Agent 的脚本执行器：

```go
executor := engine.NewAgentScriptExecutor()

// 执行工具
result, err := executor.ExecuteTool(ctx, script, "calculator", map[string]interface{}{
    "a": 10,
    "b": 20,
})

// 数据转换
result, err := executor.ExecuteTransform(ctx, script, rawData)

// 过滤
match, err := executor.ExecuteFilter(ctx, script, item)
```

### Sandbox

安全沙箱执行环境：

```go
limits := &engine.SandboxLimits{
    MaxExecutionTime: 5 * time.Second,
    MaxMemoryMB:      50,
    MaxOutputSize:    1024 * 1024,
    AllowedModules:   []string{"utils", "url"},
    BlockedFunctions: []string{"http", "eval", "Function"},
}
sandbox := engine.NewSandbox(limits)
result, err := sandbox.Run(ctx, script, input)
```

## 设置全局变量和函数

```go
engine := engine.NewEngine(nil).(*engine.GojaEngine)

// 设置全局变量
engine.SetGlobal("myVar", "hello")
engine.SetGlobal("myNum", 42)

// 设置全局函数
engine.SetFunction("greet", func(name string) string {
    return "Hello, " + name
})
```

脚本中使用：

```javascript
function ScriptRun(input) {
    return {
        message: myVar + " " + myNum,
        greeting: greet("World")
    };
}
```

## 执行结果

```go
type Result struct {
    Success bool        `json:"success"`        // 是否成功
    Value   interface{} `json:"value,omitempty"` // 返回值
    Error   string      `json:"error,omitempty"` // 错误信息
    Logs    []LogEntry  `json:"logs,omitempty"`  // 日志记录
}

type LogEntry struct {
    Level   string `json:"level"`   // 日志级别
    Message string `json:"message"` // 日志消息
    Time    int64  `json:"time"`    // 时间戳
}
```

## 配置选项

```go
type Config struct {
    Timeout          time.Duration // 默认执行超时
    MemoryLimit      uint64        // 内存限制（字节）
    MaxCallStackSize int           // 最大调用栈深度
}
```

## 错误处理

脚本执行错误会返回在 `Result.Error` 中：

```go
result, err := e.Execute(ctx, script, input)
if err != nil {
    // 系统级错误
    panic(err)
}

if !result.Success {
    // 脚本执行错误
    fmt.Printf("Script error: %s\n", result.Error)
}
```

常见错误：
- `script must define a ScriptRun function` - 脚本未定义 `ScriptRun` 函数
- `ScriptRun must be a function` - `ScriptRun` 不是函数类型
- `script execution timeout` - 脚本执行超时
- `script error: ...` - JavaScript 运行时错误

## 许可证

MIT License
