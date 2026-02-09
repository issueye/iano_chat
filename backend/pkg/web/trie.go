package web

import (
	"strings"
)

type trieNode struct {
	children   map[string]*trieNode
	paramChild *trieNode
	paramName  string
	wildcard   bool
	routes     map[string]*Route
}

func newTrieNode() *trieNode {
	return &trieNode{
		children: make(map[string]*trieNode),
		routes:   make(map[string]*Route),
	}
}

type TrieRouter struct {
	root *trieNode
}

func NewTrieRouter() *TrieRouter {
	return &TrieRouter{
		root: newTrieNode(),
	}
}

func (tr *TrieRouter) insert(pattern string, route *Route) {
	parts := splitPath(pattern)
	node := tr.root

	for i, part := range parts {
		if part == "" {
			continue
		}

		if strings.HasPrefix(part, ":") {
			if node.paramChild == nil {
				node.paramChild = newTrieNode()
				node.paramChild.paramName = strings.TrimPrefix(part, ":")
			}
			node = node.paramChild
		} else if part == "*" {
			node.wildcard = true
			if _, ok := node.routes[route.method]; !ok {
				node.routes[route.method] = route
			}
			return
		} else {
			if _, ok := node.children[part]; !ok {
				node.children[part] = newTrieNode()
			}
			node = node.children[part]
		}

		if i == len(parts)-1 {
			node.routes[route.method] = route
		}
	}
}

func (tr *TrieRouter) search(method, path string) (*Route, map[string]string) {
	parts := splitPath(path)
	node := tr.root
	params := make(map[string]string)

	for _, part := range parts {
		if part == "" {
			continue
		}

		if child, ok := node.children[part]; ok {
			node = child
		} else if node.paramChild != nil {
			params[node.paramChild.paramName] = part
			node = node.paramChild
		} else if node.wildcard {
			if route, ok := node.routes[method]; ok {
				return route, params
			}
			if route, ok := node.routes["*"]; ok {
				return route, params
			}
			return nil, nil
		} else {
			return nil, nil
		}
	}

	if route, ok := node.routes[method]; ok {
		return route, params
	}
	if route, ok := node.routes["*"]; ok {
		return route, params
	}

	return nil, nil
}

func splitPath(path string) []string {
	return strings.Split(path, "/")
}
