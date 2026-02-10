package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	web "iano_web"
)

func TestBearerAuth(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		validator  func(token string) (interface{}, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Valid Bearer Token",
			authHeader: "Bearer valid-token",
			validator: func(token string) (interface{}, error) {
				if token == "valid-token" {
					return map[string]string{"id": "123", "name": "test"}, nil
				}
				return nil, http.ErrNoCookie
			},
			wantStatus: 200,
			wantBody:   "OK",
		},
		{
			name:       "Invalid Bearer Token",
			authHeader: "Bearer invalid-token",
			validator: func(token string) (interface{}, error) {
				return nil, http.ErrNoCookie
			},
			wantStatus: 401,
			wantBody:   "Unauthorized",
		},
		{
			name:       "Missing Bearer Prefix",
			authHeader: "invalid-token",
			validator: func(token string) (interface{}, error) {
				return nil, http.ErrNoCookie
			},
			wantStatus: 401,
			wantBody:   "Unauthorized",
		},
		{
			name:       "Empty Authorization",
			authHeader: "",
			validator: func(token string) (interface{}, error) {
				return nil, http.ErrNoCookie
			},
			wantStatus: 401,
			wantBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := web.New()
			engine.Use(Auth(tt.validator))
			engine.GET("/protected", func(c *web.Context) {
				c.String(200, "OK")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if w.Body.String() != tt.wantBody {
				t.Errorf("Expected body '%s', got '%s'", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestAPIKeyAuth(t *testing.T) {
	validKey := "secret-api-key"

	tests := []struct {
		name       string
		apiKey     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Valid API Key in Header",
			apiKey:     "secret-api-key",
			wantStatus: 200,
			wantBody:   "OK",
		},
		{
			name:       "Invalid API Key",
			apiKey:     "wrong-key",
			wantStatus: 401,
			wantBody:   "Invalid API Key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := web.New()
			engine.Use(APIKeyAuth(validKey))
			engine.GET("/api/data", func(c *web.Context) {
				c.String(200, "OK")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/data", nil)
			req.Header.Set("X-API-Key", tt.apiKey)
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if w.Body.String() != tt.wantBody {
				t.Errorf("Expected body '%s', got '%s'", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestAPIKeyAuthFromQuery(t *testing.T) {
	validKey := "secret-api-key"

	engine := web.New()
	engine.Use(APIKeyAuth(validKey))
	engine.GET("/api/data", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/data?api_key="+validKey, nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestBasicAuth(t *testing.T) {
	accounts := map[string]string{
		"admin": "admin123",
		"user":  "user123",
	}

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Valid Credentials",
			username:   "admin",
			password:   "admin123",
			wantStatus: 200,
			wantBody:   "Welcome admin",
		},
		{
			name:       "Invalid Password",
			username:   "admin",
			password:   "wrong-password",
			wantStatus: 401,
			wantBody:   "Unauthorized",
		},
		{
			name:       "Unknown User",
			username:   "unknown",
			password:   "password",
			wantStatus: 401,
			wantBody:   "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := web.New()
			engine.Use(BasicAuth(accounts))
			engine.GET("/admin", func(c *web.Context) {
				user, _ := c.Get("user")
				c.String(200, "Welcome %s", user)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/admin", nil)
			req.SetBasicAuth(tt.username, tt.password)
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if w.Body.String() != tt.wantBody {
				t.Errorf("Expected body '%s', got '%s'", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestBasicAuthWithoutCredentials(t *testing.T) {
	accounts := map[string]string{
		"admin": "admin123",
	}

	engine := web.New()
	engine.Use(BasicAuth(accounts))
	engine.GET("/admin", func(c *web.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != `Basic realm="Restricted"` {
		t.Errorf("Expected WWW-Authenticate header, got '%s'", wwwAuth)
	}
}

func TestAuthWithConfig(t *testing.T) {
	config := AuthConfig{
		TokenLookup: "header:Authorization",
		TokenHeader: "Bearer",
		Validator: func(token string) (interface{}, error) {
			if token == "custom-token" {
				return "custom-user", nil
			}
			return nil, http.ErrNoCookie
		},
		Unauthorized: func(c *web.Context) {
			c.String(403, "Custom Forbidden")
		},
	}

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Valid Custom Token",
			token:      "custom-token",
			wantStatus: 200,
			wantBody:   "OK",
		},
		{
			name:       "Invalid Token",
			token:      "wrong-token",
			wantStatus: 403,
			wantBody:   "Custom Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := web.New()
			engine.Use(AuthWithConfig(config))
			engine.GET("/protected", func(c *web.Context) {
				c.String(200, "OK")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			engine.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if w.Body.String() != tt.wantBody {
				t.Errorf("Expected body '%s', got '%s'", tt.wantBody, w.Body.String())
			}
		})
	}
}
