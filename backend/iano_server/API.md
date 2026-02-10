# IANO Chat Server API 文档

## 基础信息

- **服务名称**: IANO Chat Server
- **基础URL**: `http://localhost:8080`
- **API版本**: v1
- **数据格式**: JSON
- **状态码说明**:
  - `200` - 成功
  - `201` - 创建成功
  - `400` - 请求参数错误
  - `404` - 资源不存在
  - `500` - 服务器内部错误

---

## 统一响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 失败响应
```json
{
  "code": 400,
  "message": "错误描述信息",
  "data": null
}
```

---

## 一、健康检查接口

### 1.1 健康检查

- **接口路径**: `/health`
- **请求方式**: `GET`
- **功能说明**: 检查服务是否正常运行

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

---

## 二、Agent 代理接口

### 2.1 创建 Agent

- **接口路径**: `/api/agents`
- **请求方式**: `POST`
- **功能说明**: 创建一个新的 Agent

**请求参数** (JSON Body):
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | Agent 名称 |
| description | string | 否 | Agent 描述 |
| type | string | 否 | Agent 类型 (main/sub/custom)，默认 main |
| is_sub_agent | boolean | 否 | 是否为子 Agent，默认 false |
| provider_id | string | 否 | Provider ID |
| model | string | 否 | 模型名称 |
| instructions | string | 否 | 指令配置 |
| tools | string | 否 | 工具配置 |

**请求示例**:
```json
{
  "name": "我的助手",
  "description": "一个智能助手",
  "type": "main",
  "is_sub_agent": false,
  "provider_id": "provider_001",
  "model": "gpt-4",
  "instructions": "你是一个有用的助手",
  "tools": "[]"
}
```

**响应示例** (201 Created):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "name": "我的助手",
    "description": "一个智能助手",
    "type": "main",
    "is_sub_agent": false,
    "provider_id": "provider_001",
    "model": "gpt-4",
    "instructions": "你是一个有用的助手",
    "tools": "[]",
    "created_at": "2026-02-10T10:00:00Z",
    "updated_at": "2026-02-10T10:00:00Z"
  }
}
```

### 2.2 获取所有 Agents

- **接口路径**: `/api/agents`
- **请求方式**: `GET`
- **功能说明**: 获取所有 Agent 列表

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "uuid-string",
      "name": "我的助手",
      "description": "...",
      "type": "main",
      "is_sub_agent": false,
      "provider_id": "provider_001",
      "model": "gpt-4",
      "instructions": "...",
      "tools": "[]",
      "created_at": "2026-02-10T10:00:00Z",
      "updated_at": "2026-02-10T10:00:00Z"
    }
  ]
}
```

### 2.3 按类型获取 Agents

- **接口路径**: `/api/agents/type`
- **请求方式**: `GET`
- **功能说明**: 按类型筛选 Agent 列表

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | string | 是 | Agent 类型 (main/sub/custom) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [ ... ]
}
```

### 2.4 获取单个 Agent

- **接口路径**: `/api/agents/:id`
- **请求方式**: `GET`
- **功能说明**: 根据 ID 获取 Agent 详情

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Agent ID |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "name": "我的助手",
    "description": "...",
    "type": "main",
    "is_sub_agent": false,
    "provider_id": "provider_001",
    "model": "gpt-4",
    "instructions": "...",
    "tools": "[]",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

**错误响应** (404 Not Found):
```json
{
  "code": 400,
  "message": "Agent not found",
  "data": null
}
```

### 2.5 更新 Agent

- **接口路径**: `/api/agents/:id`
- **请求方式**: `PUT`
- **功能说明**: 更新 Agent 信息

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Agent ID |

**请求参数** (JSON Body, 所有字段可选):
| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | Agent 名称 |
| description | string | Agent 描述 |
| type | string | Agent 类型 |
| is_sub_agent | boolean | 是否为子 Agent |
| provider_id | string | Provider ID |
| model | string | 模型名称 |
| instructions | string | 指令配置 |
| tools | string | 工具配置 |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 2.6 删除 Agent

- **接口路径**: `/api/agents/:id`
- **请求方式**: `DELETE`
- **功能说明**: 删除指定 Agent

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Agent ID |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Agent deleted successfully"
  }
}
```

---

## 三、Tool 工具接口

### 3.1 创建 Tool

- **接口路径**: `/api/tools`
- **请求方式**: `POST`
- **功能说明**: 创建一个新的 Tool

**请求参数** (JSON Body):
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | 工具名称 |
| desc | string | 否 | 工具描述 |
| returns | string | 否 | 返回值描述 |
| example | string | 否 | 使用示例 |
| type | string | 否 | 工具类型 (builtin/custom/external/plugin)，默认 builtin |
| config | string | 否 | 工具配置 (JSON) |
| parameters | string | 否 | 参数定义 (JSON) |
| version | string | 否 | 版本号，默认 1.0.0 |
| author | string | 否 | 作者 |

**请求示例**:
```json
{
  "name": "search",
  "desc": "搜索工具",
  "returns": "搜索结果列表",
  "example": "search('关键词')",
  "type": "builtin",
  "config": "{}",
  "parameters": "[{\"name\":\"query\",\"type\":\"string\",\"desc\":\"搜索关键词\",\"required\":true}]",
  "version": "1.0.0",
  "author": "admin"
}
```

**响应示例** (201 Created):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "name": "search",
    "desc": "搜索工具",
    "returns": "搜索结果列表",
    "example": "search('关键词')",
    "type": "builtin",
    "status": "enabled",
    "call_count": 0,
    "error_count": 0,
    "config": "{}",
    "parameters": "[{\"name\":\"query\",\"type\":\"string\",\"desc\":\"搜索关键词\",\"required\":true}]",
    "version": "1.0.0",
    "author": "admin"
  }
}
```

### 3.2 获取所有 Tools

- **接口路径**: `/api/tools`
- **请求方式**: `GET`
- **功能说明**: 获取所有 Tool 列表

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "uuid-string",
      "name": "search",
      "desc": "搜索工具",
      "type": "builtin",
      "status": "enabled",
      "call_count": 0,
      "error_count": 0,
      "version": "1.0.0",
      "author": "admin"
    }
  ]
}
```

### 3.3 按类型获取 Tools

- **接口路径**: `/api/tools/type`
- **请求方式**: `GET`
- **功能说明**: 按类型筛选 Tool 列表

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | string | 是 | Tool 类型 (builtin/custom/external/plugin) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [ ... ]
}
```

### 3.4 按状态获取 Tools

- **接口路径**: `/api/tools/status`
- **请求方式**: `GET`
- **功能说明**: 按状态筛选 Tool 列表

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| status | string | 是 | Tool 状态 (enabled/disabled/error) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [ ... ]
}
```

### 3.5 获取单个 Tool

- **接口路径**: `/api/tools/:id`
- **请求方式**: `GET`
- **功能说明**: 根据 ID 获取 Tool 详情

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Tool ID |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "name": "search",
    "desc": "搜索工具",
    "returns": "搜索结果列表",
    "type": "builtin",
    "status": "enabled",
    "call_count": 0,
    "error_count": 0,
    "config": "{}",
    "parameters": "[]",
    "version": "1.0.0",
    "author": "admin"
  }
}
```

### 3.6 更新 Tool

- **接口路径**: `/api/tools/:id`
- **请求方式**: `PUT`
- **功能说明**: 更新 Tool 信息

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Tool ID |

**请求参数** (JSON Body, 所有字段可选):
| 参数名 | 类型 | 说明 |
|--------|------|------|
| name | string | 工具名称 |
| desc | string | 工具描述 |
| returns | string | 返回值描述 |
| example | string | 使用示例 |
| type | string | 工具类型 |
| config | string | 工具配置 |
| parameters | string | 参数定义 |
| version | string | 版本号 |
| author | string | 作者 |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 3.7 更新 Tool 配置

- **接口路径**: `/api/tools/:id/config`
- **请求方式**: `PUT`
- **功能说明**: 更新 Tool 的配置信息

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Tool ID |

**请求参数** (JSON Body):
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| config | string | 是 | 工具配置 (JSON 字符串) |

**请求示例**:
```json
{
  "config": "{\"api_key\":\"xxx\",\"timeout\":30}"
}
```

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 3.8 删除 Tool

- **接口路径**: `/api/tools/:id`
- **请求方式**: `DELETE`
- **功能说明**: 删除指定 Tool

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Tool ID |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Tool deleted successfully"
  }
}
```

---

## 四、Message 消息接口

### 4.1 创建 Message

- **接口路径**: `/api/messages`
- **请求方式**: `POST`
- **功能说明**: 创建一条新消息

**请求参数** (JSON Body):
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_id | string | 是 | 会话 ID |
| key_id | string | 是 | Key ID |
| type | string | 是 | 消息类型 (user/assistant/system/tool) |
| content | string | 是 | 消息内容 |
| status | string | 否 | 消息状态 (sending/completed/failed/streaming)，默认 completed |
| parent_id | string | 否 | 父消息 ID |

**请求示例**:
```json
{
  "session_id": "session_001",
  "key_id": "key_001",
  "type": "user",
  "content": "你好，请帮我查询天气",
  "status": "completed",
  "parent_id": null
}
```

**响应示例** (201 Created):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "session_id": "session_001",
    "type": "user",
    "content": "你好，请帮我查询天气",
    "status": "completed",
    "input_tokens": 0,
    "output_tokens": 0,
    "parent_id": null,
    "feedback_rating": null,
    "feedback_comment": "",
    "created_at": "2026-02-10T10:00:00Z",
    "updated_at": "2026-02-10T10:00:00Z"
  }
}
```

### 4.2 获取所有 Messages

- **接口路径**: `/api/messages`
- **请求方式**: `GET`
- **功能说明**: 获取所有消息列表

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "uuid-string",
      "session_id": "session_001",
      "type": "user",
      "content": "你好",
      "status": "completed",
      "input_tokens": 10,
      "output_tokens": 0,
      "parent_id": null,
      "feedback_rating": null,
      "feedback_comment": "",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

### 4.3 按会话获取 Messages

- **接口路径**: `/api/messages/session`
- **请求方式**: `GET`
- **功能说明**: 获取指定会话的所有消息

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_id | string | 是 | 会话 ID (数字字符串) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [ ... ]
}
```

### 4.4 按类型获取 Messages

- **接口路径**: `/api/messages/type`
- **请求方式**: `GET`
- **功能说明**: 按类型筛选消息列表

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| type | string | 是 | 消息类型 (user/assistant/system/tool) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [ ... ]
}
```

### 4.5 获取单个 Message

- **接口路径**: `/api/messages/:id`
- **请求方式**: `GET`
- **功能说明**: 根据 ID 获取消息详情

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Message ID |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid-string",
    "session_id": "session_001",
    "type": "user",
    "content": "你好",
    "status": "completed",
    "input_tokens": 10,
    "output_tokens": 0,
    "parent_id": null,
    "feedback_rating": null,
    "feedback_comment": "",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

### 4.6 更新 Message

- **接口路径**: `/api/messages/:id`
- **请求方式**: `PUT`
- **功能说明**: 更新消息信息

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Message ID |

**请求参数** (JSON Body, 所有字段可选):
| 参数名 | 类型 | 说明 |
|--------|------|------|
| status | string | 消息状态 |
| content | string | 消息内容 |
| input_tokens | int | 输入 Token 数 |
| output_tokens | int | 输出 Token 数 |
| feedback_rating | string | 反馈评分 (like/dislike) |
| feedback_comment | string | 反馈评论 |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 4.7 删除 Message

- **接口路径**: `/api/messages/:id`
- **请求方式**: `DELETE`
- **功能说明**: 删除指定消息

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Message ID |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Message deleted successfully"
  }
}
```

### 4.8 添加消息反馈

- **接口路径**: `/api/messages/:id/feedback`
- **请求方式**: `POST`
- **功能说明**: 对消息添加反馈

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Message ID |

**请求参数** (JSON Body):
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| rating | string | 是 | 反馈评分 (like/dislike) |
| comment | string | 否 | 反馈评论 |

**请求示例**:
```json
{
  "rating": "like",
  "comment": "回答很有帮助"
}
```

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 4.9 按会话删除 Messages

- **接口路径**: `/api/messages`
- **请求方式**: `DELETE`
- **功能说明**: 删除指定会话的所有消息

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| session_id | string | 是 | 会话 ID (数字字符串) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Messages deleted successfully"
  }
}
```

---

## 五、Session 会话接口

### 5.1 创建 Session

- **接口路径**: `/api/sessions`
- **请求方式**: `POST`
- **功能说明**: 创建新的会话

**请求参数** (JSON Body):
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key_id | string | 是 | Key ID |
| title | string | 否 | 会话标题，默认 "新会话" |

**请求示例**:
```json
{
  "key_id": "key_001",
  "title": "我的第一个会话"
}
```

**响应示例** (201 Created):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "title": "我的第一个会话",
    "status": "active",
    "message_count": 0,
    "total_tokens": 0,
    "created_at": "2026-02-10T10:00:00Z",
    "updated_at": "2026-02-10T10:00:00Z",
    "last_active_at": "2026-02-10T10:00:00Z"
  }
}
```

### 5.2 获取所有 Sessions

- **接口路径**: `/api/sessions`
- **请求方式**: `GET`
- **功能说明**: 获取所有会话列表

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "title": "我的第一个会话",
      "status": "active",
      "message_count": 5,
      "total_tokens": 1000,
      "created_at": "2026-02-10T10:00:00Z",
      "updated_at": "2026-02-10T10:00:00Z",
      "last_active_at": "2026-02-10T10:00:00Z"
    }
  ]
}
```

### 5.3 按状态获取 Sessions

- **接口路径**: `/api/sessions/status`
- **请求方式**: `GET`
- **功能说明**: 按状态筛选会话列表

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| status | string | 是 | 会话状态 (active/paused/completed/archived) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": [ ... ]
}
```

### 5.4 获取单个 Session

- **接口路径**: `/api/sessions/:id`
- **请求方式**: `GET`
- **功能说明**: 根据 ID 获取会话详情

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | Session ID (数字) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "title": "我的第一个会话",
    "status": "active",
    "message_count": 5,
    "total_tokens": 1000,
    "created_at": "...",
    "updated_at": "...",
    "last_active_at": "..."
  }
}
```

### 5.5 更新 Session

- **接口路径**: `/api/sessions/:id`
- **请求方式**: `PUT`
- **功能说明**: 更新会话信息

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | Session ID (数字) |

**请求参数** (JSON Body, 所有字段可选):
| 参数名 | 类型 | 说明 |
|--------|------|------|
| title | string | 会话标题 |
| status | string | 会话状态 |
| config | object | 会话配置 |

**请求示例**:
```json
{
  "title": "更新后的会话标题",
  "status": "active",
  "config": {
    "model_id": 1,
    "temperature": 0.7,
    "max_tokens": 2000,
    "system_prompt": "你是一个有帮助的助手",
    "enable_tools": true,
    "selected_tools": ["search", "calculator"],
    "enable_summary": true,
    "keep_rounds": 4,
    "enable_rate_limit": true,
    "rate_limit_rpm": 20
  }
}
```

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 5.6 获取 Session 配置

- **接口路径**: `/api/sessions/:id/config`
- **请求方式**: `GET`
- **功能说明**: 获取会话的配置信息

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | Session ID (数字) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "model_id": 1,
    "temperature": 0.7,
    "max_tokens": 2000,
    "system_prompt": "你是一个有帮助的助手",
    "enable_tools": true,
    "selected_tools": ["search", "calculator"],
    "enable_summary": true,
    "keep_rounds": 4,
    "enable_rate_limit": true,
    "rate_limit_rpm": 20
  }
}
```

### 5.7 更新 Session 配置

- **接口路径**: `/api/sessions/:id/config`
- **请求方式**: `PUT`
- **功能说明**: 更新会话的配置信息

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | Session ID (数字) |

**请求参数** (JSON Body):
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| config | object | 是 | 会话配置对象 |

**请求示例**:
```json
{
  "config": {
    "model_id": 1,
    "temperature": 0.8,
    "max_tokens": 3000,
    "system_prompt": "你是一个专业的编程助手",
    "enable_tools": true,
    "selected_tools": ["search"],
    "enable_summary": true,
    "keep_rounds": 6,
    "enable_rate_limit": true,
    "rate_limit_rpm": 30
  }
}
```

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 5.8 删除 Session

- **接口路径**: `/api/sessions/:id`
- **请求方式**: `DELETE`
- **功能说明**: 删除指定会话

**路径参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | Session ID (数字) |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Session deleted successfully"
  }
}
```

### 5.9 按 Key ID 删除 Sessions

- **接口路径**: `/api/sessions`
- **请求方式**: `DELETE`
- **功能说明**: 删除指定 Key ID 的所有会话

**查询参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| key_id | string | 是 | Key ID |

**响应示例** (200 OK):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "Sessions deleted successfully"
  }
}
```

---

## 附录

### A. 枚举值说明

**Agent Type**:
- `main`: 主 Agent
- `sub`: 子 Agent
- `custom`: 自定义 Agent

**Tool Type**:
- `builtin`: 内置工具
- `custom`: 自定义工具
- `external`: 外部工具 (API 调用)
- `plugin`: 插件工具

**Tool Status**:
- `enabled`: 启用
- `disabled`: 禁用
- `error`: 错误

**Message Type**:
- `user`: 用户消息
- `assistant`: AI 助手消息
- `system`: 系统消息
- `tool`: 工具调用消息

**Message Status**:
- `sending`: 发送中
- `completed`: 已完成
- `failed`: 失败
- `streaming`: 流式输出中

**Feedback Rating**:
- `like`: 喜欢
- `dislike`: 不喜欢

**Session Status**:
- `active`: 活跃
- `paused`: 暂停
- `completed`: 已完成
- `archived`: 已归档

### B. Session 配置默认值

```json
{
  "temperature": 0.7,
  "max_tokens": 2000,
  "enable_tools": true,
  "selected_tools": [],
  "enable_summary": true,
  "keep_rounds": 4,
  "enable_rate_limit": true,
  "rate_limit_rpm": 20
}
```

### C. 数据模型关联

- **Session**: 会话主体，可以包含多个 Message
- **Message**: 消息，关联到 Session
- **Agent**: AI 代理，可配置 Tools
- **Tool**: 工具，可被 Agent 调用

---

## 接口统计

| 资源 | 接口数量 | HTTP 方法 |
|------|----------|-----------|
| Health | 1 | GET |
| Agents | 6 | POST, GET(x3), PUT, DELETE |
| Tools | 8 | POST, GET(x4), PUT(x2), DELETE |
| Messages | 9 | POST, GET(x4), PUT, DELETE(x2), POST |
| Sessions | 9 | POST, GET(x3), PUT(x2), DELETE(x2), GET, PUT |

**总计**: 33 个 API 接口
