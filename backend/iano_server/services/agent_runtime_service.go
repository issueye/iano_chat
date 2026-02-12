package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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

	return nil
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
	engine := script_engine.NewEngine(&script_engine.Config{
		Timeout: 30 * time.Second,
	})
	result, err := engine.Execute(ctx, tool.ScriptContent, params)
	if err != nil {
		return "", fmt.Errorf("script execution failed: %w", err)
	}
	return fmt.Sprintf("%v", result), nil
}
