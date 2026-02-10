package controllers

import (
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
	"strconv"
)

type MessageController struct {
	messageService *services.MessageService
}

func NewMessageController(messageService *services.MessageService) *MessageController {
	return &MessageController{messageService: messageService}
}

type CreateMessageRequest struct {
	SessionID int64   `json:"session_id"`
	UserID    int64   `json:"user_id"`
	Type      string  `json:"type"`
	Content   string  `json:"content"`
	Status    string  `json:"status"`
	ParentID  *string `json:"parent_id,omitempty"`
}

type UpdateMessageRequest struct {
	Status          *string `json:"status,omitempty"`
	Content         *string `json:"content,omitempty"`
	InputTokens     *int    `json:"input_tokens,omitempty"`
	OutputTokens    *int    `json:"output_tokens,omitempty"`
	FeedbackRating  *string `json:"feedback_rating,omitempty"`
	FeedbackComment *string `json:"feedback_comment,omitempty"`
}

type AddFeedbackRequest struct {
	Rating  string `json:"rating"`
	Comment string `json:"comment"`
}

func (c *MessageController) Create(ctx *web.Context) {
	var req CreateMessageRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	message := &models.Message{
		SessionID: req.SessionID,
		UserID:    req.UserID,
		Type:      models.MessageType(req.Type),
		Content:   req.Content,
		Status:    models.MessageStatus(req.Status),
		ParentID:  req.ParentID,
	}

	if err := c.messageService.Create(message); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(message))
}

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

func (c *MessageController) GetAll(ctx *web.Context) {
	messages, err := c.messageService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(messages))
}

func (c *MessageController) GetBySessionID(ctx *web.Context) {
	sessionIDStr := ctx.Query("session_id")
	if sessionIDStr == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("session_id is required"))
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid session_id"))
		return
	}

	messages, err := c.messageService.GetBySessionID(sessionID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(messages))
}

func (c *MessageController) GetByUserID(ctx *web.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("user_id is required"))
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid user_id"))
		return
	}

	messages, err := c.messageService.GetByUserID(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(messages))
}

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

func (c *MessageController) DeleteBySessionID(ctx *web.Context) {
	sessionIDStr := ctx.Query("session_id")
	if sessionIDStr == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("session_id is required"))
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid session_id"))
		return
	}

	if err := c.messageService.DeleteBySessionID(sessionID); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Messages deleted successfully"}))
}

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
