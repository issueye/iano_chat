package routes

import (
	"gorm.io/gorm"
	"iano_chat/controllers"
	"iano_chat/pkg/web"
	"iano_chat/pkg/web/middleware"
	"iano_chat/services"
)

func SetupRoutes(db *gorm.DB) *web.Engine {
	engine := web.New()

	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())

	agentService := services.NewAgentService(db)
	agentController := controllers.NewAgentController(agentService)
	baseController := &controllers.BaseController{}

	engine.GET("/health", func(c *web.Context) {
		baseController.HealthCheck(c)
	})

	engine.GET("/api/users", func(c *web.Context) {
		c.JSON(200, map[string]string{"message": "user list"})
	})

	engine.POST("/api/users", func(c *web.Context) {
		c.JSON(201, map[string]string{"message": "user created"})
	})

	engine.GET("/api/users/{id}", func(c *web.Context) {
		id := c.Param("id")
		c.JSON(200, map[string]string{"id": id, "message": "user detail"})
	})

	engine.PUT("/api/users/{id}", func(c *web.Context) {
		id := c.Param("id")
		c.JSON(200, map[string]string{"id": id, "message": "user updated"})
	})

	engine.DELETE("/api/users/{id}", func(c *web.Context) {
		id := c.Param("id")
		c.JSON(200, map[string]string{"id": id, "message": "user deleted"})
	})

	engine.POST("/api/agents", func(c *web.Context) {
		agentController.Create(c)
	})

	engine.GET("/api/agents", func(c *web.Context) {
		agentController.GetAll(c)
	})

	engine.GET("/api/agents/type", func(c *web.Context) {
		agentController.GetByType(c)
	})

	engine.GET("/api/agents/{id}", func(c *web.Context) {
		agentController.GetByID(c)
	})

	engine.PUT("/api/agents/{id}", func(c *web.Context) {
		agentController.Update(c)
	})

	engine.DELETE("/api/agents/{id}", func(c *web.Context) {
		agentController.Delete(c)
	})

	return engine
}
