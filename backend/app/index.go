package app

import (
	"iano_chat/models"
	"iano_chat/pkg/config"
	"iano_chat/pkg/database"
	"iano_chat/pkg/logger"
	"iano_chat/routes"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/gorm"
)

type App struct {
	AppName    string // 应用名称
	Version    string // 应用版本
	RootPath   string // 应用根目录
	ConfigPath string // 配置文件路径

	DB  *gorm.DB       // 数据库连接
	cfg *config.Config // 配置
	Log *slog.Logger   // 日志
}

func NewApp(rootPath string, configPath string) (*App, error) {
	a := &App{
		RootPath:   rootPath,
		ConfigPath: configPath,
	}
	// 初始化根目录
	if err := a.InitRootDirs(); err != nil {
		return nil, err
	}

	// 初始化配置
	a.InitConfig()

	// 初始化日志
	a.InitLogger()

	// 初始化数据库
	if err := a.InitDB(); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) InitDB() error {
	db, err := database.InitDB(a.cfg)
	if err != nil {
		return err
	}

	a.DB = db

	return a.DB.AutoMigrate(&models.User{})
}

func (a *App) InitLogger() {
	a.Log = logger.InitLogger(a.cfg)
}

func (a *App) InitConfig() {
	path := a.ConfigPath
	if path == "" {
		path = a.RootPath + "/config.toml"
	}

	a.cfg = config.Load(path)
}

func (a *App) InitRootDirs() error {
	dirs := []string{
		"root/logs",
		"root/data",
		"root/cache",
	}

	// 如果没有传入根目录，默认使用当前目录
	if a.RootPath == "" {
		a.RootPath = "."
	}

	// 拼接根目录
	for _, dir := range dirs {
		if err := os.MkdirAll(a.RootPath+"/"+dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Start() error {
	// 启动HTTP服务器
	router := routes.SetupRoutes()
	a.Log.Info("服务启动", slog.String("port", a.cfg.Server.Port))
	if err := http.ListenAndServe(":"+a.cfg.Server.Port, router); err != nil {
		a.Log.Error("服务启动失败", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (a *App) Shutdown() {
	a.Log.Info("服务关闭")
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func (a *App) WatchSignals() {
	// 监听信号量，ctrl+c 或 kill 命令触发
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	slog.Info("服务监听信号量", slog.String("signals", "SIGINT, SIGTERM, SIGQUIT"))
	<-sigChan
}
