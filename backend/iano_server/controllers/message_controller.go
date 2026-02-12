package controllers

import (
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
)

type MessageController struct {
	messageService *services.MessageService
}

func NewMessageController(messageService *services.MessageService) *MessageController {
	return &MessageController{messageService: messageService}
}

type CreateMessageRequest struct {
	SessionID string  `json:"session_id" example:"session-001"`
	Type      string  `json:"type" example:"user"`
	Content   string  `json:"content" example:"你好"`
	Status    string  `json:"status" example:"completed"`
	ParentID  *string `json:"parent_id,omitempty" example:"msg-001"`
}

type UpdateMessageRequest struct {
	Status          *string `json:"status,omitempty" example:"completed"`
	Content         *string `json:"content,omitempty" example:"更新后的内容"`
	InputTokens     *int    `json:"input_tokens,omitempty" example:"100"`
	OutputTokens    *int    `json:"output_tokens,omitempty" example:"200"`
	FeedbackRating  *string `json:"feedback_rating,omitempty" example:"like"`
	FeedbackComment *string `json:"feedback_comment,omitempty" example:"很有帮助"`
}

type AddFeedbackRequest struct {
	Rating  string `json:"rating" example:"like"`
	Comment string `json:"comment" example:"很有帮助"`
}

// Create godoc
// @Summary 创建消息
// @Description 创建一条新消息
// @Tags Message
// @Accept json
// @Produce json
// @Param message body CreateMessageRequest true "消息信息"
// @Success 201 {object} models.Response{data=models.Message}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/messages [post]
func (c *MessageController) Create(ctx *web.Context) {
	var req CreateMessageRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	message := &models.Message{
		SessionID: req.SessionID,
		Type:      models.MessageType(req.Type),
		Content:   req.Content,
		Status:    models.MessageStatus(req.Status),
		ParentID:  req.ParentID,
	}

	message.NewID()

	if err := c.messageService.Create(message); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(message))
}

// GetByID godoc
// @Summary 获取消息详情
// @Description 根据 ID 获取消息详情
// @Tags Message
// @Produce json
// @Param id path string true "消息 ID"
// @Success 200 {object} models.Response{data=models.Message}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/messages/{id} [get]
func (c *MessageController) GetByID(ctx *web.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("id is required"))
		return
	}

	message, err := c.messageService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Message not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(message))
}

// GetAll godoc
// @Summary 获取所有消息
// @Description 获取所有消息列表
// @Tags Message
// @Produce json
// @Success 200 {object} models.Response{data=[]models.Message}
// @Failure 500 {object} models.Response
// @Router /api/messages [get]
func (c *MessageController) GetAll(ctx *web.Context) {
	messages, err := c.messageService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(messages))
}

// GetBySessionID godoc
// @Summary 按会话获取消息
// @Description 根据会话 ID 获取消息列表
// @Tags Message
// @Produce json
// @Param session_id query string true "会话 ID"
// @Success 200 {object} models.Response{data=[]models.Message}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/messages/session [get]
func (c *MessageController) GetBySessionID(ctx *web.Context) {
	sessionIDStr := ctx.Query("session_id")
	if sessionIDStr == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("session_id is required"))
		return
	}

	messages, err := c.messageService.GetBySessionID(sessionIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(messages))
}

// GetByType godoc
// @Summary 按类型获取消息
// @Description 根据类型获取消息列表
// @Tags Message
// @Produce json
// @Param type query string false "消息类型 (user/assistant/system/tool)"
// @Success 200 {object} models.Response{data=[]models.Message}
// @Failure 500 {object} models.Response
// @Router /api/messages/type [get]
func (c *MessageController) GetByType(ctx *web.Context) {
	messageType := ctx.Query("type")
	if messageType == "" {
		c.GetAll(ctx)
		return
	}

	messages, err := c.messageService.GetByType(models.MessageType(messageType))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(messages))
}

// Update godoc
// @Summary 更新消息
// @Description 更新消息信息
// @Tags Message
// @Accept json
// @Produce json
// @Param id path string true "消息 ID"
// @Param message body UpdateMessageRequest true "更新内容"
// @Success 200 {object} models.Response{data=models.Message}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/messages/{id} [put]
func (c *MessageController) Update(ctx *web.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("id is required"))
		return
	}

	var req UpdateMessageRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.InputTokens != nil {
		updates["input_tokens"] = *req.InputTokens
	}
	if req.OutputTokens != nil {
		updates["output_tokens"] = *req.OutputTokens
	}
	if req.FeedbackRating != nil {
		updates["feedback_rating"] = *req.FeedbackRating
	}
	if req.FeedbackComment != nil {
		updates["feedback_comment"] = *req.FeedbackComment
	}

	message, err := c.messageService.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(message))
}

// Delete godoc
// @Summary 删除消息
// @Description 删除指定消息
// @Tags Message
// @Produce json
// @Param id path string true "消息 ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/messages/{id} [delete]
func (c *MessageController) Delete(ctx *web.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("id is required"))
		return
	}

	if err := c.messageService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Message deleted successfully"}))
}

// DeleteBySessionID godoc
// @Summary 删除会话所有消息
// @Description 删除指定会话的所有消息
// @Tags Message
// @Produce json
// @Param session_id query string true "会话 ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/messages [delete]
func (c *MessageController) DeleteBySessionID(ctx *web.Context) {
	sessionID := ctx.Query("session_id")
	if sessionID == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("session_id is required"))
		return
	}

	if err := c.messageService.DeleteBySessionID(sessionID); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Messages deleted successfully"}))
}

// AddFeedback godoc
// @Summary 添加消息反馈
// @Description 为消息添加点赞/点踩反馈
// @Tags Message
// @Accept json
// @Produce json
// @Param id path string true "消息 ID"
// @Param feedback body AddFeedbackRequest true "反馈信息"
// @Success 200 {object} models.Response{data=models.Message}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/messages/{id}/feedback [post]
func (c *MessageController) AddFeedback(ctx *web.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("id is required"))
		return
	}

	var req AddFeedbackRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	message, err := c.messageService.AddFeedback(id, models.FeedbackRating(req.Rating), req.Comment)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(message))
}
