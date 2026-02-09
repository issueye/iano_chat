# SSE (Server-Sent Events) 使用指南

## 概述

pkg/web 提供了完整的 SSE (Server-Sent Events) 支持，允许服务器向客户端推送实时事件。

## 基本使用

### 1. 简单 SSE 处理器

```go
engine := web.New()

engine.GET("/events", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
    // 发送简单数据
    sse.EmitData("Hello, SSE!")
    
    // 发送带类型的事件
    sse.EmitEvent("message", "This is a message")
    
    // 发送带 ID 的事件
    sse.EmitID("123", "Event with ID")
    
    // 发送完整事件
    sse.Emit(&web.SSEvent{
        ID:    "456",
        Event: "update",
        Data:  map[string]string{"status": "ok"},
        Retry: 5000, // 重连时间（毫秒）
    })
}))
```

### 2. 使用 SSE 中间件

```go
config := web.SSEMiddlewareConfig{
    HeartbeatInterval: 30 * time.Second,  // 心跳间隔
    OnConnect: func(sse *web.SSEContext, c *web.Context) {
        // 客户端连接时触发
        sse.EmitData("Connected!")
    },
    OnDisconnect: func(clientID string) {
        // 客户端断开时触发
        log.Printf("Client %s disconnected", clientID)
    },
}

engine.GET("/sse", web.SSEMiddleware(config))
```

### 3. 手动创建 SSE 上下文

```go
engine.GET("/custom", func(c *web.Context) {
    sse, err := c.SSE()
    if err != nil {
        c.String(500, "SSE not supported")
        return
    }
    
    // 启动心跳
    ticker := sse.StartHeartbeat(30 * time.Second)
    defer ticker.Stop()
    
    // 发送事件
    sse.EmitData("Hello!")
    
    // 关闭连接
    sse.Close()
})
```

## 高级用法

### 广播给所有客户端

```go
// 创建 SSE Hub
hub := web.NewSSEHub()
go hub.Run()

// 注册客户端
engine.GET("/events", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
    hub.Register(sse)
    
    // 保持连接
    <-sse.Done()
    
    // 注销
    hub.Unregister(c.Request.RemoteAddr)
}))

// 广播消息（在另一个 goroutine 中）
hub.Broadcast(&web.SSEvent{
    Event: "update",
    Data:  "New data available",
})

// 或简单广播
hub.BroadcastData("Simple message")
hub.BroadcastEvent("notification", "You have a new notification")
```

### 结合 Context 使用

```go
engine.GET("/events", func(c *web.Context) {
    sse, _ := c.SSE()
    
    // 创建可取消的上下文
    ctx, cancel := sse.ContextWithCancel(context.Background())
    defer cancel()
    
    // 使用上下文
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                if err := sse.EmitData(time.Now().String()); err != nil {
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }()
    
    // 等待连接关闭
    <-sse.Done()
})
```

### 心跳机制

```go
engine.GET("/events", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
    // 自动启动心跳（默认 30 秒）
    
    // 或自定义心跳间隔
    ticker := sse.StartHeartbeat(10 * time.Second)
    defer ticker.Stop()
    
    // 手动发送心跳
    sse.Ping()  // 发送 event: ping
}))
```

## SSE 事件格式

### 标准 SSE 格式

```
id: 123
event: message
retry: 5000
data: {"status": "ok"}

```

### 多行数据

```go
sse.EmitData("Line 1\nLine 2\nLine 3")
```

输出：
```
data: Line 1
data: Line 2
data: Line 3

```

## 客户端示例 (JavaScript)

```javascript
// 创建 EventSource
const eventSource = new EventSource('/events');

// 监听消息
eventSource.onmessage = (event) => {
    console.log('Received:', event.data);
};

// 监听特定类型的事件
eventSource.addEventListener('update', (event) => {
    console.log('Update:', event.data);
});

// 监听心跳
eventSource.addEventListener('ping', (event) => {
    console.log('Ping received');
});

// 错误处理
eventSource.onerror = (error) => {
    console.error('SSE error:', error);
};

// 关闭连接
eventSource.close();
```

## 配置选项

### SSEvent 结构

```go
type SSEvent struct {
    ID    string      // 事件 ID（可选，用于断线重连）
    Event string      // 事件类型（可选）
    Data  interface{} // 数据（必需）
    Retry int         // 重连时间（毫秒，可选）
}
```

### SSEMiddlewareConfig

```go
type SSEMiddlewareConfig struct {
    HeartbeatInterval time.Duration               // 心跳间隔
    OnConnect         func(*SSEContext, *Context) // 连接回调
    OnDisconnect      func(string)                // 断开回调
}
```

## 最佳实践

1. **始终启动心跳**: 防止连接被代理服务器关闭
2. **处理错误**: Emit 可能失败（连接已关闭）
3. **优雅关闭**: 使用 `sse.Close()` 清理资源
4. **使用 ID**: 便于客户端断线重连时恢复
5. **限制并发**: 大量 SSE 连接会消耗资源

## 与 WebSocket 的区别

| 特性 | SSE | WebSocket |
|------|-----|-----------|
| 方向 | 单向（服务器→客户端） | 双向 |
| 协议 | HTTP | WebSocket |
| 自动重连 | ✅ | 需手动实现 |
| 跨域 | 需 CORS | ✅ |
| 二进制数据 | ❌ | ✅ |
| 实时性 | 高 | 极高 |

**选择建议**:
- 只需要服务器推送 → SSE
- 需要双向通信 → WebSocket
- 简单的实时更新 → SSE
- 游戏、聊天室 → WebSocket

## 完整示例

```go
package main

import (
    "time"
    "iano_chat/pkg/web"
)

func main() {
    engine := web.New()
    
    // 创建广播 Hub
    hub := web.NewSSEHub()
    go hub.Run()
    
    // SSE 端点
    engine.GET("/events", web.HandleSSE(func(sse *web.SSEContext, c *web.Context) {
        // 注册客户端
        hub.Register(sse)
        defer hub.Unregister(c.Request.RemoteAddr)
        
        // 发送欢迎消息
        sse.EmitEvent("connected", map[string]string{
            "client_id": c.Request.RemoteAddr,
            "time":      time.Now().String(),
        })
        
        // 保持连接
        <-sse.Done()
    }))
    
    // 触发广播的端点
    engine.POST("/notify", func(c *web.Context) {
        var msg struct {
            Content string `json:"content"`
        }
        c.Bind(&msg)
        
        hub.Broadcast(&web.SSEvent{
            Event: "notification",
            Data:  msg.Content,
        })
        
        c.String(200, "Notified %d clients", hub.ClientCount())
    })
    
    engine.Run(":8080")
}
```

---

*文档版本: v1.0*  
*最后更新: 2026-02-09*
