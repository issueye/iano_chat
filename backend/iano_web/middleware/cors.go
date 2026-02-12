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
			if o == "*" {
				// 如果允许所有来源，且有具体的 Origin 头，则返回该 Origin
				// 这支持 Wails 等桌面应用的 wails:// 协议和开发服务器
				if origin != "" {
					allowOrigin = origin
				} else {
					allowOrigin = "*"
				}
				break
			}
			if o == origin {
				allowOrigin = o
				break
			}
		}

		// 设置 CORS 响应头 - 对所有请求都设置
		if allowOrigin != "" {
			c.SetHeader("Access-Control-Allow-Origin", allowOrigin)
		} else {
			// 如果没有匹配的 Origin，但请求带有 Origin 头，允许该 Origin
			if origin != "" {
				c.SetHeader("Access-Control-Allow-Origin", origin)
			}
		}

		if config.AllowCredentials {
			c.SetHeader("Access-Control-Allow-Credentials", "true")
		}

		if len(config.ExposeHeaders) > 0 {
			c.SetHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		// 处理预检请求 (OPTIONS) - 必须优先处理
		if c.Method == "OPTIONS" {
			// 设置允许的 HTTP 方法
			if len(config.AllowMethods) > 0 {
				c.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			} else {
				c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
			}

			// 设置允许的请求头 - 支持所有请求头
			requestHeaders := c.GetHeader("Access-Control-Request-Headers")
			if requestHeaders != "" {
				c.SetHeader("Access-Control-Allow-Headers", requestHeaders)
			} else if len(config.AllowHeaders) > 0 {
				c.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			} else {
				c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			}

			// 设置缓存时间
			if config.MaxAge > 0 {
				c.SetHeader("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			}

			// 返回 204 No Content
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
