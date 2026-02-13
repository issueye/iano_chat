package routes

import (
	"net/http"
	"time"

	"iano_server/container"
	"iano_server/docs"
	web "iano_web"
	webMiddleware "iano_web/middleware"
)

func SetupRoutes(cnr *container.Container) *web.Engine {
	engine := web.New()
	engine.SetMode(cnr.GetConfig().Server.Mode)
	engine.SetReadTimeout(time.Duration(cnr.GetConfig().Server.ReadTimeout) * time.Second)
	engine.SetWriteTimeout(time.Duration(cnr.GetConfig().Server.WriteTimeout) * time.Second)
	engine.SetGracefulShutdown(true)

	engine.Use(webMiddleware.CORS())
	engine.Use(webMiddleware.Recovery())
	engine.Use(webMiddleware.Logger())

	docs.SwaggerInfo.Title = "IANO Chat API"
	docs.SwaggerInfo.Description = "IANO Chat 是一个智能对话系统，支持多 Agent、工具调用、流式响应等功能。"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:" + cnr.GetConfig().Server.Port
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

	engine.GET("/health", func(c *web.Context) {
		cnr.BaseController.HealthCheck(c)
	})

	engine.POST("/api/tools", cnr.ToolController.Create)
	engine.GET("/api/tools", cnr.ToolController.GetAll)
	engine.GET("/api/tools/type", cnr.ToolController.GetByType)
	engine.GET("/api/tools/status", cnr.ToolController.GetByStatus)
	engine.GET("/api/tools/:id", cnr.ToolController.GetByID)
	engine.PUT("/api/tools/:id", cnr.ToolController.Update)
	engine.PUT("/api/tools/:id/config", cnr.ToolController.UpdateConfig)
	engine.DELETE("/api/tools/:id", cnr.ToolController.Delete)
	engine.GET("/api/tools/:id/test", cnr.ToolController.Test)

	engine.POST("/api/agents", cnr.AgentController.Create)
	engine.GET("/api/agents", cnr.AgentController.GetAll)
	engine.GET("/api/agents/type", cnr.AgentController.GetByType)
	engine.GET("/api/agents/:id", cnr.AgentController.GetByID)
	engine.PUT("/api/agents/:id", cnr.AgentController.Update)
	engine.DELETE("/api/agents/:id", cnr.AgentController.Delete)
	engine.POST("/api/agents/:id/tools", cnr.AgentController.AddTool)
	engine.DELETE("/api/agents/:id/tools/:tool_name", cnr.AgentController.RemoveTool)

	engine.POST("/api/messages", cnr.MessageController.Create)
	engine.GET("/api/messages", cnr.MessageController.GetAll)
	engine.GET("/api/messages/session", cnr.MessageController.GetBySessionID)
	engine.GET("/api/messages/type", cnr.MessageController.GetByType)
	engine.GET("/api/messages/:id", cnr.MessageController.GetByID)
	engine.PUT("/api/messages/:id", cnr.MessageController.Update)
	engine.DELETE("/api/messages/:id", cnr.MessageController.Delete)
	engine.POST("/api/messages/:id/feedback", cnr.MessageController.AddFeedback)
	engine.DELETE("/api/messages", cnr.MessageController.DeleteBySessionID)

	engine.POST("/api/sessions", cnr.SessionController.Create)
	engine.GET("/api/sessions", cnr.SessionController.GetAll)
	engine.GET("/api/sessions/status", cnr.SessionController.GetByStatus)
	engine.GET("/api/sessions/:id", cnr.SessionController.GetByID)
	engine.PUT("/api/sessions/:id", cnr.SessionController.Update)
	engine.DELETE("/api/sessions/:id", cnr.SessionController.Delete)
	engine.GET("/api/sessions/:id/config", cnr.SessionController.GetConfig)
	engine.PUT("/api/sessions/:id/config", cnr.SessionController.UpdateConfig)

	engine.POST("/api/providers", cnr.ProviderController.Create)
	engine.GET("/api/providers", cnr.ProviderController.GetAll)
	engine.GET("/api/providers/default", cnr.ProviderController.GetDefault)
	engine.GET("/api/providers/:id", cnr.ProviderController.GetByID)
	engine.PUT("/api/providers/:id", cnr.ProviderController.Update)
	engine.DELETE("/api/providers/:id", cnr.ProviderController.Delete)

	engine.POST("/api/chat/stream", cnr.ChatController.StreamChat)
	engine.DELETE("/api/chat/session/:session_id", cnr.ChatController.ClearSession)

	engine.POST("/api/mcp/servers", cnr.MCPController.CreateServer)
	engine.GET("/api/mcp/servers", cnr.MCPController.GetAllServers)
	engine.GET("/api/mcp/servers/:id", cnr.MCPController.GetServerByID)
	engine.PUT("/api/mcp/servers/:id", cnr.MCPController.UpdateServer)
	engine.DELETE("/api/mcp/servers/:id", cnr.MCPController.DeleteServer)
	engine.POST("/api/mcp/servers/connect", cnr.MCPController.ConnectServer)
	engine.POST("/api/mcp/servers/:id/disconnect", cnr.MCPController.DisconnectServer)
	engine.GET("/api/mcp/servers/:id/tools", cnr.MCPController.GetServerTools)
	engine.GET("/api/mcp/servers/:id/list-tools", cnr.MCPController.ListTools)
	engine.GET("/api/mcp/servers/:id/ping", cnr.MCPController.Ping)
	engine.POST("/api/mcp/tools/call", cnr.MCPController.CallTool)

	return engine
}
