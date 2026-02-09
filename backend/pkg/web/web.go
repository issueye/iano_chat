package web

import (
	"context"
	"fmt"
	"net/http"
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

func (e *Engine) GET(pattern string, handlers ...HandlerFunc) {
	allHandlers := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	allHandlers = append(allHandlers, e.middleware...)
	allHandlers = append(allHandlers, handlers...)
	e.router.Get(pattern, allHandlers...)
}

func (e *Engine) POST(pattern string, handlers ...HandlerFunc) {
	allHandlers := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	allHandlers = append(allHandlers, e.middleware...)
	allHandlers = append(allHandlers, handlers...)
	e.router.Post(pattern, allHandlers...)
}

func (e *Engine) PUT(pattern string, handlers ...HandlerFunc) {
	allHandlers := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	allHandlers = append(allHandlers, e.middleware...)
	allHandlers = append(allHandlers, handlers...)
	e.router.Put(pattern, allHandlers...)
}

func (e *Engine) DELETE(pattern string, handlers ...HandlerFunc) {
	allHandlers := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	allHandlers = append(allHandlers, e.middleware...)
	allHandlers = append(allHandlers, handlers...)
	e.router.Delete(pattern, allHandlers...)
}

func (e *Engine) PATCH(pattern string, handlers ...HandlerFunc) {
	allHandlers := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	allHandlers = append(allHandlers, e.middleware...)
	allHandlers = append(allHandlers, handlers...)
	e.router.Patch(pattern, allHandlers...)
}

func (e *Engine) OPTIONS(pattern string, handlers ...HandlerFunc) {
	allHandlers := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	allHandlers = append(allHandlers, e.middleware...)
	allHandlers = append(allHandlers, handlers...)
	e.router.Options(pattern, allHandlers...)
}

func (e *Engine) HEAD(pattern string, handlers ...HandlerFunc) {
	allHandlers := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	allHandlers = append(allHandlers, e.middleware...)
	allHandlers = append(allHandlers, handlers...)
	e.router.Head(pattern, allHandlers...)
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
	fn(e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	e.router.handleRequest(c)
}

func (e *Engine) Run(addr string) error {
	e.engine = &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
	}

	fmt.Printf("Server starting on %s\n", addr)
	return e.engine.ListenAndServe()
}

func (e *Engine) RunTLS(addr, certFile, keyFile string) error {
	e.engine = &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
	}

	fmt.Printf("Server starting on %s with TLS\n", addr)
	return e.engine.ListenAndServeTLS(certFile, keyFile)
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
