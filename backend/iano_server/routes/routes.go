package routes

import (
	"iano_server/controllers"
	"iano_server/pkg/web"
	"iano_server/pkg/web/middleware"
	"iano_server/services"

	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB) *web.Engine {
	engine := web.New()

	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())

	agentService := services.NewAgentService(db)
	agentController := controllers.NewAgentController(agentService)
	messageService := services.NewMessageService(db)
	messageController := controllers.NewMessageController(messageService)
	sessionService := services.NewSessionService(db)
	sessionController := controllers.NewSessionController(sessionService)
	toolService := services.NewToolService(db)
	toolController := controllers.NewToolController(toolService)
	baseController := &controllers.BaseController{}

	engine.GET("/health", func(c *web.Context) {
		baseController.HealthCheck(c)
	})

	engine.POST("/api/tools", toolController.Create)
	engine.GET("/api/tools", toolController.GetAll)
	engine.GET("/api/tools/type", toolController.GetByType)
	engine.GET("/api/tools/status", toolController.GetByStatus)
	engine.GET("/api/tools/:id", toolController.GetByID)
	engine.PUT("/api/tools/:id", toolController.Update)
	engine.PUT("/api/tools/:id/config", toolController.UpdateConfig)
	engine.DELETE("/api/tools/:id", toolController.Delete)

	engine.POST("/api/agents", agentController.Create)
	engine.GET("/api/agents", agentController.GetAll)
	engine.GET("/api/agents/type", agentController.GetByType)
	engine.GET("/api/agents/:id", agentController.GetByID)
	engine.PUT("/api/agents/:id", agentController.Update)
	engine.DELETE("/api/agents/:id", agentController.Delete)

	engine.POST("/api/messages", messageController.Create)
	engine.GET("/api/messages", messageController.GetAll)
	engine.GET("/api/messages/session", messageController.GetBySessionID)
	engine.GET("/api/messages/user", messageController.GetByUserID)
	engine.GET("/api/messages/type", messageController.GetByType)
	engine.GET("/api/messages/:id", messageController.GetByID)
	engine.PUT("/api/messages/:id", messageController.Update)
	engine.DELETE("/api/messages/:id", messageController.Delete)
	engine.POST("/api/messages/:id/feedback", messageController.AddFeedback)
	engine.DELETE("/api/messages", messageController.DeleteBySessionID)
	engine.POST("/api/messages/:id/feedback", messageController.AddFeedback)

	engine.POST("/api/sessions", sessionController.Create)
	engine.GET("/api/sessions", sessionController.GetAll)
	engine.GET("/api/sessions/user", sessionController.GetByUserID)
	engine.GET("/api/sessions/status", sessionController.GetByStatus)
	engine.GET("/api/sessions/:id", sessionController.GetByID)
	engine.PUT("/api/sessions/:id", sessionController.Update)
	engine.DELETE("/api/sessions/:id", sessionController.Delete)
	engine.DELETE("/api/sessions", sessionController.DeleteByUserID)
	engine.GET("/api/sessions/:id/config", sessionController.GetConfig)
	engine.PUT("/api/sessions/:id/config", sessionController.UpdateConfig)

	return engine
}
