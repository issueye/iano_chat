package routes

import (
	"iano_chat/pkg/web"
	"iano_chat/pkg/web/middleware"
	"log/slog"
	"time"
)

func SetupRoutes(logger *slog.Logger) *web.Engine {
	engine := web.New()

	engine.Use(middleware.RecoveryWithLog(logger))
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())
	engine.Use(middleware.IPRateLimit(1000, 10*time.Second))

	return engine
}
