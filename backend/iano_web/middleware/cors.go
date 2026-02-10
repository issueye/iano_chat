package middleware

import (
	"net/http"
	"strconv"
	"strings"

	web "iano_web"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// DefaultCORSConfig 默认 CORS 配置
var DefaultCORSConfig = CORSConfig{
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Authorization"},
	AllowCredentials: false,
	MaxAge:           86400,
}

// CORS 跨域中间件
func CORS() web.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig)
}

// CORSWithConfig 带配置的跨域中间件
func CORSWithConfig(config CORSConfig) web.HandlerFunc {
	return func(c *web.Context) {
		origin := c.GetHeader("Origin")

		// 检查允许的 Origin
		allowOrigin := ""
		for _, o := range config.AllowOrigins {
			if o == "*" || o == origin {
				allowOrigin = o
				break
			}
		}

		// 设置 CORS 响应头
		if allowOrigin != "" {
			c.SetHeader("Access-Control-Allow-Origin", allowOrigin)
		}

		if config.AllowCredentials {
			c.SetHeader("Access-Control-Allow-Credentials", "true")
		}

		if len(config.ExposeHeaders) > 0 {
			c.SetHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// 处理预检请求
		if c.Method == "OPTIONS" {
			if len(config.AllowMethods) > 0 {
				c.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			}

			if len(config.AllowHeaders) > 0 {
				c.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			}

			if config.MaxAge > 0 {
				c.SetHeader("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			c.Status(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AllowAllCORS 允许所有跨域请求
func AllowAllCORS() web.HandlerFunc {
	return func(c *web.Context) {
		c.SetHeader("Access-Control-Allow-Origin", "*")
		c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Accept, Authorization")
		c.SetHeader("Access-Control-Max-Age", "86400")

		if c.Method == "OPTIONS" {
			c.Status(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
