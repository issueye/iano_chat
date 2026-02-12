package controllers

import (
	"context"
	"fmt"
	iano "iano_agent"
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

type ChatController struct {
	agentService        *services.AgentService
	providerService     *services.ProviderService
	agentManagerService *services.AgentManagerService
	messageService      *services.MessageService
}

func NewChatController(
	agentService *services.AgentService,
	providerService *services.ProviderService,
	agentManagerService *services.AgentManagerService,
	messageService *services.MessageService,
) *ChatController {
	return &ChatController{
		agentService:        agentService,
		providerService:     providerService,
		agentManagerService: agentManagerService,
		messageService:      messageService,
	}
}

type ChatRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	AgentID   string `json:"agent_id"`
	Message   string `json:"message" validate:"required"`
}

type ChatResponse struct {
	Content    string `json:"content"`
	TokenUsage int64  `json:"token_usage"`
	Duration   int64  `json:"duration_ms"`
	SessionID  string `json:"session_id"`
	AgentID    string `json:"agent_id"`
}

type StreamChatRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	AgentID   string `json:"agent_id"`
	Message   string `json:"message" validate:"required"`
}

func (c *ChatController) getOrCreateAgent(ctx context.Context, agentID string) error {
	_, err := c.agentManagerService.GetAgentInfo(agentID)
	if err == nil {
		return nil
	}

	providers, err := c.providerService.GetAll()
	if err != nil || len(providers) == 0 {
		return fmt.Errorf("请先创建 Provider（AI 提供商配置）")
	}

	provider := providers[0]
	agent := &models.Agent{
		Name:         "Default Agent",
		Description:  "默认智能助手",
		Type:         models.AgentTypeMain,
		ProviderID:   provider.ID,
		Model:        provider.Model,
		Instructions: "你是一个智能助手，请用中文回答用户的问题。",
	}
	agent.ID = agentID
	if err := c.agentService.Create(agent); err != nil {
		return fmt.Errorf("创建默认 Agent 失败: %v", err)
	}

	return c.agentManagerService.ReloadAgent(ctx, agentID)
}

func (c *ChatController) chatWithProvider(ctx context.Context, message string) (string, error) {
	providers, err := c.providerService.GetAll()
	if err != nil || len(providers) == 0 {
		return "", fmt.Errorf("请先创建 Provider（AI 提供商配置）")
	}

	provider := providers[0]
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: provider.BaseUrl,
		APIKey:  provider.ApiKey,
		Model:   provider.Model,
	})
	if err != nil {
		return "", fmt.Errorf("创建 ChatModel 失败: %v", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个智能助手，请用中文回答用户的问题。"),
		schema.UserMessage(message),
	}

	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("调用 AI 失败: %v", err)
	}

	return resp.Content, nil
}

func (c *ChatController) Chat(ctx *web.Context) {
	var req ChatRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	agentID := req.AgentID
	if agentID == "" {
		agentID = "default"
	}

	userMsg := &models.Message{
		SessionID: req.SessionID,
		Type:      models.MessageTypeUser,
		Status:    models.MessageStatusCompleted,
	}
	userMsg.NewID()
	if err := userMsg.SetText(req.Message); err != nil {
		userMsg.Content = req.Message
	}
	c.messageService.Create(userMsg)

	// 加载会话历史记录
	historyMessages, _ := c.messageService.GetBySessionID(req.SessionID)

	var response string
	var err error

	response, err = c.agentManagerService.ChatWithHistory(ctx.Request.Context(), agentID, req.Message, historyMessages, nil)
	if err != nil {
		if cerr := c.getOrCreateAgent(ctx.Request.Context(), agentID); cerr != nil {
			response, err = c.chatWithProvider(ctx.Request.Context(), req.Message)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
				return
			}
		} else {
			response, err = c.agentManagerService.ChatWithHistory(ctx.Request.Context(), agentID, req.Message, historyMessages, nil)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
				return
			}
		}
	}

	assistantMsg := &models.Message{
		SessionID: req.SessionID,
		Type:      models.MessageTypeAssistant,
		Status:    models.MessageStatusCompleted,
	}
	assistantMsg.NewID()
	if err := assistantMsg.SetText(response); err != nil {
		assistantMsg.Content = response
	}
	c.messageService.Create(assistantMsg)

	ctx.JSON(http.StatusOK, models.Success(ChatResponse{
		Content:   response,
		SessionID: req.SessionID,
		AgentID:   agentID,
	}))
}

func (c *ChatController) StreamChat(ctx *web.Context) {
	var req StreamChatRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	sse, err := ctx.SSE()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "SSE not supported: %v", err)
		return
	}

	agentID := req.AgentID
	if agentID == "" {
		agentID = "default"
	}

	userMsg := &models.Message{
		SessionID: req.SessionID,
		Type:      models.MessageTypeUser,
		Status:    models.MessageStatusCompleted,
	}
	userMsg.NewID()
	if err := userMsg.SetText(req.Message); err != nil {
		userMsg.Content = req.Message
	}
	c.messageService.Create(userMsg)

	var accumulatedContent string
	var accumulatedToolCalls []models.ToolCall
	callback := func(content string, isToolCall bool, toolCalls []iano.ToolCallInfo) {
		if content != "" {
			accumulatedContent += content
			sse.EmitEvent("message", map[string]interface{}{
				"content":      content,
				"is_tool_call": isToolCall,
			})
		}
		if len(toolCalls) > 0 {
			for _, tc := range toolCalls {
				toolCall := models.ToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
				accumulatedToolCalls = append(accumulatedToolCalls, toolCall)
				sse.EmitEvent("tool_call", map[string]interface{}{
					"id":        tc.ID,
					"name":      tc.Name,
					"arguments": tc.Arguments,
				})
			}
		}
	}

	// 加载会话历史记录
	historyMessages, _ := c.messageService.GetBySessionID(req.SessionID)

	_, err = c.agentManagerService.ChatWithHistory(ctx.Request.Context(), agentID, req.Message, historyMessages, iano.MessageCallback(callback))
	if err != nil {
		sse.EmitEvent("error", map[string]string{"error": err.Error()})
	} else {
		assistantMsg := &models.Message{
			SessionID: req.SessionID,
			Type:      models.MessageTypeAssistant,
			Status:    models.MessageStatusCompleted,
		}
		assistantMsg.NewID()
		msgContent := &models.MessageContent{
			Text:      accumulatedContent,
			ToolCalls: accumulatedToolCalls,
		}
		if err := assistantMsg.SetContent(msgContent); err != nil {
			assistantMsg.Content = accumulatedContent
		}
		c.messageService.Create(assistantMsg)
	}

	sse.EmitEvent("done", map[string]string{"status": "completed"})
	sse.Close()
}

func (c *ChatController) ClearSession(ctx *web.Context) {
	sessionID := ctx.Param("session_id")
	if sessionID == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("session_id is required"))
		return
	}

	err := c.messageService.DeleteBySessionID(sessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Session cleared successfully"}))
}

func (c *ChatController) GetConversationInfo(ctx *web.Context) {
	sessionID := ctx.Query("session_id")
	agentID := ctx.Query("agent_id")

	if sessionID == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("session_id is required"))
		return
	}

	if agentID == "" {
		agentID = "default"
	}

	info, err := c.agentManagerService.GetAgentInfo(agentID)
	if err != nil {
		ctx.JSON(http.StatusOK, models.Success(map[string]interface{}{
			"sessionId": sessionID,
			"agentId":   agentID,
		}))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(info))
}

func (c *ChatController) GetPoolStats(ctx *web.Context) {
	stats := c.agentManagerService.GetManagerStats()
	ctx.JSON(http.StatusOK, models.Success(stats))
}
