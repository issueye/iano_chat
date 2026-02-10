package main

import (
	"fmt"
	web "iano_web"
	"log"
	"net/http"
	"time"
)

func main() {
	engine := web.New()

	engine.Use(func(c *web.Context) {
		start := time.Now()
		fmt.Printf("[%s] %s\n", c.Method, c.Path)
		c.Next()
		fmt.Printf("Request completed in %v\n", time.Since(start))
	})

	engine.Use(func(c *web.Context) {
		c.SetHeader("X-Powered-By", "iano-chat-web")
		c.Next()
	})

	engine.GET("/", func(c *web.Context) {
		c.HTML(200, `
			<!DOCTYPE html>
			<html>
			<head><title>Welcome</title></head>
			<body>
				<h1>Welcome to iano-chat-web</h1>
				<p>A lightweight Express-style HTTP framework for Go</p>
				<ul>
					<li><a href="/hello?name=World">Hello World</a></li>
					<li><a href="/users/123">User 123</a></li>
					<li><a href="/api/v1/data">API Data</a></li>
					<li><a href="/admin/dashboard">Admin Dashboard</a></li>
				</ul>
			</body>
			</html>
		`)
	})

	engine.GET("/hello", func(c *web.Context) {
		name := c.Query("name")
		if name == "" {
			name = "World"
		}
		c.JSON(200, map[string]interface{}{
			"message": fmt.Sprintf("Hello, %s!", name),
			"method":  c.Method,
			"path":    c.Path,
		})
	})

	engine.GET("/users/:id", func(c *web.Context) {
		userID := c.Param("id")
		c.JSON(200, map[string]interface{}{
			"id":     userID,
			"name":   fmt.Sprintf("User %s", userID),
			"email":  fmt.Sprintf("user%s@example.com", userID),
			"active": true,
		})
	})

	engine.GET("/users/:id/posts/:postId", func(c *web.Context) {
		userID := c.Param("id")
		postID := c.Param("postId")
		c.JSON(200, map[string]interface{}{
			"userId":  userID,
			"postId":  postID,
			"title":   fmt.Sprintf("Post %s by User %s", postID, userID),
			"content": "This is a sample post content.",
		})
	})

	engine.POST("/users", func(c *web.Context) {
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := c.Bind(&user); err != nil {
			c.String(400, "Invalid request body: "+err.Error())
			return
		}
		c.JSON(201, map[string]interface{}{
			"id":     1,
			"name":   user.Name,
			"email":  user.Email,
			"status": "created",
		})
	})

	engine.PUT("/users/:id", func(c *web.Context) {
		userID := c.Param("id")
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := c.Bind(&user); err != nil {
			c.String(400, "Invalid request body")
			return
		}
		c.JSON(200, map[string]interface{}{
			"id":     userID,
			"name":   user.Name,
			"email":  user.Email,
			"status": "updated",
		})
	})

	engine.DELETE("/users/:id", func(c *web.Context) {
		userID := c.Param("id")
		c.JSON(200, map[string]interface{}{
			"id":     userID,
			"status": "deleted",
		})
	})

	engine.GET("/api/*filepath", func(c *web.Context) {
		filepath := c.Param("filepath")
		c.JSON(200, map[string]interface{}{
			"filepath": filepath,
			"message":  "Wildcard route matched",
		})
	})

	engine.Group("/api/v1", func(e *web.Engine) {
		e.GET("/data", func(c *web.Context) {
			c.JSON(200, map[string]interface{}{
				"version": "v1",
				"data":    []string{"item1", "item2", "item3"},
			})
		})
		e.GET("/info", func(c *web.Context) {
			c.JSON(200, map[string]interface{}{
				"version": "v1",
				"info":    "API v1 information",
			})
		})
	})

	engine.Group("/admin", func(e *web.Engine) {
		e.Use(func(c *web.Context) {
			auth := c.GetHeader("Authorization")
			if auth == "" {
				c.String(401, "Unauthorized")
				c.Abort()
				return
			}
			c.Next()
		})
		e.GET("/dashboard", func(c *web.Context) {
			c.JSON(200, map[string]interface{}{
				"page": "dashboard",
				"user": "admin",
			})
		})
		e.GET("/settings", func(c *web.Context) {
			c.JSON(200, map[string]interface{}{
				"page": "settings",
				"user": "admin",
			})
		})
	})

	engine.GET("/redirect", func(c *web.Context) {
		c.Redirect(302, "/")
	})

	engine.GET("/cookie", func(c *web.Context) {
		cookie, err := c.Cookie("session")
		if err != nil {
			c.String(200, "No session cookie found")
		} else {
			c.String(200, "Session cookie: "+cookie.Value)
		}
	})

	engine.GET("/set-cookie", func(c *web.Context) {
		c.SetCookie(&http.Cookie{
			Name:     "session",
			Value:    "abc123",
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
		})
		c.String(200, "Cookie set")
	})

	engine.Any("/any-method", func(c *web.Context) {
		c.JSON(200, map[string]interface{}{
			"message": "This route accepts any HTTP method",
			"method":  c.Method,
		})
	})

	fmt.Println("===========================================")
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("===========================================")
	fmt.Println("\nAvailable routes:")
	fmt.Println("  GET  /")
	fmt.Println("  GET  /hello?name=World")
	fmt.Println("  GET  /users/:id")
	fmt.Println("  GET  /users/:id/posts/:postId")
	fmt.Println("  POST /users")
	fmt.Println("  PUT  /users/:id")
	fmt.Println("  DELETE /users/:id")
	fmt.Println("  GET  /api/*filepath")
	fmt.Println("  GET  /api/v1/data")
	fmt.Println("  GET  /admin/dashboard (requires Authorization header)")
	fmt.Println("  GET  /redirect")
	fmt.Println("  GET  /cookie")
	fmt.Println("  GET  /set-cookie")
	fmt.Println("  ANY  /any-method")
	fmt.Println("===========================================\n")

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
