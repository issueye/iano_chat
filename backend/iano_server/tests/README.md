# HTTP API 测试

本项目为 `iano_server` 提供了完整的 HTTP API 接口测试套件。

## 测试文件结构

```
tests/
├── test_helper.go              # 测试辅助函数和工具
├── agent_controller_test.go    # Agent 控制器测试
├── session_controller_test.go  # Session 控制器测试
├── message_controller_test.go  # Message 控制器测试
├── tool_controller_test.go     # Tool 控制器测试
└── integration_test.go         # 集成测试
```

## 运行测试

### 运行所有测试

```bash
cd backend/iano_server
go test ./tests/... -v
```

### 运行特定测试文件

```bash
# 只运行 Agent 控制器测试
go test ./tests/... -v -run TestAgentController

# 只运行 Session 控制器测试
go test ./tests/... -v -run TestSessionController

# 只运行集成测试
go test ./tests/... -v -run TestIntegration
```

### 运行特定测试用例

```bash
# 运行特定的测试用例
go test ./tests/... -v -run TestAgentController/Create_Agent
```

### 生成测试覆盖率报告

```bash
# 生成覆盖率报告
go test ./tests/... -cover

# 生成详细的覆盖率报告并打开
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 测试特点

### 1. 使用内存数据库

所有测试都使用 SQLite 内存数据库 (`:memory:`)，确保：
- 测试之间相互隔离
- 不会污染真实数据
- 测试执行速度快

### 2. 测试辅助函数

`test_helper.go` 提供了以下辅助函数：

- `NewTestDB()` - 创建测试数据库实例
- `MakeRequest()` - 创建 HTTP 请求
- `ParseResponse()` - 解析 JSON 响应
- `AssertStatusCode()` - 断言 HTTP 状态码
- `AssertSuccess()` - 断言成功响应
- `AssertError()` - 断言错误响应

### 3. 测试覆盖

测试覆盖了以下 API：

#### Agent API
- `POST /api/agents` - 创建 Agent
- `GET /api/agents` - 获取所有 Agent
- `GET /api/agents/:id` - 根据 ID 获取 Agent
- `GET /api/agents/type` - 根据类型获取 Agent
- `PUT /api/agents/:id` - 更新 Agent
- `DELETE /api/agents/:id` - 删除 Agent

#### Session API
- `POST /api/sessions` - 创建 Session
- `GET /api/sessions` - 获取所有 Session
- `GET /api/sessions/:id` - 根据 ID 获取 Session
- `GET /api/sessions/user` - 根据用户 ID 获取 Session
- `GET /api/sessions/status` - 根据状态获取 Session
- `PUT /api/sessions/:id` - 更新 Session
- `DELETE /api/sessions/:id` - 删除 Session
- `GET /api/sessions/:id/config` - 获取 Session 配置
- `PUT /api/sessions/:id/config` - 更新 Session 配置

#### Message API
- `POST /api/messages` - 创建消息
- `GET /api/messages` - 获取所有消息
- `GET /api/messages/:id` - 根据 ID 获取消息
- `GET /api/messages/session` - 根据 Session ID 获取消息
- `GET /api/messages/user` - 根据用户 ID 获取消息
- `GET /api/messages/type` - 根据类型获取消息
- `PUT /api/messages/:id` - 更新消息
- `DELETE /api/messages/:id` - 删除消息
- `POST /api/messages/:id/feedback` - 添加消息反馈

#### Tool API
- `POST /api/tools` - 创建工具
- `GET /api/tools` - 获取所有工具
- `GET /api/tools/:id` - 根据 ID 获取工具
- `GET /api/tools/type` - 根据类型获取工具
- `GET /api/tools/status` - 根据状态获取工具
- `PUT /api/tools/:id` - 更新工具
- `PUT /api/tools/:id/config` - 更新工具配置
- `DELETE /api/tools/:id` - 删除工具

#### 其他 API
- `GET /health` - 健康检查

### 4. 集成测试

`integration_test.go` 包含：
- 完整工作流程测试
- 并发请求测试
- 错误处理测试

## 编写新测试

参考以下模板编写新的测试：

```go
func TestYourFeature(t *testing.T) {
    // 创建测试数据库
    testDB, err := NewTestDB()
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    defer testDB.Close()

    // 创建路由引擎
    engine := routes.SetupRoutes(testDB.DB)

    t.Run("Test Case Name", func(t *testing.T) {
        // 准备请求数据
        reqBody := `{"field": "value"}`

        // 创建请求
        req := httptest.NewRequest(http.MethodPost, "/api/endpoint", strings.NewReader(reqBody))
        req.Header.Set("Content-Type", "application/json")
        rr := httptest.NewRecorder()

        // 执行请求
        engine.ServeHTTP(rr, req)

        // 断言结果
        AssertStatusCode(t, rr, http.StatusOK)
        response := ParseResponse(t, rr)
        AssertSuccess(t, response)
    })
}
```

## 注意事项

1. 每个测试都会创建独立的数据库实例，测试之间不会相互影响
2. 测试完成后会自动清理资源
3. 使用 `t.Parallel()` 可以并行运行测试（谨慎使用，可能导致数据库冲突）
4. 测试中的数据不会持久化，每次测试都是全新的环境
