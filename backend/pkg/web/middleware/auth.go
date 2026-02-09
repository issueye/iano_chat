package middleware

import (
	"strings"

	"iano_chat/pkg/web"
)

// AuthConfig 认证配置
type AuthConfig struct {
	TokenLookup   string
	TokenHeader   string
	TokenQuery    string
	Validator     func(token string) (interface{}, error)
	Unauthorized  func(c *web.Context)
}

// DefaultAuthConfig 默认认证配置
var DefaultAuthConfig = AuthConfig{
	TokenLookup:  "header:Authorization",
	TokenHeader:  "Bearer",
	TokenQuery:   "token",
	Validator:    nil,
	Unauthorized: defaultUnauthorized,
}

func defaultUnauthorized(c *web.Context) {
	c.String(401, "Unauthorized")
}

// Auth 认证中间件
func Auth(validator func(token string) (interface{}, error)) web.HandlerFunc {
	config := DefaultAuthConfig
	config.Validator = validator
	return AuthWithConfig(config)
}

// AuthWithConfig 带配置的认证中间件
func AuthWithConfig(config AuthConfig) web.HandlerFunc {
	return func(c *web.Context) {
		token := ""

		// 从 Header 获取
		if strings.HasPrefix(config.TokenLookup, "header:") {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == config.TokenHeader {
					token = parts[1]
				} else if len(parts) == 1 {
					token = parts[0]
				}
			}
		}

		// 从 Query 获取
		if token == "" && strings.Contains(config.TokenLookup, "query") {
			token = c.Query(config.TokenQuery)
		}

		// 验证 Token
		if config.Validator != nil {
			user, err := config.Validator(token)
			if err != nil {
				if config.Unauthorized != nil {
					config.Unauthorized(c)
				} else {
					defaultUnauthorized(c)
				}
				c.Abort()
				return
			}
			// 将用户信息存储到上下文中
			c.Set("user", user)
		} else if token == "" {
			if config.Unauthorized != nil {
				config.Unauthorized(c)
			} else {
				defaultUnauthorized(c)
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// BearerAuth Bearer Token 认证
func BearerAuth(validator func(token string) (interface{}, error)) web.HandlerFunc {
	return Auth(validator)
}

// APIKeyAuth API Key 认证
func APIKeyAuth(key string) web.HandlerFunc {
	return func(c *web.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey != key {
			c.String(401, "Invalid API Key")
			c.Abort()
			return
		}

		c.Next()
	}
}

// BasicAuth 基础认证
func BasicAuth(accounts map[string]string) web.HandlerFunc {
	return func(c *web.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
			c.String(401, "Unauthorized")
			c.Abort()
			return
		}

		expectedPassword, exists := accounts[username]
		if !exists || expectedPassword != password {
			c.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
			c.String(401, "Unauthorized")
			c.Abort()
			return
		}

		c.Set("user", username)
		c.Next()
	}
}
