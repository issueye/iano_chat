# iano_chat - Wails v2 桌面应用

这是一个基于 Wails v2 的桌面应用程序，使用 Go 后端和 Vue 3 前端。

## 项目结构

```
iano_chat/
├── main.go              # Wails 应用程序入口
├── app.go               # Wails 应用程序逻辑和绑定方法
├── wails.json           # Wails 配置文件
├── go.mod               # Go 模块定义
├── frontend/            # Vue 3 前端项目
│   ├── src/
│   │   ├── lib/wails/   # Wails 运行时和 Go 绑定
│   │   └── main.js      # 前端入口
│   └── vite.config.js   # Vite 配置
└── backend/             # 现有后端模块
    ├── iano_agent/
    ├── iano_script_engine/
    ├── iano_server/
    └── iano_web/
```

## 环境要求

- Go 1.21+
- Node.js 18+
- Wails CLI v2.10.1+

## 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 开发

### 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend
npm install
```

### 启动开发服务器

```bash
# 使用 Wails CLI
wails dev

# 或在前端目录中使用 npm
npm run wails:dev
```

开发服务器将在 http://localhost:34115 启动，并启用热重载。

### 构建生产版本

```bash
# 构建当前平台的应用程序
wails build

# 构建所有平台的应用程序
wails build -platform windows,linux,darwin

# 或在前端目录中
npm run wails:build
```

## Go 绑定方法

在 `app.go` 中定义的方法会自动暴露给前端 JavaScript：

- `Greet(name string) string` - 问候语
- `GetOS() string` - 获取操作系统类型
- `Minimize()` - 最小化窗口
- `Maximize()` - 最大化窗口
- `Unmaximize()` - 恢复窗口大小
- `Close()` - 关闭应用程序

## 前端使用 Go 绑定

```javascript
// 在 Vue 组件中
const result = await window.go.main.App.Greet('World')
console.log(result) // "Hello World!"

// 或使用生成的 TypeScript 绑定
import { Greet } from '@/lib/wails/go/main/App'
const result = await Greet('World')
```

## 与现有后端集成

当前 Wails 应用可以与现有的后端模块集成：

```go
import (
    "iano_chat/backend/iano_agent"
    "iano_chat/backend/iano_server"
)

func (a *App) SomeMethod() {
    // 使用现有的后端模块
}
```

## 配置说明

### wails.json

- `frontend.dir`: 前端目录路径
- `frontend.dev`: 开发服务器命令
- `frontend.build`: 构建命令
- `wailsjsDir`: Wails JS 运行时生成目录

### vite.config.js

- `server.port`: 开发服务器端口 (34115)
- `server.strictPort`: 强制使用指定端口
- `build.outDir`: 构建输出目录 (dist)

## 注意事项

1. 开发时请确保端口 34115 未被占用
2. 首次运行 `wails dev` 会自动生成 TypeScript 绑定
3. 前端代理配置仅在浏览器模式下有效，Wails 模式下请使用 Go 绑定
4. 确保 `go.work` 包含根目录模块，以便导入后端包

## 许可证

Copyright © 2024
