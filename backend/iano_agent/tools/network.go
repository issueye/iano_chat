package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type PingTool struct{}

func NewPingTool() *PingTool {
	return &PingTool{}
}

func (t *PingTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "ping",
		Desc: "检测主机是否可达",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"host": {
				Type:     schema.String,
				Desc:     "目标主机名或 IP 地址",
				Required: true,
			},
			"count": {
				Type:     schema.Number,
				Desc:     "Ping 次数（默认 4）",
				Required: false,
			},
			"timeout": {
				Type:     schema.Number,
				Desc:     "超时时间（秒，默认 5）",
				Required: false,
			},
		}),
	}, nil
}

func (t *PingTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Host    string `json:"host"`
		Count   int    `json:"count"`
		Timeout int    `json:"timeout"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Host == "" {
		return "", fmt.Errorf("目标主机不能为空")
	}

	count := args.Count
	if count <= 0 || count > 10 {
		count = 4
	}

	timeout := time.Duration(args.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ips, err := net.LookupIP(args.Host)
	if err != nil {
		return fmt.Sprintf("无法解析主机: %s", args.Host), nil
	}

	var success int
	start := time.Now()

	for i := 0; i < count; i++ {
		conn, err := net.DialTimeout("ip4:icmp", args.Host, timeout)
		if err == nil {
			conn.Close()
			success++
		}
		time.Sleep(time.Second)
	}

	avgTime := time.Since(start) / time.Duration(count)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Ping %s [%s]\n", args.Host, ips[0].String()))
	result.WriteString(fmt.Sprintf("回复: %d/%d, 丢失: %d%%\n", success, count, (count-success)*100/count))
	if success > 0 {
		result.WriteString(fmt.Sprintf("往返时间: ~%v\n", avgTime))
	}

	return result.String(), nil
}

type DNSLookupTool struct{}

func NewDNSLookupTool() *DNSLookupTool {
	return &DNSLookupTool{}
}

func (t *DNSLookupTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "dns_lookup",
		Desc: "DNS 查询",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"host": {
				Type:     schema.String,
				Desc:     "要查询的主机名",
				Required: true,
			},
			"record_type": {
				Type:     schema.String,
				Desc:     "记录类型: A, AAAA, CNAME, MX, TXT, NS",
				Required: false,
			},
		}),
	}, nil
}

func (t *DNSLookupTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Host       string `json:"host"`
		RecordType string `json:"record_type"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.Host == "" {
		return "", fmt.Errorf("主机名不能为空")
	}

	recordType := strings.ToUpper(args.RecordType)
	if recordType == "" {
		recordType = "A"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("DNS 查询: %s (类型: %s)\n\n", args.Host, recordType))

	switch recordType {
	case "A":
		ips, err := net.LookupIP(args.Host)
		if err != nil {
			return fmt.Sprintf("查询失败: %v", err), nil
		}
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				result.WriteString(fmt.Sprintf("A %s -> %s\n", args.Host, ipv4.String()))
			}
		}
	case "AAAA":
		ips, err := net.LookupIP(args.Host)
		if err != nil {
			return fmt.Sprintf("查询失败: %v", err), nil
		}
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 == nil {
				result.WriteString(fmt.Sprintf("AAAA %s -> %s\n", args.Host, ip.String()))
			}
		}
	case "CNAME":
		cnames, err := net.LookupCNAME(args.Host)
		if err != nil {
			return fmt.Sprintf("查询失败: %v", err), nil
		}
		for _, cname := range cnames {
			result.WriteString(fmt.Sprintf("CNAME %s -> %s\n", args.Host, cname))
		}
	case "MX":
		mxs, err := net.LookupMX(args.Host)
		if err != nil {
			return fmt.Sprintf("查询失败: %v", err), nil
		}
		for _, mx := range mxs {
			result.WriteString(fmt.Sprintf("MX %s -> %s (优先级: %d)\n", args.Host, mx.Host, mx.Pref))
		}
	case "TXT":
		txts, err := net.LookupTXT(args.Host)
		if err != nil {
			return fmt.Sprintf("查询失败: %v", err), nil
		}
		for _, txt := range txts {
			result.WriteString(fmt.Sprintf("TXT %s -> \"%s\"\n", args.Host, txt))
		}
	case "NS":
		nss, err := net.LookupNS(args.Host)
		if err != nil {
			return fmt.Sprintf("查询失败: %v", err), nil
		}
		for _, ns := range nss {
			result.WriteString(fmt.Sprintf("NS %s -> %s\n", args.Host, ns.Host))
		}
	default:
		return fmt.Sprintf("不支持的记录类型: %s", recordType), nil
	}

	return result.String(), nil
}

type HTTPHeadersTool struct{}

func NewHTTPHeadersTool() *HTTPHeadersTool {
	return &HTTPHeadersTool{}
}

func (t *HTTPHeadersTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "http_headers",
		Desc: "获取 URL 的 HTTP 响应头",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type:     schema.String,
				Desc:     "目标 URL",
				Required: true,
			},
		}),
	}, nil
}

func (t *HTTPHeadersTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		URL string `json:"url"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	if args.URL == "" {
		return "", fmt.Errorf("URL 不能为空")
	}

	if !strings.HasPrefix(strings.ToLower(args.URL), "http") {
		args.URL = "https://" + args.URL
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", args.URL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "IanoChat-Agent/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("URL: %s\n", args.URL))
	result.WriteString(fmt.Sprintf("状态: %s\n", resp.Status))
	result.WriteString("\n响应头:\n")
	for k, v := range resp.Header {
		result.WriteString(fmt.Sprintf("  %s: %s\n", k, strings.Join(v, ", ")))
	}

	return result.String(), nil
}
