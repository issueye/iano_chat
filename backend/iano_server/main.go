package main

import (
	"flag"
	"iano_server/app"
	"log/slog"
	"os"
)

// @title IANO Chat API
// @version 1.0
// @description IANO Chat 是一个智能对话系统，支持多 Agent、工具调用、流式响应等功能。
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

var (
	rootPath   = flag.String("root", ".", "应用根目录")
	configPath = flag.String("config", "", "配置文件路径")
)

func main() {
	app, err := app.NewApp(*rootPath, *configPath)
	if err != nil {
		slog.Error("初始化服务失败", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 启动服务
	if err := app.Start(); err != nil {
		slog.Error("服务启动失败", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 监听
	app.WatchSignals()

	// 关闭
	app.Shutdown()
}
