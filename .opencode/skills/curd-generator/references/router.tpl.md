### Route Registration Template

```go
// In routes/routes.go SetupRoutes function:
{modelVar}Service := services.New{ModelName}Service(db)
{modelVar}Controller := controllers.New{ModelName}Controller({modelVar}Service)

// Routes
engine.POST("/api/{resource}s", {modelVar}Controller.Create)
engine.GET("/api/{resource}s", {modelVar}Controller.GetAll)
engine.GET("/api/{resource}s/:id", {modelVar}Controller.GetByID)
engine.PUT("/api/{resource}s/:id", {modelVar}Controller.Update)
engine.DELETE("/api/{resource}s/:id", {modelVar}Controller.Delete)
// Add query-based routes as needed
engine.GET("/api/{resource}s/user", {modelVar}Controller.GetByUserID)
```