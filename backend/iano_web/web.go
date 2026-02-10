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

	fmt.Printf("Server starting on %s\n", addr)
	return e.engine.ListenAndServe()
}

func (e *Engine) runWithGracefulShutdown(addr string) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("Server starting on %s (with graceful shutdown)\n", addr)
		serverErr <- e.engine.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		fmt.Printf("\nReceived signal: %v\n", sig)
		fmt.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := e.engine.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		fmt.Println("Server gracefully stopped")
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

	fmt.Printf("Server starting on %s with TLS\n", addr)
	return e.engine.ListenAndServeTLS(certFile, keyFile)
}

func (e *Engine) runTLSWithGracefulShutdown(addr, certFile, keyFile string) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("Server starting on %s with TLS (with graceful shutdown)\n", addr)
		serverErr <- e.engine.ListenAndServeTLS(certFile, keyFile)
	}()

	select {
	case err := <-serverErr:
		return err
	case sig := <-quit:
		fmt.Printf("\nReceived signal: %v\n", sig)
		fmt.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := e.engine.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		fmt.Println("Server gracefully stopped")
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
