package controllers

import (
	"context"
	"encoding/json"
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
	agentRuntimeService *services.AgentRuntimeService,
) *ChatController {
	return &ChatController{
		agentService:        agentService,
		providerService:     providerService,
		messageService:      messageService,
		agentRuntimeService: agentRuntimeService,
	}
}

type ChatRequest struct {
	SessionID string `json:"session_id" validate:"required" example:"session-001"`
	AgentID   string `json:"agent_id" example:"default"`
	Message   string `json:"message" validate:"required" example:"你好"`
}

type ChatResponse struct {
	Content    string `json:"content" example:"你好！有什么可以帮助你的吗？"`
	TokenUsage int64  `json:"token_usage" example:"150"`
	Duration   int64  `json:"duration_ms" example:"1234"`
	SessionID  string `json:"session_id" example:"session-001"`
	AgentID    string `json:"agent_id" example:"default"`
}

type StreamChatRequest struct {
	SessionID string `json:"session_id" validate:"required" example:"session-001"`
	AgentID   string `json:"agent_id" example:"default"`
	Message   string `json:"message" validate:"required" example:"你好"`
	WorkDir   string `json:"work_dir" example:"E:\\codes\\project"`
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

// StreamChat godoc
// @Summary 流式聊天
// @Description 与 AI 进行流式对话，支持工具调用
// @Tags Chat
// @Accept json
// @Produce text/event-stream
// @Param request body StreamChatRequest true "聊天请求"
// @Success 200 {string} string "SSE 流式响应"
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/chat/stream [post]
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
	err = userMsg.SetText(req.Message)
	if err == nil {
		userMsg.Content = req.Message
	}
	c.messageService.Create(userMsg)

	sse.EmitEvent("message_created", map[string]interface{}{
		"type":       "user",
		"id":         userMsg.ID,
		"session_id": userMsg.SessionID,
		"content":    userMsg.Content,
		"created_at": userMsg.CreatedAt,
	})

	assistantMsg := &models.Message{
		SessionID: req.SessionID,
		Type:      models.MessageTypeAssistant,
		Status:    models.MessageStatusStreaming,
	}
	assistantMsg.NewID()

	sse.EmitEvent("message_created", map[string]interface{}{
		"type":       "assistant",
		"id":         assistantMsg.ID,
		"session_id": assistantMsg.SessionID,
		"content":    JSONString(map[string]interface{}{"blocks": []interface{}{}, "text": "", "tool_calls": []interface{}{}}),
		"status":     "streaming",
		"created_at": assistantMsg.CreatedAt,
	})

	var accumulatedContent string
	var accumulatedToolCalls []models.ToolCall
	var contentBlocks []models.ContentBlock
	var accumulatedReasoning string
	callback := func(content string, isToolCall bool, toolCalls *iano.ToolCallInfo, reasoning string) {
		if reasoning != "" {
			accumulatedReasoning = reasoning
			sse.EmitEvent("reasoning", map[string]interface{}{
				"reasoning": reasoning,
			})
		}

		if isToolCall && toolCalls != nil {
			toolCall := models.ToolCall{
				ID:   toolCalls.ID,
				Type: "function",
				Function: models.Function{
					Name:      toolCalls.Name,
					Arguments: toolCalls.Arguments,
				},
			}
			accumulatedToolCalls = append(accumulatedToolCalls, toolCall)

			contentBlocks = append(contentBlocks, models.ContentBlock{
				Type:     "tool_call",
				ToolCall: &toolCall,
			})

			sse.EmitEvent("content_block", map[string]interface{}{
				"type": "tool_call",
				"tool_call": map[string]interface{}{
					"id":        toolCalls.ID,
					"name":      toolCalls.Name,
					"arguments": toolCalls.Arguments,
				},
			})
		}

		if content != "" {
			accumulatedContent += content

			if len(contentBlocks) > 0 && contentBlocks[len(contentBlocks)-1].Type == "text" {
				contentBlocks[len(contentBlocks)-1].Text += content
			} else {
				contentBlocks = append(contentBlocks, models.ContentBlock{
					Type: "text",
					Text: content,
				})
			}

			sse.EmitEvent("content_block", map[string]interface{}{
				"type": "text",
				"text": content,
			})
		}
	}

	historyMessages, err := c.messageService.GetBySessionID(req.SessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	agent, err := c.agentRuntimeService.GetAgent(ctx.Request.Context(), agentID, req.WorkDir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

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
		assistantMsg.Status = models.MessageStatusFailed
		assistantMsg.SetContent(&models.MessageContent{
			Blocks:           contentBlocks,
			Text:             accumulatedContent,
			ToolCalls:        accumulatedToolCalls,
			ReasoningContent: accumulatedReasoning,
		})
		c.messageService.Create(assistantMsg)

		sse.EmitEvent("message_completed", map[string]interface{}{
			"id":     assistantMsg.ID,
			"status": "failed",
			"error":  err.Error(),
		})
		sse.EmitEvent("error", map[string]string{"error": err.Error()})
	} else {
		assistantMsg.Status = models.MessageStatusCompleted
		msgContent := &models.MessageContent{
			Blocks:           contentBlocks,
			Text:             accumulatedContent,
			ToolCalls:        accumulatedToolCalls,
			ReasoningContent: accumulatedReasoning,
		}
		if err := assistantMsg.SetContent(msgContent); err != nil {
			assistantMsg.Content = accumulatedContent
		}
		c.messageService.Create(assistantMsg)

		sse.EmitEvent("message_completed", map[string]interface{}{
			"id":      assistantMsg.ID,
			"status":  "completed",
			"content": assistantMsg.Content,
		})
	}

	sse.EmitEvent("done", map[string]string{"status": "completed"})
	sse.Close()
}

func JSONString(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// ClearSession godoc
// @Summary 清空会话消息
// @Description 清空指定会话的所有消息
// @Tags Chat
// @Produce json
// @Param session_id path string true "会话 ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/chat/session/{session_id} [delete]
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
