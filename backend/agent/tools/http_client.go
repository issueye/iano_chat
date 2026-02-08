package tools

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type HTTPClientTool struct {
}

func (t *HTTPClientTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "http_request",
		Desc: "用于发送 HTTP 请求",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"method": {
				Type:     schema.String,
				Desc:     "请求方法",
				Required: true,
			},
			"url": {
				Type:     schema.String,
				Desc:     "请求 URL",
				Required: true,
			},
			"body": {
				Type: schema.String,
				Desc: "请求体",
			},
			"headers": {
				Type: schema.Object,
				Desc: "请求头",
			},
			"query": {
				Type: schema.Object,
				Desc: "查询参数",
			},
		}),
	}, nil
}

func (t *HTTPClientTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Method  string            `json:"method"`
		URL     string            `json:"url"`
		Body    string            `json:"body"`
		Headers map[string]string `json:"headers"`
		Query   map[string]string `json:"query"`
	}

	// 进行http请求
	req, err := http.NewRequestWithContext(ctx, args.Method, args.URL, strings.NewReader(args.Body))
	if err != nil {
		return "", err
	}
	for k, v := range args.Headers {
		req.Header.Set(k, v)
	}
	q := req.URL.Query()
	for k, v := range args.Query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
