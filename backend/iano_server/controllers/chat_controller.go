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
	agentSSEClientMap   *services.AgentSSEClientMap
}

func NewChatController(
	agentService *services.AgentService,
	providerService *services.ProviderService,
	messageService *services.MessageService,
	agentRuntimeService *services.AgentRuntimeService,
	agentSSEClientMap *services.AgentSSEClientMap,
) *ChatController {
	return &ChatController{
		agentService:        agentService,
		providerService:     providerService,
		messageService:      messageService,
		agentRuntimeService: agentRuntimeService,
		agentSSEClientMap:   agentSSEClientMap,
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

	// 检查会话是否存在
	if !c.agentSSEClientMap.CheckSession(req.SessionID) {
		c.agentSSEClientMap.CreateSession(req.SessionID)
	}

	agentID := req.AgentID
	if agentID == "" {
		agentID = "default"
	}

	// 检查 Agent 是否已绑定
	if !c.agentSSEClientMap.CheckAgent(req.SessionID) {
		// 创建用户消息
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

		// 创建助手消息
		assistantMsg := &models.Message{
			SessionID: req.SessionID,
			Type:      models.MessageTypeAssistant,
			Status:    models.MessageStatusCompleted,
		}
		assistantMsg.NewID()
		c.messageService.Create(assistantMsg)

		sse.EmitDataToID(req.SessionID, models.MessageEventCreated.ToString(), map[string]interface{}{
			"type":       models.MessageTypeUser.ToString(),
			"id":         userMsg.ID,
			"session_id": userMsg.SessionID,
			"content":    userMsg.Content,
			"created_at": userMsg.CreatedAt,
		})

		// 获取 Agent 实例
		agentParams := &services.AgentParams{
			AgentID:  agentID,
			WorkDir:  req.WorkDir,
			Callback: Callback(req.SessionID, sse),
		}
		agent, err := c.agentRuntimeService.GetAgent(ctx.Request.Context(), agentParams)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
			return
		}
		c.agentSSEClientMap.AddAgent(req.SessionID, agent)

		// 从数据库加载历史消息
		historyMessages, err := c.messageService.GetBySessionID(req.SessionID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
			return
		}

		// 转换为 schema.Message 格式
		chatMessages := make([]*schema.Message, 0, len(historyMessages))
		for _, msg := range historyMessages {
			switch msg.Type {
			case models.MessageTypeUser:
				chatMessages = append(chatMessages, schema.UserMessage(msg.Content))
			case models.MessageTypeAssistant:
				chatMessages = append(chatMessages, schema.AssistantMessage(msg.Content, nil))
			case models.MessageTypeTool:
				chatMessages = append(chatMessages, schema.ToolMessage(msg.Content, msg.Content))
			case models.MessageTypeSystem:
				chatMessages = append(chatMessages, schema.SystemMessage(msg.Content))
			}
		}

		// 调用 Agent 进行聊天
		_, err = agent.Chat(ctx.Request.Context(), chatMessages)
		if err != nil {
			errSend := models.CreateErrCompleted(req.SessionID, models.MessageStatusFailed, err.Error())
			sse.EmitDataToID(req.SessionID, models.MessageEventCompleted.ToString(), errSend)
		}

		sse.EmitDataToID(req.SessionID, models.MessageEventCompleted.ToString(), map[string]string{"status": "completed"})
	} else {
		// Agent 已绑定，继续聊天
		agent := c.agentSSEClientMap.GetSessionAgent(req.SessionID)
		if agent == nil {
			ctx.JSON(http.StatusInternalServerError, models.Fail("会话不存在"))
			return
		}

		agent.Agent.AppendCB(Callback(req.SessionID, sse))
		agent.Agent.WaitForResponse()
	}

	sse.EmitDataToID(req.SessionID, models.MessageEventDone.ToString(), map[string]string{"status": "completed"})
	sse.Close()
}

func Callback(sessionID string, sse *web.SSEContext) func(msg *iano.Message) {
	return func(msg *iano.Message) {
		sse.EmitDataToID(sessionID, models.MessageEventContent.ToString(), msg)
	}
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
