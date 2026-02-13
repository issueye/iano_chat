package tools

import (
	"context"
	"fmt"
	"net/http"
	"time"

	duckduckgoV2 "github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
)

type DuckDuckGoTool struct {
	config *duckduckgoV2.Config
	tool   tool.InvokableTool
}

func NewDuckDuckGoTool(timeout int) (*DuckDuckGoTool, error) {
	ctx := context.Background()
	cfg := &duckduckgoV2.Config{ // All of these parameters are default values, for demonstration purposes only
		ToolName: "duckduckgo_search",                        // 工具的名称
		ToolDesc: "search web for information by duckduckgo", // 工具的描述信息
		Timeout:  30,                                         // 单次请求的最大耗时, 默认 30 秒// 发送HTTP请求的客户端实例// 若设置了HTTPClient，则Timeout配置项将不再生效// 可选配置。默认值：&http.client{Timeout: Timeout}
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		MaxResults: 3,                     // 返回结果的数量
		Region:     duckduckgoV2.RegionCN, // 地理区域限定
	}

	tool, err := duckduckgoV2.NewTextSearchTool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("NewTextSearchTool of duckduckgo failed, err=%w", err)
	}

	return &DuckDuckGoTool{config: cfg, tool: tool}, nil
}
