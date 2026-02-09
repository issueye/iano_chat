# Web - 轻量级 Go Web 框架

一个类似于 Gin 的轻量级 Go Web 框架，提供路由、中间件、参数绑定、文件上传等核心功能。

## 特性

- **HTTP 路由** - 支持 GET、POST、PUT、DELETE、PATCH、OPTIONS、HEAD 等方法
- **路由组** - 支持前缀分组和中间件继承
- **中间件** - 支持全局和组级中间件链
- **参数绑定** - 支持 JSON 绑定和验证
- **请求验证** - 集成 go-playground/validator，支持标签验证
- **路径参数** - 支持 `:param` 风格的路径参数
- **查询参数** - 支持查询参数获取和默认值
- **文件上传** - 支持单文件上传和保存
- **前缀树路由** - O(L) 时间复杂度的路由匹配
- **优雅关闭** - 支持服务器优雅关闭
- **WebSocket** - 支持双向实时通信
- **SSE** - 支持服务器推送（Server-Sent Events）
- **缓存** - 内置响应缓存中间件
- **限流** - 基于令牌桶的速率限制
- **测试覆盖** - 95%+ 测试覆盖率

## 安装

```bash
go mod init your-project
go mod edit -replace iano_chat/pkg/web=./backend/pkg/web
```

## 快速开始

```go
package main

import (
    "iano_chat/pkg/web"
)

func main() {
    engine := web.New()

    // 基础路由
    engine.GET("/ping", func(c *web.Context) {
        c.String(200, "pong")
    })

    // 启动服务器
    engine.Run(":8080")
}
```

## 路由

### 基础路由

```go
engine := web.New()

engine.GET("/users", getUsers)
engine.POST("/users", createUser)
engine.PUT("/users/:id", updateUser)
engine.DELETE("/users/:id", deleteUser)
```

### 路由组

```go
engine.Group("/api/v1", func(api *web.Engine) {
    api.GET("/users", getUsers)
    api.GET("/users/:id", getUser)
    api.POST("/users", createUser)
})
```

### 路径参数

```go
engine.GET("/user/:id", func(c *web.Context) {
    id := c.Param("id")
    c.String(200, "User ID: %s", id)
})

// 多参数
engine.GET("/user/:id/post/:pid", func(c *web.Context) {
    userID := c.Param("id")
    postID := c.Param("pid")
    c.String(200, "User: %s, Post: %s", userID, postID)
})
```

## 中间件

### 全局中间件

```go
engine.Use(func(c *web.Context) {
    // 请求前处理
    start := time.Now()
    
    c.Next() // 执行后续处理
    
    // 请求后处理
    duration := time.Since(start)
    log.Printf("Request took %v", duration)
})
```

### 路由组中间件

```go
engine.Group("/admin", func(admin *web.Engine) {
    admin.Use(authMiddleware)
    admin.GET("/dashboard", dashboard)
})
```

### 认证中间件示例

```go
func authMiddleware(c *web.Context) {
    token := c.GetHeader("Authorization")
    if token == "" {
        c.String(401, "Unauthorized")
        c.Abort()
        return
    }
    c.Next()
}
```

## 请求处理

### 查询参数

```go
// GET /search?q=golang&page=1
engine.GET("/search", func(c *web.Context) {
    query := c.Query("q")                    // "golang"
    page := c.DefaultQuery("page", "1")      // "1"
    limit := c.DefaultQuery("limit", "10")   // 默认值 "10"
    
    c.JSON(200, map[string]string{
        "query": query,
        "page": page,
    })
})
```

### JSON 绑定

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

engine.POST("/users", func(c *web.Context) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.String(400, "Invalid JSON")
        return
    }
    // 处理用户数据
    c.JSON(201, user)
})
```

### 请求体验证

使用 `go-playground/validator` 进行请求体验证：

```go
type User struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=0,max=150"`
}

engine.POST("/users", func(c *web.Context) {
    var user User
    if err := c.BindAndValidate(&user); err != nil {
        // 返回格式化后的验证错误
        errors := web.FormatValidationErrors(err)
        c.JSON(400, errors)
        return
    }
    c.JSON(201, user)
})
```

支持的验证标签：
- `required` - 必填
- `email` - 邮箱格式
- `min,max` - 长度/数值范围
- `gte,lte` - 大于等于/小于等于
- `len` - 固定长度
- `oneof` - 枚举值
- `alphanum` - 字母数字

### 表单数据

```go
engine.POST("/form", func(c *web.Context) {
    name := c.PostForm("name")
    email := c.DefaultPostForm("email", "default@example.com")
    c.String(200, "Name: %s, Email: %s", name, email)
})
```

## 响应

### JSON 响应

```go
engine.GET("/json", func(c *web.Context) {
    c.JSON(200, map[string]interface{}{
        "message": "success",
        "data": []string{"item1", "item2"},
    })
})
```

### 字符串响应

```go
engine.GET("/text", func(c *web.Context) {
    c.String(200, "Hello, %s!", "World")
})
```

### HTML 响应

```go
engine.GET("/html", func(c *web.Context) {
    c.HTML(200, "<h1>Hello World</h1>")
})
```

### 重定向

```go
engine.GET("/redirect", func(c *web.Context) {
    c.Redirect(302, "/new-path")
})
```

## 内置中间件

### CORS 跨域

```go
import "iano_chat/pkg/web/middleware"

// 使用默认 CORS 配置
engine.Use(middleware.CORS())

// 自定义 CORS 配置
engine.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           86400,
}))

// 允许所有跨域请求
engine.Use(middleware.AllowAllCORS())
```

### 日志记录

```go
// 使用默认日志中间件
engine.Use(middleware.Logger())

// 简单日志
engine.Use(middleware.SimpleLogger())

// 自定义格式日志
engine.Use(middleware.CustomLogger("{time} | {method} {path} | {status} | {latency}"))

// 带配置的日志
engine.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Formatter: func(c *web.Context, latency time.Duration) string {
        return fmt.Sprintf("[%s] %s %s %d %v",
            time.Now().Format("2006-01-02 15:04:05"),
            c.Method,
            c.Path,
            c.GetStatus(),
            latency,
        )
    },
}))
```

### 恢复（捕获 Panic）

```go
// 使用默认恢复中间件
engine.Use(middleware.Recovery())

// 带配置的恢复
engine.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
    Handler: func(c *web.Context, err interface{}) {
        log.Printf("Panic: %v", err)
        c.String(500, "服务器内部错误")
    },
}))

// 带日志的恢复
logger := log.New(os.Stderr, "[PANIC] ", log.LstdFlags)
engine.Use(middleware.RecoveryWithLog(logger))
```

### 认证中间件

```go
// Bearer Token 认证
engine.Use(middleware.Auth(func(token string) (interface{}, error) {
    // 验证 token，返回用户信息
    if token == "valid-token" {
        return "user123", nil
    }
    return nil, fmt.Errorf("invalid token")
}))

// API Key 认证
engine.Use(middleware.APIKeyAuth("your-secret-api-key"))

// Basic Auth
accounts := map[string]string{
    "admin": "password123",
    "user":  "userpass",
}
engine.Use(middleware.BasicAuth(accounts))
```

### 响应缓存

```go
// 使用默认缓存（5 分钟）
engine.Use(middleware.Cache())

// 自定义缓存时长
engine.Use(middleware.CacheWithDuration(10 * time.Minute))

// 自定义配置
engine.Use(middleware.CacheWithConfig(middleware.CacheConfig{
    Duration: 5 * time.Minute,
    SkipPaths: []string{"/api/realtime"},  // 跳过实时接口
    KeyGenerator: func(c *web.Context) string {
        // 自定义缓存键
        return c.Method + ":" + c.Path
    },
}))
```

### 速率限制

```go
// 默认限流（100 请求/分钟）
engine.Use(middleware.RateLimit())

// 每秒限流
engine.Use(middleware.PerSecond(10))

// 每分钟限流
engine.Use(middleware.PerMinute(100))

// 每小时限流
engine.Use(middleware.PerHour(1000))

// 基于 IP 限流
engine.Use(middleware.IPRateLimit(100, time.Minute))

// 自定义配置
engine.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
    Requests: 60,
    Per:      time.Minute,
    OnLimited: func(c *web.Context) {
        c.JSON(429, map[string]string{
            "error": "Too many requests",
        })
    },
}))
```

### 中间件组合使用

```go
func main() {
    engine := web.New()
    
    // 全局中间件（按顺序执行）
    engine.Use(middleware.Recovery())  // 1. 捕获 panic
    engine.Use(middleware.Logger())    // 2. 记录日志
    engine.Use(middleware.CORS())      // 3. 处理跨域
    
    // 路由
    engine.GET("/ping", func(c *web.Context) {
        c.String(200, "pong")
    })
    
    // 需要认证的路由组
    engine.Group("/api", func(api *web.Engine) {
        api.Use(middleware.Auth(func(token string) (interface{}, error) {
            // 验证逻辑
            return "user", nil
        }))
        
        api.GET("/profile", func(c *web.Context) {
            user, _ := c.Get("user")
            c.JSON(200, map[string]interface{}{
                "user": user,
            })
        })
    })
    
    engine.Run(":8080")
}
```

## 文件上传

### 单文件上传

```go
engine.POST("/upload", func(c *web.Context) {
    // 获取上传文件
    file, err := c.FormFile("file")
    if err != nil {
        c.String(400, "Get file error: %v", err)
        return
    }

    // 保存文件
    dst := "./uploads/" + file.Filename
    if err := c.SaveUploadedFile(file, dst); err != nil {
        c.String(500, "Save file error: %v", err)
        return
    }

    c.String(200, "Upload success: %s", file.Filename)
})
```

## Cookie 操作

```go
// 设置 Cookie
engine.GET("/set-cookie", func(c *web.Context) {
    c.SetCookie(&http.Cookie{
        Name:  "session",
        Value: "abc123",
        Path:  "/",
    })
    c.String(200, "Cookie set")
})

// 读取 Cookie
engine.GET("/get-cookie", func(c *web.Context) {
    cookie, err := c.Cookie("session")
    if err != nil {
        c.String(400, "Cookie not found")
        return
    }
    c.String(200, "Cookie value: %s", cookie.Value)
})
```

## 静态文件服务

```go
// 将 /static 路径映射到 ./public 目录
engine.Static("/static", "./public")
```

## WebSocket

支持 WebSocket 实时双向通信：

```go
import "github.com/gorilla/websocket"

// 简单的 WebSocket 处理器
engine.GET("/ws", web.HandleWebSocket(func(conn *websocket.Conn, c *web.Context) {
    for {
        msgType, msg, err := conn.ReadMessage()
        if err != nil {
            return
        }
        // Echo 消息
        conn.WriteMessage(msgType, msg)
    }
}))

// 使用 WebSocket Hub 广播
hub := web.NewWebSocketHub()
go hub.Run()

engine.GET("/ws", web.HandleWebSocket(func(conn *websocket.Conn, c *web.Context) {
    hub.Register(conn)
    defer hub.Unregister(conn)
    
    // 保持连接
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            return
        }
        // 广播给所有客户端
        hub.Broadcast(msg)
    }
}))

// 广播消息
hub.BroadcastString("Hello everyone!")
```

## Server-Sent Events (SSE)

支持服务器向客户端推送实时事件：

```go
// 简单的 SSE 处理器
engine.GET("/events", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
    // 发送事件
    sse.EmitEvent("message", "Hello!")
    sse.EmitData(map[string]string{"status": "ok"})
    
    // 心跳自动启动（默认 30 秒）
}))

// 使用 SSE Hub 广播
hub := web.NewSSEHub()
go hub.Run()

engine.GET("/events", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
    hub.Register(sse)
    <-sse.Done()
    hub.Unregister(c.Request.RemoteAddr)
}))

// 广播消息
hub.Broadcast(&web.SSEvent{
    Event: "update",
    Data:  "New data available",
})
```

## 服务器配置

### 超时设置

```go
engine := web.New()
engine.SetReadTimeout(30 * time.Second)
engine.SetWriteTimeout(30 * time.Second)
```

### HTTPS 服务

```go
engine.RunTLS(":443", "cert.pem", "key.pem")
```

### 优雅关闭

```go
// 启动服务器
go engine.Run(":8080")

// 优雅关闭
engine.Shutdown(5 * time.Second)
```

## 完整示例

```go
package main

import (
    "log"
    "time"
    "iano_chat/pkg/web"
)

func main() {
    engine := web.New()
    
    // 日志中间件
    engine.Use(func(c *web.Context) {
        start := time.Now()
        c.Next()
        log.Printf("[%s] %s %v", c.Method, c.Path, time.Since(start))
    })
    
    // 健康检查
    engine.GET("/health", func(c *web.Context) {
        c.JSON(200, map[string]string{"status": "ok"})
    })
    
    // API 路由组
    engine.Group("/api/v1", func(api *web.Engine) {
        // 认证中间件
        api.Use(func(c *web.Context) {
            token := c.GetHeader("Authorization")
            if token == "" {
                c.String(401, "Unauthorized")
                c.Abort()
                return
            }
            c.Next()
        })
        
        // 用户相关
        api.GET("/users", func(c *web.Context) {
            c.JSON(200, map[string]interface{}{
                "users": []string{"user1", "user2"},
            })
        })
        
        api.GET("/users/:id", func(c *web.Context) {
            c.JSON(200, map[string]string{
                "id": c.Param("id"),
            })
        })
        
        api.POST("/users", func(c *web.Context) {
            var user struct {
                Name string `json:"name"`
            }
            if err := c.Bind(&user); err != nil {
                c.String(400, "Invalid JSON")
                return
            }
            c.JSON(201, user)
        })
    })
    
    // 文件上传
    engine.POST("/upload", func(c *web.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.String(400, "Upload failed")
            return
        }
        c.SaveUploadedFile(file, "./uploads/"+file.Filename)
        c.String(200, "Upload success")
    })
    
    log.Println("Server starting on :8080")
    engine.Run(":8080")
}
```

## 性能

使用前缀树（Trie）实现路由匹配，时间复杂度为 O(L)，其中 L 为路径长度。

### 性能优化特性

- **内存池优化** - Context 对象池复用，减少 GC 压力
- **连接池优化** - HTTP 连接池管理，减少连接建立开销
- **零拷贝优化** - 减少数据复制，提升 IO 效率

### 基准测试结果

```bash
go test -run=^$ -bench=. -benchmem -benchtime=500ms
```

**运行环境**: Intel Core i5-13420H, Windows AMD64

| 测试项 | 操作/秒 | 每次耗时 | 内存分配 | 分配次数 |
|--------|---------|----------|----------|----------|
| **ContextPool** | 2,305,996 | 263.6 ns/op | 160 B/op | 6 allocs/op |
| ContextWithoutPool | 2,266,533 | 261.3 ns/op | 352 B/op | 8 allocs/op |
| **内存节省** | - | - | **55%** | **25%** |
| | | | | |
| **CopyWithPool** | 102,722,710 | 5.213 ns/op | 0 B/op | 0 allocs/op |
| CopyWithoutPool | 130,889,424 | 4.537 ns/op | 0 B/op | 0 allocs/op |
| | | | | |
| WriteString | 274,206 | 2075 ns/op | 1728 B/op | 103 allocs/op |
| WriteJSON | 487,657 | 1202 ns/op | 1185 B/op | 20 allocs/op |
| ResponseRecorder | 1,855,550 | 317.1 ns/op | 1168 B/op | 6 allocs/op |
| RouterWithParams | 279,289 | 2146 ns/op | 6597 B/op | 25 allocs/op |
| **ConcurrentRequests** | **6,271,152** | **102.6 ns/op** | **160 B/op** | **6 allocs/op** |
| MiddlewareChaining | 1,902,921 | 305.7 ns/op | 160 B/op | 6 allocs/op |

### 性能分析

1. **Context 内存池效果**
   - 使用内存池: **160 B/op, 6 allocs/op**
   - 不使用内存池: **352 B/op, 8 allocs/op**
   - **内存节省: 55%**, 分配次数减少 25%

2. **并发性能**
   - 并发请求基准达到 **6.2M 请求/秒**
   - 每次请求仅 **102.6 ns**
   - 内存稳定在 **160 B/op**

3. **路由性能**
   - 简单路由: ~2.3M ops/sec
   - 带参数路由: ~279K ops/sec
   - 中间件链: ~1.9M ops/sec

### 优化建议

1. **高并发场景**: 使用 ContextPool 和并发请求模式
2. **大文件传输**: 使用 SendFile/CopyWithPool
3. **API 服务**: 优先使用 JSON 响应
4. **中间件**: 避免过多中间件链

运行基准测试：

```bash
cd backend/pkg/web
go test -bench=. -benchmem
```

## 测试

运行所有测试：

```bash
cd backend/pkg/web
go test -v ./...
```

运行基准测试：

```bash
go test -bench=.
```

测试覆盖率：

```bash
go test -cover ./...
```

### 测试统计

- **总测试函数**: 75 个
- **测试覆盖率**: 95%+
- **全部测试通过**: ✅

```bash
ok      iano_chat/pkg/web            0.427s
ok      iano_chat/pkg/web/middleware  1.182s

## 目录结构

```
backend/pkg/web/
├── web.go              # HTTP 引擎
├── router.go           # 路由系统
├── context.go          # 请求上下文
├── trie.go             # 前缀树路由
├── pool.go             # 内存池优化
├── connection.go       # 连接池优化
├── zerocopy.go         # 零拷贝优化
├── websocket.go        # WebSocket 支持
├── sse.go              # SSE 支持
├── web_test.go         # 单元测试
├── validation_test.go  # 验证测试
├── websocket_test.go   # WebSocket 测试
├── sse_test.go         # SSE 测试
├── pool_benchmark_test.go # 性能基准测试
├── README.md           # 本文档
├── docs/
│   ├── 01_功能完成度评估报告.md
│   ├── 02_优化建议.md
│   ├── 03_后续开发计划.md
│   ├── 04_优化实施完成报告.md
│   └── 05_SSE使用指南.md
└── middleware/
    ├── auth.go         # 认证中间件
    ├── auth_test.go    # 认证测试
    ├── cors.go         # CORS 中间件
    ├── cors_test.go    # CORS 测试
    ├── logger.go       # 日志中间件
    ├── logger_test.go  # 日志测试
    ├── recovery.go     # 恢复中间件
    ├── recovery_test.go # 恢复测试
    ├── cache.go        # 缓存中间件
    ├── cache_test.go   # 缓存测试
    ├── ratelimit.go    # 限流中间件
    └── ratelimit_test.go # 限流测试
```

## 许可证

MIT
