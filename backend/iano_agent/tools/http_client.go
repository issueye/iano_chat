package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// 默认 HTTP 客户端配置
const (
	defaultTimeout         = 30 * time.Second
	defaultMaxIdleConns    = 100
	defaultMaxConnsPerHost = 10
	defaultIdleConnTimeout = 90 * time.Second

	// 安全限制
	maxResponseSize  = 10 * 1024 * 1024 // 最大响应体 10MB
	maxRedirectCount = 5                // 最大重定向次数
)

// 允许的 HTTP 方法
var allowedMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   true,
	http.MethodDelete:  true,
	http.MethodHead:    true,
	http.MethodOptions: true,
}

// httpClient 自定义 HTTP 客户端，带超时和连接池配置
var httpClient = &http.Client{
	Timeout: defaultTimeout,
	Transport: &http.Transport{
		MaxIdleConns:        defaultMaxIdleConns,
		MaxIdleConnsPerHost: defaultMaxConnsPerHost,
		IdleConnTimeout:     defaultIdleConnTimeout,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirectCount {
			return fmt.Errorf("重定向次数超过限制: %d", maxRedirectCount)
		}
		return nil
	},
}

type HTTPClientTool struct{}

// Info 返回工具信息
func (t *HTTPClientTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "http_request",
		Desc: "用于发送 HTTP 请求，支持 GET/POST/PUT/DELETE 等方法",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"method": {
				Type:     schema.String,
				Desc:     "请求方法，可选值: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS",
				Required: true,
			},
			"url": {
				Type:     schema.String,
				Desc:     "请求 URL，必须是有效的 HTTP/HTTPS URL",
				Required: true,
			},
			"body": {
				Type: schema.String,
				Desc: "请求体，仅用于 POST/PUT/PATCH 方法",
			},
			"headers": {
				Type: schema.Object,
				Desc: "请求头，键值对格式",
			},
			"query": {
				Type: schema.Object,
				Desc: "查询参数，键值对格式",
			},
		}),
	}, nil
}

// InvokableRun 执行 HTTP 请求
func (t *HTTPClientTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Method  string            `json:"method"`
		URL     string            `json:"url"`
		Body    string            `json:"body"`
		Headers map[string]string `json:"headers"`
		Query   map[string]string `json:"query"`
	}

	// 解析参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	// 验证参数
	if err := t.validateArgs(&args); err != nil {
		return "", err
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, args.Method, args.URL, strings.NewReader(args.Body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for k, v := range args.Headers {
		// 过滤掉一些敏感头
		if isSensitiveHeader(k) {
			continue
		}
		req.Header.Set(k, v)
	}

	// 设置默认 User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "IanoChat-Agent/1.0")
	}

	// 设置查询参数
	q := req.URL.Query()
	for k, v := range args.Query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	// 限制响应体大小
	limitedReader := io.LimitReader(resp.Body, maxResponseSize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查是否超过大小限制
	if int64(len(body)) > maxResponseSize {
		return "", fmt.Errorf("响应体超过最大限制 %d MB", maxResponseSize/1024/1024)
	}

	// 构造响应结果
	result := map[string]interface{}{
		"statusCode": resp.StatusCode,
		"status":     resp.Status,
		"headers":    resp.Header,
		"body":       string(body),
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("序列化响应失败: %w", err)
	}

	return string(resultJSON), nil
}

// validateArgs 验证请求参数
func (t *HTTPClientTool) validateArgs(args *struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
	Query   map[string]string `json:"query"`
}) error {
	// 验证方法
	method := strings.ToUpper(strings.TrimSpace(args.Method))
	if method == "" {
		return fmt.Errorf("请求方法不能为空")
	}
	if !allowedMethods[method] {
		return fmt.Errorf("不支持的请求方法: %s", args.Method)
	}
	args.Method = method

	// 验证 URL
	args.URL = strings.TrimSpace(args.URL)
	if args.URL == "" {
		return fmt.Errorf("请求 URL 不能为空")
	}

	parsedURL, err := url.Parse(args.URL)
	if err != nil {
		return fmt.Errorf("无效的 URL: %w", err)
	}

	// 只允许 HTTP/HTTPS 协议
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("不支持的协议: %s，仅支持 HTTP/HTTPS", parsedURL.Scheme)
	}

	// 禁止访问本地地址
	if isLocalhost(parsedURL.Hostname()) {
		return fmt.Errorf("禁止访问本地地址")
	}

	return nil
}

// isSensitiveHeader 检查是否为敏感请求头
func isSensitiveHeader(header string) bool {
	sensitiveHeaders := []string{
		"Authorization",
		"Cookie",
		"Set-Cookie",
		"Proxy-Authorization",
	}

	headerLower := strings.ToLower(header)
	for _, h := range sensitiveHeaders {
		if headerLower == strings.ToLower(h) {
			return true
		}
	}
	return false
}

// isLocalhost 检查是否为本地地址
func isLocalhost(host string) bool {
	localhostNames := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"0.0.0.0",
	}

	hostLower := strings.ToLower(host)
	for _, name := range localhostNames {
		if hostLower == name {
			return true
		}
	}

	// 检查是否为内网 IP
	if strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.") {
		return true
	}

	return false
}
