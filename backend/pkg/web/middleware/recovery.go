package middleware

import (
	"log"
	"log/slog"
	"runtime/debug"

	"iano_chat/pkg/web"
)

// RecoveryConfig 恢复配置
type RecoveryConfig struct {
	Handler func(c *web.Context, err interface{})
}

// defaultRecoveryHandler 默认恢复处理
func defaultRecoveryHandler(c *web.Context, err interface{}) {
	log.Printf("[PANIC] %v\n%s", err, debug.Stack())
	c.String(500, "Internal Server Error")
}

// Recovery 恢复中间件（捕获 panic）
func Recovery() web.HandlerFunc {
	return RecoveryWithConfig(RecoveryConfig{
		Handler: defaultRecoveryHandler,
	})
}

// RecoveryWithConfig 带配置的恢复中间件
func RecoveryWithConfig(config RecoveryConfig) web.HandlerFunc {
	return func(c *web.Context) {
		defer func() {
			if err := recover(); err != nil {
				if config.Handler != nil {
					config.Handler(c, err)
				} else {
					defaultRecoveryHandler(c, err)
				}
			}
		}()
		c.Next()
	}
}

// RecoveryWithLog 带日志的恢复中间件
func RecoveryWithLog(logger *slog.Logger) web.HandlerFunc {
	return func(c *web.Context) {
		defer func() {
			if err := recover(); err != nil {
				// logger.Printf("[PANIC] %s %s - %v\n%s",
				// 	c.Method,
				// 	c.Path,
				// 	err,
				// 	debug.Stack(),
				// )
				logger.ErrorContext(c.Request.Context(), "[PANIC]",
					slog.String("method", c.Method),
					slog.String("path", c.Path),
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)
				c.String(500, "Internal Server Error: %v", err)
			}
		}()
		c.Next()
	}
}
