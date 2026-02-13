package routes

import (
	"net/http"
	"time"

	"iano_server/controllers"
	"iano_server/docs"
	"iano_server/pkg/config"
	"iano_server/services"
	web "iano_web"
	webMiddleware "iano_web/middleware"

	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, cfg *config.Config) *web.Engine {
	engine := web.New()
	engine.SetMode(cfg.Server.Mode)
	engine.SetReadTimeout(time.Duration(cfg.Server.ReadTimeout) * time.Second)
	engine.SetWriteTimeout(time.Duration(cfg.Server.WriteTimeout) * time.Second)
	engine.SetGracefulShutdown(true)

	engine.Use(webMiddleware.CORS())
	engine.Use(webMiddleware.Recovery())
	engine.Use(webMiddleware.Logger())

	docs.SwaggerInfo.Title = "IANO Chat API"
	docs.SwaggerInfo.Description = "IANO Chat 是一个智能对话系统，支持多 Agent、工具调用、流式响应等功能。"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:" + cfg.Server.Port
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	engine.GET("/swagger/doc.json", func(c *web.Context) {
		c.SetHeader("Content-Type", "application/json; charset=utf-8")
		c.Status(http.StatusOK)
		c.Writer.Write([]byte(docs.SwaggerInfo.ReadDoc()))
	})

	engine.GET("/swagger/*any", func(c *web.Context) {
		html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>IANO Chat API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; padding:0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/swagger/doc.json",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout"
            })
        }
    </script>
</body>
</html>`
		c.SetHeader("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		c.Writer.Write([]byte(html))
	})

	agentService := services.NewAgentService(db)
	messageService := services.NewMessageService(db)
	sessionService := services.NewSessionService(db)
	toolService := services.NewToolService(db)
	providerService := services.NewProviderService(db)
	mcpService := services.NewMCPService(db)
	agentRuntimeService := services.NewAgentRuntimeServiceWithMCP(db, agentService, providerService, toolService, mcpService)

	agentController := controllers.NewAgentController(agentService, agentRuntimeService)
	messageController := controllers.NewMessageController(messageService)
	sessionController := controllers.NewSessionController(sessionService)
	toolController := controllers.NewToolController(toolService)
	providerController := controllers.NewProviderController(providerService)
	chatController := controllers.NewChatController(agentService, providerService, messageService, agentRuntimeService)
	mcpController := controllers.NewMCPController(mcpService)
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
	engine.GET("/api/tools/:id/test", toolController.Test)

	engine.POST("/api/agents", agentController.Create)
	engine.GET("/api/agents", agentController.GetAll)
	engine.GET("/api/agents/type", agentController.GetByType)
	engine.GET("/api/agents/:id", agentController.GetByID)
	engine.PUT("/api/agents/:id", agentController.Update)
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
	engine.GET("/api/providers/default", providerController.GetDefault)
	engine.GET("/api/providers/:id", providerController.GetByID)
	engine.PUT("/api/providers/:id", providerController.Update)
	engine.DELETE("/api/providers/:id", providerController.Delete)

	engine.POST("/api/chat/stream", chatController.StreamChat)
	engine.DELETE("/api/chat/session/:session_id", chatController.ClearSession)

	engine.POST("/api/mcp/servers", mcpController.CreateServer)
	engine.GET("/api/mcp/servers", mcpController.GetAllServers)
	engine.GET("/api/mcp/servers/:id", mcpController.GetServerByID)
	engine.PUT("/api/mcp/servers/:id", mcpController.UpdateServer)
	engine.DELETE("/api/mcp/servers/:id", mcpController.DeleteServer)
	engine.POST("/api/mcp/servers/connect", mcpController.ConnectServer)
	engine.POST("/api/mcp/servers/:id/disconnect", mcpController.DisconnectServer)
	engine.GET("/api/mcp/servers/:id/tools", mcpController.GetServerTools)
	engine.GET("/api/mcp/servers/:id/list-tools", mcpController.ListTools)
	engine.GET("/api/mcp/servers/:id/ping", mcpController.Ping)
	engine.POST("/api/mcp/tools/call", mcpController.CallTool)

	return engine
}
