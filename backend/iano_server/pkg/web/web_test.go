package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEngineBasic(t *testing.T) {
	engine := New()

	engine.GET("/ping", func(c *Context) {
		c.String(200, "pong")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "pong" {
		t.Errorf("Expected body 'pong', got '%s'", w.Body.String())
	}
}

func TestEngineMethods(t *testing.T) {
	engine := New()

	engine.GET("/get", func(c *Context) { c.String(200, "GET") })
	engine.POST("/post", func(c *Context) { c.String(200, "POST") })
	engine.PUT("/put", func(c *Context) { c.String(200, "PUT") })
	engine.DELETE("/delete", func(c *Context) { c.String(200, "DELETE") })
	engine.PATCH("/patch", func(c *Context) { c.String(200, "PATCH") })

	tests := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/get", "GET"},
		{"POST", "/post", "POST"},
		{"PUT", "/put", "PUT"},
		{"DELETE", "/delete", "DELETE"},
		{"PATCH", "/patch", "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			engine.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
			if w.Body.String() != tt.want {
				t.Errorf("Expected body '%s', got '%s'", tt.want, w.Body.String())
			}
		})
	}
}

func TestEngineGroup(t *testing.T) {
	engine := New()

	engine.Group("/v1", func(v1 *Engine) {
		v1.GET("/users", func(c *Context) {
			c.String(200, "v1 users")
		})
		v1.GET("/posts", func(c *Context) {
			c.String(200, "v1 posts")
		})
	})

	engine.Group("/v2", func(v2 *Engine) {
		v2.GET("/users", func(c *Context) {
			c.String(200, "v2 users")
		})
	})

	tests := []struct {
		path string
		want string
	}{
		{"/v1/users", "v1 users"},
		{"/v1/posts", "v1 posts"},
		{"/v2/users", "v2 users"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			engine.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
			if w.Body.String() != tt.want {
				t.Errorf("Expected body '%s', got '%s'", tt.want, w.Body.String())
			}
		})
	}
}

func TestEngineMiddleware(t *testing.T) {
	engine := New()

	engine.Use(func(c *Context) {
		c.SetHeader("X-Middleware", "applied")
		c.Next()
	})

	engine.GET("/test", func(c *Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Header().Get("X-Middleware") != "applied" {
		t.Error("Middleware was not applied")
	}
}

func TestEngineGroupMiddleware(t *testing.T) {
	engine := New()

	engine.Use(func(c *Context) {
		c.SetHeader("X-Global", "global")
		c.Next()
	})

	engine.Group("/api", func(api *Engine) {
		api.Use(func(c *Context) {
			c.SetHeader("X-Group", "group")
			c.Next()
		})
		api.GET("/test", func(c *Context) {
			c.String(200, "ok")
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	engine.ServeHTTP(w, req)

	if w.Header().Get("X-Global") != "global" {
		t.Error("Global middleware was not applied")
	}
	if w.Header().Get("X-Group") != "group" {
		t.Error("Group middleware was not applied")
	}
}

func TestContextJSON(t *testing.T) {
	engine := New()

	type response struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}

	engine.GET("/json", func(c *Context) {
		c.JSON(200, response{Message: "hello", Status: 200})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/json", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type to contain 'application/json', got '%s'", contentType)
	}

	var resp response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}
	if resp.Message != "hello" || resp.Status != 200 {
		t.Errorf("Unexpected response: %+v", resp)
	}
}

func TestContextBind(t *testing.T) {
	engine := New()

	type user struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	engine.POST("/user", func(c *Context) {
		var u user
		if err := c.Bind(&u); err != nil {
			c.String(400, "bad request")
			return
		}
		c.JSON(200, u)
	})

	body := `{"name":"test","email":"test@example.com"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp user
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}
	if resp.Name != "test" || resp.Email != "test@example.com" {
		t.Errorf("Unexpected response: %+v", resp)
	}
}

func TestContextQuery(t *testing.T) {
	engine := New()

	engine.GET("/query", func(c *Context) {
		name := c.Query("name")
		age := c.DefaultQuery("age", "18")
		c.String(200, "name=%s,age=%s", name, age)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/query?name=test&age=25", nil)
	engine.ServeHTTP(w, req)

	if w.Body.String() != "name=test,age=25" {
		t.Errorf("Unexpected response: %s", w.Body.String())
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/query?name=test", nil)
	engine.ServeHTTP(w2, req2)

	if w2.Body.String() != "name=test,age=18" {
		t.Errorf("Unexpected default response: %s", w2.Body.String())
	}
}

func TestContextParam(t *testing.T) {
	engine := New()

	engine.GET("/user/:id", func(c *Context) {
		id := c.Param("id")
		c.String(200, "user id: %s", id)
	})

	engine.GET("/user/:id/post/:pid", func(c *Context) {
		uid := c.Param("id")
		pid := c.Param("pid")
		c.String(200, "user=%s,post=%s", uid, pid)
	})

	tests := []struct {
		path string
		want string
	}{
		{"/user/123", "user id: 123"},
		{"/user/abc/post/xyz", "user=abc,post=xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			engine.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
			if w.Body.String() != tt.want {
				t.Errorf("Expected body '%s', got '%s'", tt.want, w.Body.String())
			}
		})
	}
}

func TestContextAbort(t *testing.T) {
	engine := New()

	engine.Use(func(c *Context) {
		c.Abort()
		c.String(401, "unauthorized")
	})

	engine.GET("/protected", func(c *Context) {
		c.String(200, "should not reach here")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
	if w.Body.String() != "unauthorized" {
		t.Errorf("Expected body 'unauthorized', got '%s'", w.Body.String())
	}
}

func TestNotFound(t *testing.T) {
	engine := New()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notexist", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestTrieRouter(t *testing.T) {
	tr := NewTrieRouter()

	route1 := &Route{method: "GET", pattern: "/user/:id"}
	route2 := &Route{method: "GET", pattern: "/user/:id/post/:pid"}
	route3 := &Route{method: "POST", pattern: "/user"}

	tr.insert("/user/:id", route1)
	tr.insert("/user/:id/post/:pid", route2)
	tr.insert("/user", route3)

	tests := []struct {
		method string
		path   string
		route  *Route
		params map[string]string
	}{
		{"GET", "/user/123", route1, map[string]string{"id": "123"}},
		{"GET", "/user/abc/post/xyz", route2, map[string]string{"id": "abc", "pid": "xyz"}},
		{"POST", "/user", route3, map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			route, params := tr.search(tt.method, tt.path)
			if route != tt.route {
				t.Error("Route mismatch")
			}
			for k, v := range tt.params {
				if params[k] != v {
					t.Errorf("Expected param %s=%s, got %s", k, v, params[k])
				}
			}
		})
	}
}

func BenchmarkEngine(b *testing.B) {
	engine := New()

	engine.GET("/user/:id", func(c *Context) {
		c.String(200, "%s", c.Param("id"))
	})

	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/user/123", nil)
		engine.ServeHTTP(w, req)
	}
}

func BenchmarkEngineWithMiddleware(b *testing.B) {
	engine := New()

	engine.Use(func(c *Context) {
		c.Next()
	})

	engine.GET("/user/:id", func(c *Context) {
		c.String(200, "%s", c.Param("id"))
	})

	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/user/123", nil)
		engine.ServeHTTP(w, req)
	}
}
