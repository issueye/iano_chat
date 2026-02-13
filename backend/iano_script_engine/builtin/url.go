// Package builtin - URL 模块
// 提供 URL 处理功能

package builtin

import (
	"net/url"

	"github.com/dop251/goja"
)

// URLModule URL 处理模块
type URLModule struct{}

// NewURLModule 创建 URL 模块
func NewURLModule() *URLModule {
	return &URLModule{}
}

// Name 模块名称
func (m *URLModule) Name() string {
	return "url"
}

// Register 注册模块
func (m *URLModule) Register(vm *goja.Runtime) error {
	urlObj := map[string]interface{}{
		"parse": func(rawURL string) (map[string]interface{}, error) {
			u, err := url.Parse(rawURL)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"scheme":   u.Scheme,
				"host":     u.Host,
				"path":     u.Path,
				"query":    u.RawQuery,
				"fragment": u.Fragment,
			}, nil
		},
		"encode": url.QueryEscape,
		"decode": url.QueryUnescape,
	}

	return vm.Set("url", urlObj)
}
