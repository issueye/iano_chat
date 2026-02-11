package controllers

import (
	iano "iano_agent"
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
)

type ChatController struct {
	chatService     *services.ChatService
	agentService    *services.AgentService
	providerService *services.ProviderService
}

func NewChatController(chatService *services.ChatService, agentService *services.AgentService, providerService *services.ProviderService) *ChatController {
	return &ChatController{
		chatService:     chatService,
		agentService:    agentService,
		providerService: providerService,
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

	agent, err := c.agentService.GetByID(agentID)
	if err != nil {
		agent = &models.Agent{
			Name:         "Default Agent",
			Instructions: "你是一个智能助手。",
		}
	}

	chatReq := &services.ChatRequest{
		SessionID: req.SessionID,
		AgentID:   agentID,
		Message:   req.Message,
		Agent:     agent,
	}

	resp, err := c.chatService.Chat(ctx.Request.Context(), chatReq, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	userMsg := &models.Message{
		SessionID: req.SessionID,
		Type:      models.MessageTypeUser,
		Content:   req.Message,
		Status:    models.MessageStatusCompleted,
	}
	userMsg.NewID()
	c.chatService.SaveMessage(userMsg)

	assistantMsg := &models.Message{
		SessionID:   req.SessionID,
		Type:        models.MessageTypeAssistant,
		Content:     resp.Content,
		Status:      models.MessageStatusCompleted,
		InputTokens: int(resp.TokenUsage.PromptTokens),
	}
	assistantMsg.NewID()
	c.chatService.SaveMessage(assistantMsg)

	ctx.JSON(http.StatusOK, models.Success(ChatResponse{
		Content:    resp.Content,
		TokenUsage: resp.TokenUsage.TotalTokens,
		Duration:   resp.Duration.Milliseconds(),
		SessionID:  req.SessionID,
		AgentID:    agentID,
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

	callback := func(content string, isToolCall bool) {
		sse.EmitEvent("message", map[string]interface{}{
			"content":      content,
			"is_tool_call": isToolCall,
		})
	}

	chatReq := &services.ChatRequest{
		SessionID: req.SessionID,
		AgentID:   agentID,
		Message:   req.Message,
	}

	_, err = c.chatService.Chat(ctx.Request.Context(), chatReq, iano.MessageCallback(callback))
	if err != nil {
		sse.EmitEvent("error", map[string]string{"error": err.Error()})
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

	if err := c.chatService.ClearSession(ctx.Request.Context(), sessionID); err != nil {
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

	info, err := c.chatService.GetConversationInfo(sessionID, agentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(info))
}

func (c *ChatController) GetPoolStats(ctx *web.Context) {
	stats := c.chatService.GetPoolStats()
	ctx.JSON(http.StatusOK, models.Success(stats))
}
