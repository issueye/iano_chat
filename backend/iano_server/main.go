package main

import (
	"flag"
	"iano_server/app"
	"log/slog"
	"os"
)

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
