package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	web "iano_web"
)

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		method         string
		wantStatus     int
		wantCORSOrigin string
	}{
		{
			name:           "Simple Request with Origin",
			origin:         "http://localhost:3000",
			method:         "GET",
			wantStatus:     200,
			wantCORSOrigin: "*",
		},
		{
			name:           "Simple Request without Origin",
			origin:         "",
			method:         "GET",
			wantStatus:     200,
			wantCORSOrigin: "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := web.New()
			engine.Use(CORS())
			engine.GET("/test", func(c *web.Context) {
				c.String(200, "OK")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if corsOrigin != tt.wantCORSOrigin {
				t.Errorf("Expected CORS origin '%s', got '%s'", tt.wantCORSOrigin, corsOrigin)
			}
		})
	}
}

func TestCORSWithSpecificOrigin(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	tests := []struct {
		name           string
		origin         string
		wantCORSEmpty  bool
		wantCORSOrigin string
	}{
		{
			name:           "Allowed Origin",
			origin:         "http://localhost:3000",
			wantCORSEmpty:  false,
			wantCORSOrigin: "http://localhost:3000",
		},
		{
			name:          "Disallowed Origin",
			origin:        "http://evil.com",
			wantCORSEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := web.New()
			engine.Use(CORSWithConfig(config))
			engine.GET("/test", func(c *web.Context) {
				c.String(200, "OK")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)
			engine.ServeHTTP(w, req)

			corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if tt.wantCORSEmpty && corsOrigin != "" {
				t.Errorf("Expected empty CORS origin, got '%s'", corsOrigin)
			}
			if !tt.wantCORSEmpty && corsOrigin != tt.wantCORSOrigin {
				t.Errorf("Expected CORS origin '%s', got '%s'", tt.wantCORSOrigin, corsOrigin)
			}
		})
	}
}

func TestCORSWithWildcard(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST"},
	}

	engine := web.New()
	engine.Use(CORSWithConfig(config))
	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	engine.ServeHTTP(w, req)

	corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if corsOrigin != "*" {
		t.Errorf("Expected CORS origin '*', got '%s'", corsOrigin)
	}
}

func TestCORSPreFlight(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Custom-Header"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Total-Count"},
		MaxAge:           86400,
	}

	engine := web.New()
	engine.Use(CORSWithConfig(config))
	// 使用 Any 方法支持所有 HTTP 方法，包括 OPTIONS
	engine.Any("/test", func(c *web.Context) {
		if c.Method == "OPTIONS" {
			c.Status(204)
			return
		}
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	engine.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("Expected status 204 for preflight, got %d", w.Code)
	}

	// 检查预检响应头
	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods != "GET, POST, PUT, DELETE" {
		t.Errorf("Expected Allow-Methods 'GET, POST, PUT, DELETE', got '%s'", allowMethods)
	}

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders != "Origin, Content-Type, Authorization, X-Custom-Header" {
		t.Errorf("Expected Allow-Headers, got '%s'", allowHeaders)
	}

	allowCredentials := w.Header().Get("Access-Control-Allow-Credentials")
	if allowCredentials != "true" {
		t.Errorf("Expected Allow-Credentials 'true', got '%s'", allowCredentials)
	}

	maxAge := w.Header().Get("Access-Control-Max-Age")
	if maxAge != "86400" {
		t.Errorf("Expected Max-Age '86400', got '%s'", maxAge)
	}
}

func TestAllowAllCORS(t *testing.T) {
	engine := web.New()
	engine.Use(AllowAllCORS())
	engine.GET("/test", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if corsOrigin != "*" {
		t.Errorf("Expected CORS origin '*', got '%s'", corsOrigin)
	}

	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods != "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS" {
		t.Errorf("Expected specific Allow-Methods, got '%s'", allowMethods)
	}
}

func TestAllowAllCORSPreFlight(t *testing.T) {
	engine := web.New()
	engine.Use(AllowAllCORS())
	// 使用 Any 方法支持所有 HTTP 方法
	engine.Any("/test", func(c *web.Context) {
		if c.Method == "OPTIONS" {
			c.Status(204)
			return
		}
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	engine.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("Expected status 204 for preflight, got %d", w.Code)
	}
}
