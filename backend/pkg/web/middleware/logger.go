package middleware

import (
	"fmt"
	"log"
	"os"
	"time"

	"iano_chat/pkg/web"
)

// LoggerConfig 日志配置
type LoggerConfig struct {
	Output    *log.Logger
	Formatter LogFormatter
}

// LogFormatter 日志格式化函数
type LogFormatter func(c *web.Context, latency time.Duration) string

// defaultFormatter 默认日志格式
func defaultFormatter(c *web.Context, latency time.Duration) string {
	return fmt.Sprintf("[%s] %s %s %d %v",
		time.Now().Format("2006-01-02 15:04:05"),
		c.Method,
		c.Path,
		c.GetStatus(),
		latency,
	)
}

// Logger 日志中间件
func Logger() web.HandlerFunc {
	return LoggerWithConfig(LoggerConfig{
		Output:    log.New(os.Stdout, "", 0),
		Formatter: defaultFormatter,
	})
}

// LoggerWithConfig 带配置的日志中间件
func LoggerWithConfig(config LoggerConfig) web.HandlerFunc {
	return func(c *web.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)

		if config.Formatter != nil {
			config.Output.Println(config.Formatter(c, latency))
		} else {
			config.Output.Println(defaultFormatter(c, latency))
		}
	}
}

// SimpleLogger 简单日志中间件
func SimpleLogger() web.HandlerFunc {
	return func(c *web.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		log.Printf("[%s] %s %s %v",
			c.Method,
			c.Path,
			getStatusColor(c.GetStatus()),
			latency,
		)
	}
}

// getStatusColor 获取状态码颜色（用于终端输出）
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return fmt.Sprintf("\033[32m%d\033[0m", status) // 绿色
	case status >= 300 && status < 400:
		return fmt.Sprintf("\033[33m%d\033[0m", status) // 黄色
	case status >= 400 && status < 500:
		return fmt.Sprintf("\033[31m%d\033[0m", status) // 红色
	case status >= 500:
		return fmt.Sprintf("\033[35m%d\033[0m", status) // 紫色
	default:
		return fmt.Sprintf("%d", status)
	}
}

// CustomLogger 自定义格式日志
func CustomLogger(format string) web.HandlerFunc {
	return func(c *web.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.GetStatus()

		// 简单的格式化替换
		output := format
		output = replaceAll(output, "{time}", time.Now().Format("2006-01-02 15:04:05"))
		output = replaceAll(output, "{method}", c.Method)
		output = replaceAll(output, "{path}", c.Path)
		output = replaceAll(output, "{status}", fmt.Sprintf("%d", status))
		output = replaceAll(output, "{latency}", latency.String())
		output = replaceAll(output, "{ip}", c.Request.RemoteAddr)
		output = replaceAll(output, "{user-agent}", c.GetHeader("User-Agent"))

		log.Println(output)
	}
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := 0
		for i := 0; i <= len(s)-len(old); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx == 0 && (len(s) < len(old) || s[:len(old)] != old) {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}
