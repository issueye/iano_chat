package web

import (
	"net/http"
	"regexp"
	"strings"
)

type HandlerFunc func(c *Context)

type MiddlewareFunc = HandlerFunc

type Route struct {
	handlers  []HandlerFunc
	method    string
	pattern   string
	paramKeys []string
	regex     *regexp.Regexp
}

type Router struct {
	routes      []*Route
	paramRoutes map[string]*Route
	prefix      string
	middleware  []MiddlewareFunc
}

func NewRouter() *Router {
	return &Router{
		paramRoutes: make(map[string]*Route),
	}
}

func (r *Router) addRoute(method, pattern string, handlers []HandlerFunc, paramKeys []string) *Route {
	route := &Route{
		handlers:  handlers,
		method:    method,
		pattern:   pattern,
		paramKeys: paramKeys,
	}

	if len(paramKeys) > 0 {
		regexStr := r.convertToRegex(pattern)
		regex, _ := regexp.Compile(regexStr)
		route.regex = regex
		r.paramRoutes[pattern] = route
	} else {
		r.routes = append(r.routes, route)
	}

	return route
}

func (r *Router) convertToRegex(pattern string) string {
	parts := strings.Split(pattern, "/")
	var result []string
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			result = append(result, "([^/]+)")
		} else if part == "*" {
			result = append(result, "(.*)")
		} else if part != "" {
			result = append(result, part)
		}
	}
	return "^" + strings.Join(result, "/") + "/?$"
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
	var matchedRoute *Route
	var params map[string]string

	for _, route := range r.routes {
		if (route.method == c.Method || route.method == "*") && route.pattern == c.Path {
			matchedRoute = route
			break
		}
	}

	if matchedRoute == nil {
		for _, route := range r.paramRoutes {
			if route.method == c.Method || route.method == "*" {
				if route.regex != nil {
					matches := route.regex.FindStringSubmatch(c.Path)
					if matches != nil {
						params = make(map[string]string)
						for i, key := range route.paramKeys {
							if i+1 < len(matches) {
								params[key] = matches[i+1]
							}
						}
						matchedRoute = route
						break
					}
				}
			}
		}
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

func (r *Router) match(c *Context) bool {
	for _, route := range r.routes {
		if (route.method == c.Method || route.method == "*") && route.pattern == c.Path {
			return true
		}
	}
	return false
}
