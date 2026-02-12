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
	messageService      *services.MessageService
	agentRuntimeService *services.AgentRuntimeService
}

func NewChatController(
	agentService *services.AgentService,
	providerService *services.ProviderService,
	messageService *services.MessageService,
) *ChatController {
	return &ChatController{
		agentService:    agentService,
		providerService: providerService,
		messageService:  messageService,
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
	SessionID string `json:"session_id" validate:"required"` // 会话 ID，用于关联消息
	AgentID   string `json:"agent_id"`                       // 可选，默认使用 "default"
	Message   string `json:"message" validate:"required"`    // 用户输入的消息
	WorkDir   string `json:"work_dir"`                       // 用户选择的工作目录，可能是项目目录
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
	callback := func(content string, isToolCall bool, toolCalls *iano.ToolCallInfo) {
		if content != "" {
			accumulatedContent += content
			sse.EmitEvent("message", map[string]interface{}{
				"content":      content,
				"is_tool_call": isToolCall,
			})
		}
	}

	// 加载会话历史记录
	historyMessages, err := c.messageService.GetBySessionID(req.SessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	agent, err := c.agentRuntimeService.GetAgent(ctx.Request.Context(), agentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	// 转换聊天内容为 schema.Message 格式
	chatMessages := make([]*schema.Message, 0, len(historyMessages))
	for _, msg := range historyMessages {
		switch msg.Type {
		case models.MessageTypeUser:
			chatMessages = append(chatMessages, schema.UserMessage(msg.Content))
		case models.MessageTypeAssistant:
			chatMessages = append(chatMessages, schema.AssistantMessage(msg.Content, nil))
		case models.MessageTypeTool:
			chatMessages = append(chatMessages, schema.ToolMessage(msg.Content, msg.Content))
		}
	}

	_, err = agent.Chat(ctx.Request.Context(), chatMessages, callback)
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
