package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type DuckDuckGoTool struct {
	config *duckduckgo.Config
	tool   tool.InvokableTool
}

func NewDuckDuckGoTool() (*DuckDuckGoTool, error) {
	ctx := context.Background()
	config := &duckduckgo.Config{
		MaxResults: 3, // Limit to return 20 results
		Region:     duckduckgo.RegionWT,
		Timeout:    10 * time.Second,
	}

	tool, err := duckduckgo.NewTextSearchTool(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("NewTextSearchTool of duckduckgo failed, err=%w", err)
	}

	return &DuckDuckGoTool{config: config, tool: tool}, nil
}

func (t *DuckDuckGoTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "web_search",
		Desc: "用于搜索互联网",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.String,
				Desc:     "搜索查询",
				Required: true,
			},
		}),
	}, nil
}

func (t *DuckDuckGoTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Query string `json:"query"`
	}

	// 解析参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	searchReq := &duckduckgo.TextSearchRequest{
		Query: args.Query,
	}

	jsonReq, err := json.Marshal(searchReq)
	if err != nil {
		return "", err
	}

	resp, err := t.tool.InvokableRun(ctx, string(jsonReq), opts...)
	if err != nil {
		return "", err
	}

	var searchResp duckduckgo.TextSearchResponse
	if err = json.Unmarshal([]byte(resp), &searchResp); err != nil {
		return "", fmt.Errorf("Unmarshal of search response failed, err=%w", err)
	}

	// 提取搜索结果
	var results []string
	for _, result := range searchResp.Results {
		results = append(results, result.Summary)
	}

	return fmt.Sprintf("搜索结果\n%s", strings.Join(results, "\n")), nil
}
