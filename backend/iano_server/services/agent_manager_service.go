package services

import (
	"context"
	"encoding/json"
	"fmt"
	iano "iano_agent"
	script_engine "iano_script_engine"
	"iano_server/models"
	"log/slog"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
)

type AgentManagerService struct {
	db              *gorm.DB
	agentService    *AgentService
	providerService *ProviderService
	toolService     *ToolService
	manager         *iano.Manager
	modelCache      map[string]model.ToolCallingChatModel
}

func NewAgentManagerService(
	db *gorm.DB,
	agentService *AgentService,
	providerService *ProviderService,
	toolService *ToolService,
) *AgentManagerService {
	return &AgentManagerService{
		db:              db,
		agentService:    agentService,
		providerService: providerService,
		toolService:     toolService,
		modelCache:      make(map[string]model.ToolCallingChatModel),
	}
}

func (s *AgentManagerService) Initialize(ctx context.Context) error {
	s.manager = iano.NewManager(nil)
	slog.Info("Agent manager created")

	s.ensureDefaultAgents(ctx)

	agents, err := s.agentService.GetAll()
	if err != nil {
		return fmt.Errorf("failed to load agents: %w", err)
	}

	for _, agent := range agents {
		if err := s.loadAgent(ctx, &agent); err != nil {
			slog.Error("Failed to load agent", "id", agent.ID, "error", err)
		}
	}

	slog.Info("Agent manager initialized", "count", len(agents))
	return nil
}

func (s *AgentManagerService) ensureDefaultAgents(ctx context.Context) {
	defaultAgents := []struct {
		ID           string
		Name         string
		Description  string
		Instructions string
		Type         models.AgentType
	}{
		{
			ID:          "plan",
			Name:        "Plan Agent",
			Description: "负责分析和规划任务，将复杂问题分解为可执行的步骤",
			Instructions: `你是一个专业的任务规划助手。你的职责是：
1. 分析用户的需求和目标
2. 将复杂任务分解为清晰、可执行的步骤
3. 评估每个步骤的优先级和依赖关系
4. 提供结构化的执行计划

请用中文回复，保持简洁明了。`,
			Type: models.AgentTypeMain,
		},
		{
			ID:          "build",
			Name:        "Build Agent",
			Description: "负责执行具体任务，包括代码编写、文件操作等实际工作",
			Instructions: `你是一个专业的任务执行助手。你的职责是：
1. 根据规划执行具体的任务步骤
2. 编写代码、创建文件、修改配置等
3. 解决执行过程中遇到的技术问题
4. 验证执行结果是否符合预期

请用中文回复，提供详细的执行说明。`,
			Type: models.AgentTypeMain,
		},
	}

	for _, cfg := range defaultAgents {
		existing, err := s.agentService.GetByID(cfg.ID)
		if err == nil && existing != nil {
			continue
		}

		agent := &models.Agent{
			Name:         cfg.Name,
			Description:  cfg.Description,
			Instructions: cfg.Instructions,
			Type:         cfg.Type,
			IsSubAgent:   false,
		}
		agent.ID = cfg.ID

		if err := s.agentService.Create(agent); err != nil {
			slog.Error("Failed to create default agent", "id", cfg.ID, "error", err)
		} else {
			slog.Info("Created default agent", "id", cfg.ID, "name", cfg.Name)
		}
	}
}

func (s *AgentManagerService) loadAgent(ctx context.Context, agent *models.Agent) error {
	chatModel, err := s.getOrCreateChatModel(ctx, agent.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to get chat model: %w", err)
	}

	allowedTools := s.parseTools(agent.Tools)

	cfg := &iano.CreateAgentConfig{
		ID:           agent.ID,
		Name:         agent.Name,
		SystemPrompt: agent.Instructions,
		AllowedTools: allowedTools,
		MaxRounds:    50,
		Model:        chatModel,
	}

	_, err = s.manager.CreateAgent(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create agent instance: %w", err)
	}

	return nil
}

func (s *AgentManagerService) getOrCreateChatModel(ctx context.Context, providerID string) (model.ToolCallingChatModel, error) {
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
		return nil, fmt.Errorf("no default provider configured, please set a provider as default")
	}

	if m, exists := s.modelCache[defaultProvider.ID]; exists {
		return m, nil
	}

	return s.createChatModelFromProvider(ctx, defaultProvider)
}

func (s *AgentManagerService) createChatModelFromProvider(ctx context.Context, provider *models.Provider) (model.ToolCallingChatModel, error) {
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

func (s *AgentManagerService) parseTools(toolsStr string) []string {
	if toolsStr == "" {
		return nil
	}

	var tools []string
	if err := json.Unmarshal([]byte(toolsStr), &tools); err != nil {
		return []string{toolsStr}
	}
	return tools
}

func (s *AgentManagerService) CreateAgent(ctx context.Context, req *CreateAgentRequest) (*models.Agent, error) {
	agent := &models.Agent{
		Name:         req.Name,
		Description:  req.Description,
		Type:         models.AgentType(req.Type),
		IsSubAgent:   req.IsSubAgent,
		ProviderID:   req.ProviderID,
		Model:        req.Model,
		Instructions: req.Instructions,
		Tools:        req.Tools,
	}
	agent.NewID()

	if err := s.agentService.Create(agent); err != nil {
		return nil, fmt.Errorf("failed to create agent in database: %w", err)
	}

	if err := s.loadAgent(ctx, agent); err != nil {
		slog.Error("Failed to load agent instance", "id", agent.ID, "error", err)
	}

	return agent, nil
}

type CreateAgentRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	IsSubAgent   bool   `json:"is_sub_agent"`
	ProviderID   string `json:"provider_id"`
	Model        string `json:"model"`
	Instructions string `json:"instructions"`
	Tools        string `json:"tools"`
}

func (s *AgentManagerService) DeleteAgent(agentID string) error {
	if s.manager != nil {
		if err := s.manager.DeleteAgent(agentID); err != nil {
			slog.Warn("Failed to delete agent instance", "id", agentID, "error", err)
		}
	}

	return s.agentService.Delete(agentID)
}

func (s *AgentManagerService) Chat(ctx context.Context, agentID string, message string, callback iano.MessageCallback) (string, error) {
	if s.manager == nil {
		return "", fmt.Errorf("agent manager not initialized")
	}

	if _, exists := s.manager.GetAgent(agentID); !exists {
		chatModel, err := s.getOrCreateChatModel(ctx, "")
		if err != nil {
			return "", fmt.Errorf("failed to get chat model for agent %s: %w", agentID, err)
		}

		cfg := &iano.CreateAgentConfig{
			ID:        agentID,
			Name:      agentID,
			MaxRounds: 50,
			Model:     chatModel,
		}

		if _, err := s.manager.CreateAgent(ctx, cfg); err != nil {
			return "", fmt.Errorf("failed to create temporary agent: %w", err)
		}
	}

	return s.manager.Chat(ctx, agentID, message, callback)
}

// ChatWithHistory 使用历史消息进行对话
func (s *AgentManagerService) ChatWithHistory(ctx context.Context, agentID string, message string, history []models.Message, callback iano.MessageCallback) (string, error) {
	if s.manager == nil {
		return "", fmt.Errorf("agent manager not initialized")
	}

	if _, exists := s.manager.GetAgent(agentID); !exists {
		chatModel, err := s.getOrCreateChatModel(ctx, "")
		if err != nil {
			return "", fmt.Errorf("failed to get chat model for agent %s: %w", agentID, err)
		}

		cfg := &iano.CreateAgentConfig{
			ID:        agentID,
			Name:      agentID,
			MaxRounds: 50,
			Model:     chatModel,
		}

		if _, err := s.manager.CreateAgent(ctx, cfg); err != nil {
			return "", fmt.Errorf("failed to create temporary agent: %w", err)
		}
	}

	// 获取 Agent 实例
	instance, exists := s.manager.GetAgent(agentID)
	if !exists {
		return "", fmt.Errorf("agent with id %s not found", agentID)
	}

	// 将历史消息加载到 Agent 的对话中
	if len(history) > 0 {
		instance.Agent.LoadConversationHistory(ctx, convertToConversationRounds(history))
	}

	return s.manager.Chat(ctx, agentID, message, callback)
}

// convertToConversationRounds 将数据库消息转换为对话轮次
func convertToConversationRounds(messages []models.Message) []*iano.ConversationRound {
	rounds := make([]*iano.ConversationRound, 0)
	var currentRound *iano.ConversationRound

	for _, msg := range messages {
		content, _ := msg.GetContent()
		if content == nil {
			continue
		}

		switch msg.Type {
		case models.MessageTypeUser:
			if currentRound != nil {
				rounds = append(rounds, currentRound)
			}
			currentRound = &iano.ConversationRound{
				UserMessage: schema.UserMessage(content.Text),
				Timestamp:   msg.CreatedAt,
			}
		case models.MessageTypeAssistant:
			if currentRound != nil {
				// 处理工具调用
				var toolCalls []schema.ToolCall
				for _, tc := range content.ToolCalls {
					toolCalls = append(toolCalls, schema.ToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						Function: schema.FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					})
				}
				currentRound.AssistantMessage = schema.AssistantMessage(content.Text, toolCalls)
				currentRound.TokenCount = estimateTokens(content.Text)
			}
		}
	}

	if currentRound != nil {
		rounds = append(rounds, currentRound)
	}

	return rounds
}

// estimateTokens 估算 token 数量（简单实现）
func estimateTokens(text string) int {
	// 简单估算：中文字符算 2 个 token，英文单词算 1 个 token
	count := 0
	for _, r := range text {
		if r > 127 {
			count += 2
		} else if r != ' ' && r != '\n' && r != '\t' {
			count++
		}
	}
	return count
}

func (s *AgentManagerService) GetAgentInfo(agentID string) (map[string]interface{}, error) {
	if s.manager == nil {
		return nil, fmt.Errorf("agent manager not initialized")
	}

	return s.manager.GetAgentInfo(agentID)
}

func (s *AgentManagerService) AddToolToAgent(ctx context.Context, agentID string, toolID string) error {
	if s.manager == nil {
		return fmt.Errorf("agent manager not initialized")
	}

	tool, err := s.toolService.GetByID(toolID)
	if err != nil {
		return fmt.Errorf("tool not found: %w", err)
	}

	dynamicTool, err := s.createDynamicTool(tool)
	if err != nil {
		return fmt.Errorf("failed to create dynamic tool: %w", err)
	}

	return s.manager.AddToolToAgent(ctx, agentID, tool.Name, dynamicTool)
}

func (s *AgentManagerService) RemoveToolFromAgent(agentID string, toolName string) error {
	if s.manager == nil {
		return fmt.Errorf("agent manager not initialized")
	}

	return s.manager.RemoveToolFromAgent(agentID, toolName)
}

func (s *AgentManagerService) createDynamicTool(tool *models.Tool) (*iano.DynamicTool, error) {
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

func (s *AgentManagerService) createToolHandler(tool *models.Tool) iano.DynamicToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (string, error) {
		s.toolService.IncrementCallCount(tool.ID)

		switch tool.Type {
		case models.ToolTypeBuiltin:
			return s.handleBuiltinTool(ctx, tool, params)
		case models.ToolTypeCustom:
			return s.handleCustomTool(ctx, tool, params)
		case models.ToolTypeExternal:
			return s.handleExternalTool(ctx, tool, params)
		case models.ToolTypeScript:
			return s.handleScriptTool(ctx, tool, params)
		default:
			return fmt.Sprintf("Tool %s executed with params: %v", tool.Name, params), nil
		}
	}
}

func (s *AgentManagerService) handleBuiltinTool(ctx context.Context, tool *models.Tool, params map[string]interface{}) (string, error) {
	return fmt.Sprintf("Builtin tool %s executed", tool.Name), nil
}

func (s *AgentManagerService) handleCustomTool(ctx context.Context, tool *models.Tool, params map[string]interface{}) (string, error) {
	return fmt.Sprintf("Custom tool %s executed with params: %v", tool.Name, params), nil
}

func (s *AgentManagerService) handleExternalTool(ctx context.Context, tool *models.Tool, params map[string]interface{}) (string, error) {
	return fmt.Sprintf("External tool %s called", tool.Name), nil
}

func (s *AgentManagerService) handleScriptTool(ctx context.Context, tool *models.Tool, params map[string]interface{}) (string, error) {
	engine := script_engine.NewEngine(nil)
	return engine.ExecuteWithStringResult(ctx, tool.Config, params)
}

func (s *AgentManagerService) ListAgentInstances() []*iano.AgentInstance {
	if s.manager == nil {
		return nil
	}
	return s.manager.ListAgents()
}

func (s *AgentManagerService) GetManagerStats() map[string]interface{} {
	if s.manager == nil {
		return map[string]interface{}{
			"initialized": false,
		}
	}
	return map[string]interface{}{
		"initialized":   true,
		"agentCount":    s.manager.Count(),
		"modelCacheLen": len(s.modelCache),
	}
}

func (s *AgentManagerService) ReloadAgent(ctx context.Context, agentID string) error {
	agent, err := s.agentService.GetByID(agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	if s.manager != nil {
		s.manager.DeleteAgent(agentID)
	}

	return s.loadAgent(ctx, agent)
}

func (s *AgentManagerService) ClearAllAgents() {
	if s.manager != nil {
		s.manager.Clear()
	}
}
