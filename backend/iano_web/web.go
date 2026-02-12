package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Engine struct {
	router       *Router
	prefix       string
	middleware   []MiddlewareFunc
	engine       *http.Server
	graceful     bool
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	mode         string // debug or release
}

func New() *Engine {
	return &Engine{
		router:      NewRouter(),
		ReadTimeout: 30 * time.Second,
	}
}

func (e *Engine) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

func (e *Engine) combineHandlers(handlers []HandlerFunc) []HandlerFunc {
	allHandlers := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	allHandlers = append(allHandlers, e.middleware...)
	allHandlers = append(allHandlers, handlers...)
	return allHandlers
}

func (e *Engine) addRoute(method, pattern string, handlers ...HandlerFunc) {
	fullPattern := e.prefix + pattern
	e.router.addRoute(method, fullPattern, e.combineHandlers(handlers), e.router.extractParamKeys(fullPattern))

	// 自动为每个路由添加 OPTIONS 支持，用于处理 CORS 预检请求
	// 使用 CORS 中间件处理 OPTIONS 请求
	if method != "OPTIONS" {
		optionsHandlers := []HandlerFunc{func(c *Context) {
			// OPTIONS 请求由 CORS 中间件处理
			// 如果执行到这里，说明 CORS 中间件没有拦截请求
			c.Status(http.StatusNoContent)
		}}
		e.router.addRoute("OPTIONS", fullPattern, e.combineHandlers(optionsHandlers), e.router.extractParamKeys(fullPattern))
	}
}

func (e *Engine) GET(pattern string, handlers ...HandlerFunc) {
	e.addRoute("GET", pattern, handlers...)
}

func (e *Engine) POST(pattern string, handlers ...HandlerFunc) {
	e.addRoute("POST", pattern, handlers...)
}

func (e *Engine) PUT(pattern string, handlers ...HandlerFunc) {
	e.addRoute("PUT", pattern, handlers...)
}

func (e *Engine) DELETE(pattern string, handlers ...HandlerFunc) {
	e.addRoute("DELETE", pattern, handlers...)
}

func (e *Engine) PATCH(pattern string, handlers ...HandlerFunc) {
	e.addRoute("PATCH", pattern, handlers...)
}

func (e *Engine) OPTIONS(pattern string, handlers ...HandlerFunc) {
	e.addRoute("OPTIONS", pattern, handlers...)
}

func (e *Engine) HEAD(pattern string, handlers ...HandlerFunc) {
	e.addRoute("HEAD", pattern, handlers...)
}

func (e *Engine) Any(pattern string, handlers ...HandlerFunc) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, method := range methods {
		switch method {
		case "GET":
			e.GET(pattern, handlers...)
		case "POST":
			e.POST(pattern, handlers...)
		case "PUT":
			e.PUT(pattern, handlers...)
		case "DELETE":
			e.DELETE(pattern, handlers...)
		case "PATCH":
			e.PATCH(pattern, handlers...)
		case "OPTIONS":
			e.OPTIONS(pattern, handlers...)
		case "HEAD":
			e.HEAD(pattern, handlers...)
		}
	}
}

func (e *Engine) Static(pattern, root string) {
	e.GET(pattern+"/*filepath", func(c *Context) {
		file := c.Param("filepath")
		file = root + "/" + file
		c.SetHeader("Content-Type", "application/octet-stream")
		http.ServeFile(c.Writer, c.Request, file)
	})
}

func (e *Engine) Group(prefix string, fn func(*Engine)) {
	groupEngine := &Engine{
		router:     e.router,
		prefix:     e.prefix + prefix,
		middleware: append([]MiddlewareFunc{}, e.middleware...),
	}
	fn(groupEngine)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := acquireContext(w, r)
	defer releaseContext(c)
	e.router.handleRequest(c)
}

func (e *Engine) Run(addr string) error {
	e.engine = &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
	}

	if e.graceful {
		return e.runWithGracefulShutdown(addr)
	}

	if e.mode == "debug" {
		e.PrintRoutes()
	}

	fmt.Printf("服务 %s 启动中...\n", addr)
	return e.engine.ListenAndServe()
}

func (e *Engine) runWithGracefulShutdown(addr string) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if e.mode == "debug" {
		e.PrintRoutes()
	}

	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("服务 %s 启动中...\n", addr)
		serverErr <- e.engine.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		fmt.Printf("\nReceived signal: %v\n", sig)
		fmt.Printf("服务 %s 关闭中...\n", addr)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := e.engine.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		fmt.Printf("服务 %s 已正常关闭\n", addr)
		return nil
	}
}

func (e *Engine) RunTLS(addr, certFile, keyFile string) error {
	e.engine = &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
	}

	if e.graceful {
		return e.runTLSWithGracefulShutdown(addr, certFile, keyFile)
	}

	if e.mode == "debug" {
		e.PrintRoutes()
	}

	fmt.Printf("Server starting on %s with TLS\n", addr)
	return e.engine.ListenAndServeTLS(certFile, keyFile)
}

func (e *Engine) runTLSWithGracefulShutdown(addr, certFile, keyFile string) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if e.mode == "debug" {
		e.PrintRoutes()
	}

	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("服务 %s 启动中...\n", addr)
		serverErr <- e.engine.ListenAndServeTLS(certFile, keyFile)
	}()

	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		fmt.Printf("\nReceived signal: %v\n", sig)
		fmt.Printf("服务 %s 关闭中...\n", addr)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := e.engine.Shutdown(ctx); err != nil {
			return fmt.Errorf("服务 %s 关闭失败: %w", addr, err)
		}

		fmt.Printf("服务 %s 已正常关闭\n", addr)
		return nil
	}
}

func (e *Engine) Shutdown(timeout time.Duration) error {
	if e.engine != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return e.engine.Shutdown(ctx)
	}
	return nil
}

func (e *Engine) SetReadTimeout(timeout time.Duration) {
	e.ReadTimeout = timeout
}

func (e *Engine) SetWriteTimeout(timeout time.Duration) {
	e.WriteTimeout = timeout
}

func (e *Engine) SetGracefulShutdown(enable bool) {
	e.graceful = enable
}

func (e *Engine) SetMode(mode string) {
	e.mode = mode
}

// Start 在使用 http.ListenAndServe 前调用，用于触发 debug 模式下的路由打印
func (e *Engine) Start() {
	if e.mode == "debug" {
		e.PrintRoutes()
	}
}

func (e *Engine) PrintRoutes() {
	fmt.Println("\n========== 已注册路由 ==========")
	e.router.printRoutes("/", e.router.trie.root)
	fmt.Println("=======================================")
}
