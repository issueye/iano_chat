package routes

import (
	"context"
	"log/slog"

	"iano_server/controllers"
	"iano_server/services"
	web "iano_web"
	webMiddleware "iano_web/middleware"

	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB) *web.Engine {
	engine := web.New()
	engine.SetMode("debug")

	engine.Use(webMiddleware.Recovery())
	engine.Use(webMiddleware.Logger())
	engine.Use(webMiddleware.CORS())

	agentService := services.NewAgentService(db)
	messageService := services.NewMessageService(db)
	sessionService := services.NewSessionService(db)
	toolService := services.NewToolService(db)
	providerService := services.NewProviderService(db)
	chatService := services.NewChatService(db, nil, nil, nil)

	agentManagerService := services.NewAgentManagerService(
		db,
		agentService,
		providerService,
		toolService,
	)

	if err := agentManagerService.Initialize(context.Background()); err != nil {
		slog.Error("Failed to initialize agent manager", "error", err)
	}

	agentController := controllers.NewAgentController(agentService, agentManagerService)
	messageController := controllers.NewMessageController(messageService)
	sessionController := controllers.NewSessionController(sessionService)
	toolController := controllers.NewToolController(toolService, agentManagerService)
	providerController := controllers.NewProviderController(providerService)
	chatController := controllers.NewChatController(chatService, agentService, providerService)
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
	engine.POST("/api/tools/:id/register", toolController.RegisterToAgent)
	engine.GET("/api/tools/:id/test", toolController.Test)

	engine.POST("/api/agents", agentController.Create)
	engine.GET("/api/agents", agentController.GetAll)
	engine.GET("/api/agents/type", agentController.GetByType)
	engine.GET("/api/agents/instances", agentController.ListInstances)
	engine.GET("/api/agents/stats", agentController.GetStats)
	engine.GET("/api/agents/:id", agentController.GetByID)
	engine.GET("/api/agents/:id/info", agentController.GetInstanceInfo)
	engine.PUT("/api/agents/:id", agentController.Update)
	engine.POST("/api/agents/:id/reload", agentController.Reload)
	engine.DELETE("/api/agents/:id", agentController.Delete)
	engine.POST("/api/agents/:id/tools", agentController.AddTool)
	engine.DELETE("/api/agents/:id/tools/:tool_name", agentController.RemoveTool)

	engine.POST("/api/messages", messageController.Create)
	engine.GET("/api/messages", messageController.GetAll)
	engine.GET("/api/messages/session", messageController.GetBySessionID)
	engine.GET("/api/messages/type", messageController.GetByType)
	engine.GET("/api/messages/:id", messageController.GetByID)
	engine.PUT("/api/messages/:id", messageController.Update)
	engine.DELETE("/api/messages/:id", messageController.Delete)
	engine.POST("/api/messages/:id/feedback", messageController.AddFeedback)
	engine.DELETE("/api/messages", messageController.DeleteBySessionID)

	engine.POST("/api/sessions", sessionController.Create)
	engine.GET("/api/sessions", sessionController.GetAll)
	engine.GET("/api/sessions/status", sessionController.GetByStatus)
	engine.GET("/api/sessions/:id", sessionController.GetByID)
	engine.PUT("/api/sessions/:id", sessionController.Update)
	engine.DELETE("/api/sessions/:id", sessionController.Delete)
	engine.GET("/api/sessions/:id/config", sessionController.GetConfig)
	engine.PUT("/api/sessions/:id/config", sessionController.UpdateConfig)

	engine.POST("/api/providers", providerController.Create)
	engine.GET("/api/providers", providerController.GetAll)
	engine.GET("/api/providers/:id", providerController.GetByID)
	engine.PUT("/api/providers/:id", providerController.Update)
	engine.DELETE("/api/providers/:id", providerController.Delete)

	engine.POST("/api/chat", chatController.Chat)
	engine.POST("/api/chat/stream", chatController.StreamChat)
	engine.DELETE("/api/chat/session/:session_id", chatController.ClearSession)
	engine.GET("/api/chat/conversation", chatController.GetConversationInfo)
	engine.GET("/api/chat/pool-stats", chatController.GetPoolStats)

	return engine
}
