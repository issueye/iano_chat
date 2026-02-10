package web

import (
	"net/http"
	"strings"
)

type HandlerFunc func(c *Context)

type MiddlewareFunc = HandlerFunc

type Route struct {
	handlers []HandlerFunc
	method   string
	pattern  string
}

type Router struct {
	prefix     string
	middleware []MiddlewareFunc
	trie       *TrieRouter
}

func NewRouter() *Router {
	return &Router{
		trie: NewTrieRouter(),
	}
}

func (r *Router) addRoute(method, pattern string, handlers []HandlerFunc, paramKeys []string) *Route {
	route := &Route{
		handlers: handlers,
		method:   method,
		pattern:  pattern,
	}

	r.trie.insert(pattern, route)

	return route
}

func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.middleware = append(r.middleware, middleware...)
}

func (r *Router) Get(pattern string, handlers ...HandlerFunc) {
	r.Use(r.middleware...)
	allHandlers := make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
	allHandlers = append(allHandlers, r.middleware...)
	allHandlers = append(allHandlers, handlers...)
	r.addRoute("GET", pattern, allHandlers, r.extractParamKeys(pattern))
}

func (r *Router) Post(pattern string, handlers ...HandlerFunc) {
	r.Use(r.middleware...)
	allHandlers := make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
	allHandlers = append(allHandlers, r.middleware...)
	allHandlers = append(allHandlers, handlers...)
	r.addRoute("POST", pattern, allHandlers, r.extractParamKeys(pattern))
}

func (r *Router) Put(pattern string, handlers ...HandlerFunc) {
	r.Use(r.middleware...)
	allHandlers := make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
	allHandlers = append(allHandlers, r.middleware...)
	allHandlers = append(allHandlers, handlers...)
	r.addRoute("PUT", pattern, allHandlers, r.extractParamKeys(pattern))
}

func (r *Router) Delete(pattern string, handlers ...HandlerFunc) {
	r.Use(r.middleware...)
	allHandlers := make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
	allHandlers = append(allHandlers, r.middleware...)
	allHandlers = append(allHandlers, handlers...)
	r.addRoute("DELETE", pattern, allHandlers, r.extractParamKeys(pattern))
}

func (r *Router) Patch(pattern string, handlers ...HandlerFunc) {
	r.Use(r.middleware...)
	allHandlers := make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
	allHandlers = append(allHandlers, r.middleware...)
	allHandlers = append(allHandlers, handlers...)
	r.addRoute("PATCH", pattern, allHandlers, r.extractParamKeys(pattern))
}

func (r *Router) Options(pattern string, handlers ...HandlerFunc) {
	r.Use(r.middleware...)
	allHandlers := make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
	allHandlers = append(allHandlers, r.middleware...)
	allHandlers = append(allHandlers, handlers...)
	r.addRoute("OPTIONS", pattern, allHandlers, r.extractParamKeys(pattern))
}

func (r *Router) Head(pattern string, handlers ...HandlerFunc) {
	r.Use(r.middleware...)
	allHandlers := make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
	allHandlers = append(allHandlers, r.middleware...)
	allHandlers = append(allHandlers, handlers...)
	r.addRoute("HEAD", pattern, allHandlers, r.extractParamKeys(pattern))
}

func (r *Router) All(pattern string, handlers ...HandlerFunc) {
	r.Use(r.middleware...)
	allHandlers := make([]HandlerFunc, 0, len(r.middleware)+len(handlers))
	allHandlers = append(allHandlers, r.middleware...)
	allHandlers = append(allHandlers, handlers...)
	r.addRoute("*", pattern, allHandlers, r.extractParamKeys(pattern))
}

func (r *Router) Group(prefix string, fn func(r *Router)) {
	fn(r)
}

func (r *Router) extractParamKeys(pattern string) []string {
	var keys []string
	parts := strings.Split(pattern, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			keys = append(keys, strings.TrimPrefix(part, ":"))
		}
	}
	return keys
}

func (r *Router) handleRequest(c *Context) {
	matchedRoute, params := r.trie.search(c.Method, c.Path)

	if matchedRoute == nil {
		matchedRoute, params = r.trie.search("*", c.Path)
	}

	if matchedRoute == nil {
		c.Status(http.StatusNotFound)
		c.String(http.StatusNotFound, "404 Not Found")
		return
	}

	c.Params = params
	c.handlers = matchedRoute.handlers
	c.Next()
}
