// Package builtin - HTTP 模块
// 提供 HTTP 请求功能

package builtin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// HTTPModule HTTP 请求模块
type HTTPModule struct {
	client  *http.Client
	headers map[string]string
}

// NewHTTPModule 创建 HTTP 模块
func NewHTTPModule(timeout time.Duration) *HTTPModule {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPModule{
		client: &http.Client{
			Timeout: timeout,
		},
		headers: make(map[string]string),
	}
}

// Name 模块名称
func (m *HTTPModule) Name() string {
	return "http"
}

// Register 注册模块
func (m *HTTPModule) Register(vm *goja.Runtime) error {
	httpObj := map[string]interface{}{
		"get":    m.makeGet(vm),
		"post":   m.makePost(vm),
		"put":    m.makePut(vm),
		"delete": m.makeDelete(vm),
		"setHeader": func(key, value string) {
			m.headers[key] = value
		},
	}

	return vm.Set("http", httpObj)
}

// makeGet 创建 GET 请求函数
func (m *HTTPModule) makeGet(vm *goja.Runtime) func(string, map[string]interface{}) (map[string]interface{}, error) {
	return func(urlStr string, options map[string]interface{}) (map[string]interface{}, error) {
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return nil, err
		}

		// 添加查询参数
		if options != nil {
			if params, ok := options["params"].(map[string]interface{}); ok {
				q := req.URL.Query()
				for key, value := range params {
					q.Add(key, fmt.Sprint(value))
				}
				req.URL.RawQuery = q.Encode()
			}
		}

		return m.doRequest(req)
	}
}

// makePost 创建 POST 请求函数
func (m *HTTPModule) makePost(vm *goja.Runtime) func(string, map[string]interface{}) (map[string]interface{}, error) {
	return func(urlStr string, options map[string]interface{}) (map[string]interface{}, error) {
		var body io.Reader

		if options != nil {
			if data, ok := options["json"].(map[string]interface{}); ok {
				jsonData, _ := json.Marshal(data)
				body = bytes.NewReader(jsonData)
			} else if data, ok := options["body"].(string); ok {
				body = strings.NewReader(data)
			}
		}

		req, err := http.NewRequest("POST", urlStr, body)
		if err != nil {
			return nil, err
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		return m.doRequest(req)
	}
}

// makePut 创建 PUT 请求函数
func (m *HTTPModule) makePut(vm *goja.Runtime) func(string, map[string]interface{}) (map[string]interface{}, error) {
	return func(urlStr string, options map[string]interface{}) (map[string]interface{}, error) {
		var body io.Reader

		if options != nil {
			if data, ok := options["json"].(map[string]interface{}); ok {
				jsonData, _ := json.Marshal(data)
				body = bytes.NewReader(jsonData)
			}
		}

		req, err := http.NewRequest("PUT", urlStr, body)
		if err != nil {
			return nil, err
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		return m.doRequest(req)
	}
}

// makeDelete 创建 DELETE 请求函数
func (m *HTTPModule) makeDelete(vm *goja.Runtime) func(string, map[string]interface{}) (map[string]interface{}, error) {
	return func(urlStr string, options map[string]interface{}) (map[string]interface{}, error) {
		req, err := http.NewRequest("DELETE", urlStr, nil)
		if err != nil {
			return nil, err
		}
		return m.doRequest(req)
	}
}

// doRequest 执行 HTTP 请求
func (m *HTTPModule) doRequest(req *http.Request) (map[string]interface{}, error) {
	// 添加默认 headers
	for key, value := range m.headers {
		req.Header.Set(key, value)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    resp.Header,
		"body":       string(body),
	}

	// 尝试解析 JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err == nil {
		result["json"] = jsonData
	}

	return result, nil
}
