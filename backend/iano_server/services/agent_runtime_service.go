package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	iano "iano_agent"
	script_engine "iano_script_engine"
	"iano_server/models"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

// AgentRuntimeService Agent 运行时服务
// 完全解耦会话和 Agent，只负责根据 Agent 配置创建和使用 Agent 实例
type AgentRuntimeService struct {
	db              *gorm.DB
	agentService    *AgentService
	providerService *ProviderService
	toolService     *ToolService
	mcpService      *MCPService
	modelCache      map[string]model.ToolCallingChatModel
}

// NewAgentRuntimeService 创建 Agent 运行时服务
func NewAgentRuntimeService(
	db *gorm.DB,
	agentService *AgentService,
	providerService *ProviderService,
	toolService *ToolService,
) *AgentRuntimeService {
	return &AgentRuntimeService{
		db:              db,
		agentService:    agentService,
		providerService: providerService,
		toolService:     toolService,
		modelCache:      make(map[string]model.ToolCallingChatModel),
	}
}

// NewAgentRuntimeServiceWithMCP 创建带 MCP 支持的 Agent 运行时服务
func NewAgentRuntimeServiceWithMCP(
	db *gorm.DB,
	agentService *AgentService,
	providerService *ProviderService,
	toolService *ToolService,
	mcpService *MCPService,
) *AgentRuntimeService {
	return &AgentRuntimeService{
		db:              db,
		agentService:    agentService,
		providerService: providerService,
		toolService:     toolService,
		mcpService:      mcpService,
		modelCache:      make(map[string]model.ToolCallingChatModel),
	}
}

// GetAgent 根据 Agent ID 获取 Agent 实例
// 每次调用都创建新的 Agent 实例，不缓存，不与会话绑定
// workDir 用于限制文件操作工具的操作范围，为空则使用当前目录
func (s *AgentRuntimeService) GetAgent(ctx context.Context, agentID string, workDir string) (*AgentWrapper, error) {
	agent, err := s.agentService.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	chatModel, err := s.getOrCreateChatModel(ctx, agent.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat model: %w", err)
	}

	allowedTools := s.parseTools(agent.Tools)
	allowedCommands := s.getAllowedCommands(allowedTools)

	opts := []iano.Option{
		iano.WithSystemPrompt(agent.Instructions),
	}
	if len(allowedTools) > 0 {
		opts = append(opts, iano.WithAllowedTools(allowedTools))
	}
	if workDir != "" {
		opts = append(opts, iano.WithWorkDir(workDir))
	}
	if len(allowedCommands) > 0 {
		opts = append(opts, iano.WithAllowedCommands(allowedCommands))
	}

	agentInstance, err := iano.NewAgent(chatModel, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent instance: %w", err)
	}

	if err := s.loadToolsToAgent(ctx, agentInstance, agent); err != nil {
		slog.Warn("Failed to load tools to agent", "agentID", agentID, "error", err)
	}

	return &AgentWrapper{
		Agent:  agentInstance,
		Config: agent,
	}, nil
}

// AgentWrapper Agent 包装器
type AgentWrapper struct {
	Agent  *iano.Agent
	Config *models.Agent
}

// Chat 执行对话
func (w *AgentWrapper) Chat(ctx context.Context, messages []*schema.Message, callback iano.MessageCallback) (string, error) {
	return w.Agent.ChatWithHistory(ctx, messages, callback)
}

// GetAgentInfo 获取 Agent 信息
func (w *AgentWrapper) GetAgentInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":           w.Config.ID,
		"name":         w.Config.Name,
		"description":  w.Config.Description,
		"instructions": w.Config.Instructions,
		"tools":        w.Config.Tools,
		"providerID":   w.Config.ProviderID,
		"model":        w.Config.Model,
	}
}

// getOrCreateChatModel 获取或创建 ChatModel
func (s *AgentRuntimeService) getOrCreateChatModel(ctx context.Context, providerID string) (model.ToolCallingChatModel, error) {
	if providerID != "" {
		if m, exists := s.modelCache[providerID]; exists {
			return m, nil
		}

		provider, err := s.providerService.GetByID(providerID)
		if err != nil {
			return nil, fmt.Errorf("provider not found: %w", err)
		}

		return s.createChatModelFromProvider(ctx, provider)
	}

	defaultProvider, err := s.providerService.GetDefault()
	if err != nil {
		return nil, fmt.Errorf("no default provider configured")
	}

	if m, exists := s.modelCache[defaultProvider.ID]; exists {
		return m, nil
	}

	return s.createChatModelFromProvider(ctx, defaultProvider)
}

// createChatModelFromProvider 从 Provider 创建 ChatModel
func (s *AgentRuntimeService) createChatModelFromProvider(ctx context.Context, provider *models.Provider) (model.ToolCallingChatModel, error) {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     provider.BaseUrl,
		APIKey:      provider.ApiKey,
		Model:       provider.Model,
		Temperature: &provider.Temperature,
		MaxTokens:   &provider.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	s.modelCache[provider.ID] = chatModel
	return chatModel, nil
}

// parseTools 解析工具列表
func (s *AgentRuntimeService) parseTools(toolsStr string) []string {
	if toolsStr == "" {
		return nil
	}

	var tools []string
	if err := json.Unmarshal([]byte(toolsStr), &tools); err != nil {
		return []string{toolsStr}
	}
	return tools
}

// getAllowedCommands 从工具列表中获取允许执行的命令配置
func (s *AgentRuntimeService) getAllowedCommands(toolIDs []string) []string {
	var allowedCommands []string

	for _, toolID := range toolIDs {
		tool, err := s.toolService.GetByID(toolID)
		if err != nil {
			continue
		}

		if tool.Name == "command_execute" || tool.Name == "command" {
			cmdConfig := tool.GetCommandConfig()
			if cmdConfig != nil && len(cmdConfig.AllowedCommands) > 0 {
				allowedCommands = append(allowedCommands, cmdConfig.AllowedCommands...)
			}
		}
	}

	return allowedCommands
}

// loadToolsToAgent 加载工具到 Agent
func (s *AgentRuntimeService) loadToolsToAgent(ctx context.Context, agent *iano.Agent, config *models.Agent) error {
	tools := s.parseTools(config.Tools)
	for _, toolID := range tools {
		tool, err := s.toolService.GetByID(toolID)
		if err != nil {
			slog.Warn("Tool not found", "toolID", toolID, "error", err)
			continue
		}

		dynamicTool, err := s.createDynamicTool(tool)
		if err != nil {
			slog.Warn("Failed to create dynamic tool", "toolID", toolID, "error", err)
			continue
		}

		if err := agent.AddTool(tool.Name, dynamicTool); err != nil {
			slog.Warn("Failed to add tool to agent", "toolID", toolID, "error", err)
		}
	}

	if s.mcpService != nil {
		if err := s.loadMCPToolsToAgent(ctx, agent, config); err != nil {
			slog.Warn("Failed to load MCP tools to agent", "agentID", config.ID, "error", err)
		}
	}

	return nil
}

// parseMCPServers 解析 MCP 服务器列表
func (s *AgentRuntimeService) parseMCPServers(mcpServersStr string) []string {
	if mcpServersStr == "" {
		return nil
	}

	var servers []string
	if err := json.Unmarshal([]byte(mcpServersStr), &servers); err != nil {
		return []string{mcpServersStr}
	}
	return servers
}

// loadMCPToolsToAgent 加载 MCP 工具到 Agent
func (s *AgentRuntimeService) loadMCPToolsToAgent(ctx context.Context, agent *iano.Agent, config *models.Agent) error {
	mcpServers := s.parseMCPServers(config.MCPServers)
	if len(mcpServers) == 0 {
		return nil
	}

	for _, serverID := range mcpServers {
		server, err := s.mcpService.ServerService.GetByID(serverID)
		if err != nil {
			slog.Warn("MCP Server not found", "serverID", serverID, "error", err)
			continue
		}

		if server.Status != models.MCPServerStatusConnected {
			slog.Warn("MCP Server not connected", "serverID", serverID, "status", server.Status)
			continue
		}

		tools, err := s.mcpService.ServerToolService.GetByServerID(serverID)
		if err != nil {
			slog.Warn("Failed to get MCP server tools", "serverID", serverID, "error", err)
			continue
		}

		for _, tool := range tools {
			dynamicTool, err := s.createMCPDynamicTool(serverID, &tool)
			if err != nil {
				slog.Warn("Failed to create MCP dynamic tool", "toolName", tool.Name, "error", err)
				continue
			}

			toolName := fmt.Sprintf("mcp_%s_%s", serverID[:8], tool.Name)
			if err := agent.AddTool(toolName, dynamicTool); err != nil {
				slog.Warn("Failed to add MCP tool to agent", "toolName", toolName, "error", err)
			}
		}
	}

	return nil
}

// createMCPDynamicTool 创建 MCP 动态工具
func (s *AgentRuntimeService) createMCPDynamicTool(serverID string, tool *models.MCPServerTool) (*iano.DynamicTool, error) {
	var params []iano.ToolParamDef
	if tool.InputSchema != "" {
		var schema map[string]interface{}
		if err := json.Unmarshal([]byte(tool.InputSchema), &schema); err == nil {
			params = s.parseJSONSchemaToToolParams(schema)
		}
	}

	cfg := &iano.DynamicToolConfig{
		Name:       tool.Name,
		Desc:       tool.Description,
		Parameters: params,
		Handler:    s.createMCPToolHandler(serverID, tool.Name),
	}

	return iano.NewDynamicTool(cfg), nil
}

// parseJSONSchemaToToolParams 将 JSON Schema 转换为工具参数
func (s *AgentRuntimeService) parseJSONSchemaToToolParams(schema map[string]interface{}) []iano.ToolParamDef {
	var params []iano.ToolParamDef

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return params
	}

	required, _ := schema["required"].([]interface{})
	requiredMap := make(map[string]bool)
	for _, r := range required {
		if rStr, ok := r.(string); ok {
			requiredMap[rStr] = true
		}
	}

	for name, prop := range properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			continue
		}

		param := iano.ToolParamDef{
			Name:     name,
			Required: requiredMap[name],
		}

		if desc, ok := propMap["description"].(string); ok {
			param.Desc = desc
		}

		if paramType, ok := propMap["type"].(string); ok {
			param.Type = paramType
		}

		params = append(params, param)
	}

	return params
}

// createMCPToolHandler 创建 MCP 工具处理器
func (s *AgentRuntimeService) createMCPToolHandler(serverID string, toolName string) iano.DynamicToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (string, error) {
		result, err := s.mcpService.CallTool(ctx, serverID, toolName, params)
		if err != nil {
			return "", fmt.Errorf("MCP tool call failed: %w", err)
		}

		if result.IsError {
			return "", fmt.Errorf("MCP tool returned error: %v", result.Content)
		}

		if len(result.Content) == 0 {
			return "{}", nil
		}

		contentBytes, err := json.Marshal(result.Content)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}

		return string(contentBytes), nil
	}
}

// createDynamicTool 创建动态工具
func (s *AgentRuntimeService) createDynamicTool(tool *models.Tool) (*iano.DynamicTool, error) {
	params, err := iano.ToolParamsFromJSON(tool.Parameters)
	if err != nil {
		params = nil
	}

	cfg := &iano.DynamicToolConfig{
		Name:       tool.Name,
		Desc:       tool.Desc,
		Parameters: params,
		Handler:    s.createToolHandler(tool),
	}

	return iano.NewDynamicTool(cfg), nil
}

// createToolHandler 创建工具处理器
func (s *AgentRuntimeService) createToolHandler(tool *models.Tool) iano.DynamicToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (string, error) {
		s.toolService.IncrementCallCount(tool.ID)

		switch tool.Type {
		case models.ToolTypeScript:
			return s.executeScriptTool(ctx, tool, params)
		default:
			return "", fmt.Errorf("unsupported tool type: %s", tool.Type)
		}
	}
}

// executeScriptTool 执行脚本工具
func (s *AgentRuntimeService) executeScriptTool(ctx context.Context, tool *models.Tool, params map[string]interface{}) (string, error) {
	engine := script_engine.NewEngine(script_engine.DefaultConfig())
	result, err := engine.Execute(ctx, tool.ScriptContent, params)
	if err != nil {
		return "", fmt.Errorf("script execution failed: %w", err)
	}
	return fmt.Sprintf("%v", result), nil
}
