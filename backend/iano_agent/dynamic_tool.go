package iano_agent

import (
	"context"
	"encoding/json"
	"fmt"

	script_engine "iano_script_engine"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type DynamicTool struct {
	name       string
	desc       string
	parameters []ToolParamDef
	handler    DynamicToolHandler
}

type ToolParamDef struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Desc     string      `json:"desc"`
	Required bool        `json:"required"`
	Default  interface{} `json:"default,omitempty"`
	Enum     []string    `json:"enum,omitempty"`
}

type DynamicToolHandler func(ctx context.Context, params map[string]interface{}) (string, error)

type DynamicToolConfig struct {
	Name       string
	Desc       string
	Parameters []ToolParamDef
	Handler    DynamicToolHandler
}

func NewDynamicTool(cfg *DynamicToolConfig) *DynamicTool {
	return &DynamicTool{
		name:       cfg.Name,
		desc:       cfg.Desc,
		parameters: cfg.Parameters,
		handler:    cfg.Handler,
	}
}

func (t *DynamicTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	params := make(map[string]*schema.ParameterInfo)
	for _, p := range t.parameters {
		params[p.Name] = &schema.ParameterInfo{
			Type:     schema.DataType(p.Type),
			Desc:     p.Desc,
			Required: p.Required,
		}
	}

	return &schema.ToolInfo{
		Name:        t.name,
		Desc:        t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}, nil
}

func (t *DynamicTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	for _, p := range t.parameters {
		if p.Required {
			if _, exists := params[p.Name]; !exists {
				if p.Default != nil {
					params[p.Name] = p.Default
				} else {
					return "", fmt.Errorf("required parameter '%s' is missing", p.Name)
				}
			}
		}
		if _, exists := params[p.Name]; !exists && p.Default != nil {
			params[p.Name] = p.Default
		}
	}

	if t.handler == nil {
		return "", fmt.Errorf("tool handler not configured")
	}

	return t.handler(ctx, params)
}

type HTTPToolConfig struct {
	Name         string
	Desc         string
	Method       string
	URL          string
	Headers      map[string]string
	QueryParams  map[string]string
	BodyTemplate string
}

func NewHTTPTool(cfg *HTTPToolConfig) *DynamicTool {
	return NewDynamicTool(&DynamicToolConfig{
		Name: cfg.Name,
		Desc: cfg.Desc,
		Parameters: []ToolParamDef{
			{Name: "url", Type: "string", Desc: "Request URL", Required: true},
			{Name: "body", Type: "string", Desc: "Request body", Required: false},
			{Name: "headers", Type: "object", Desc: "Request headers", Required: false},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (string, error) {
			return fmt.Sprintf("HTTP %s to %s", cfg.Method, params["url"]), nil
		},
	})
}

type FunctionToolConfig struct {
	Name       string
	Desc       string
	Parameters []ToolParamDef
	Function   string
}

func NewFunctionTool(cfg *FunctionToolConfig) *DynamicTool {
	return NewDynamicTool(&DynamicToolConfig{
		Name:       cfg.Name,
		Desc:       cfg.Desc,
		Parameters: cfg.Parameters,
		Handler: func(ctx context.Context, params map[string]interface{}) (string, error) {
			return fmt.Sprintf("Executed function %s with params: %v", cfg.Function, params), nil
		},
	})
}

type ScriptToolConfig struct {
	Name       string
	Desc       string
	Parameters []ToolParamDef
	Script     string
	Language   string
	Engine     script_engine.Engine
}

func NewScriptTool(cfg *ScriptToolConfig) *DynamicTool {
	engine := cfg.Engine
	if engine == nil {
		engine = script_engine.NewEngine(nil)
	}

	return NewDynamicTool(&DynamicToolConfig{
		Name:       cfg.Name,
		Desc:       cfg.Desc,
		Parameters: cfg.Parameters,
		Handler: func(ctx context.Context, params map[string]interface{}) (string, error) {
			result, err := engine.Execute(ctx, cfg.Script, params)
			if err != nil {
				return "", fmt.Errorf("script execution failed: %w", err)
			}
			return result.Value.(string), nil
		},
	})
}

func ToolParamsFromJSON(jsonStr string) ([]ToolParamDef, error) {
	if jsonStr == "" {
		return nil, nil
	}

	var params []ToolParamDef
	if err := json.Unmarshal([]byte(jsonStr), &params); err != nil {
		return nil, fmt.Errorf("failed to parse tool parameters: %w", err)
	}
	return params, nil
}
