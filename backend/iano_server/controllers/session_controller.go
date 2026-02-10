package controllers

import (
	"encoding/json"
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
	"strconv"
)

type SessionController struct {
	sessionService *services.SessionService
}

func NewSessionController(sessionService *services.SessionService) *SessionController {
	return &SessionController{sessionService: sessionService}
}

type CreateSessionRequest struct {
	KeyID string `json:"key_id"`
	Title string `json:"title"`
}

type UpdateSessionRequest struct {
	Title  *string               `json:"title,omitempty"`
	Status *string               `json:"status,omitempty"`
	Config *models.SessionConfig `json:"config,omitempty"`
}

type UpdateSessionConfigRequest struct {
	Config models.SessionConfig `json:"config"`
}

func (c *SessionController) Create(ctx *web.Context) {
	var req CreateSessionRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	session := &models.Session{
		Title:  req.Title,
		Status: models.SessionStatusActive,
	}

	session.NewID()

	if err := c.sessionService.Create(session); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(session))
}

func (c *SessionController) GetByID(ctx *web.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid id"))
		return
	}

	session, err := c.sessionService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Session not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(session))
}

func (c *SessionController) GetAll(ctx *web.Context) {
	sessions, err := c.sessionService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(sessions))
}

func (c *SessionController) GetByKeyID(ctx *web.Context) {
	keyIDStr := ctx.Query("key_id")
	if keyIDStr == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("key_id is required"))
		return
	}

	sessions, err := c.sessionService.GetByKeyID(keyIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(sessions))
}

func (c *SessionController) GetByStatus(ctx *web.Context) {
	status := ctx.Query("status")
	if status == "" {
		c.GetAll(ctx)
		return
	}

	sessions, err := c.sessionService.GetByStatus(models.SessionStatus(status))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(sessions))
}

func (c *SessionController) Update(ctx *web.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid id"))
		return
	}

	var req UpdateSessionRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Config != nil {
		configData, err := json.Marshal(req.Config)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, models.Fail("invalid config"))
			return
		}
		updates["config_json"] = string(configData)
	}

	session, err := c.sessionService.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(session))
}

func (c *SessionController) UpdateConfig(ctx *web.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid id"))
		return
	}

	var req UpdateSessionConfigRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	session, err := c.sessionService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Session not found"))
		return
	}

	if err := session.SetConfig(&req.Config); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	updates := map[string]interface{}{
		"config_json": session.ConfigJSON,
	}

	updatedSession, err := c.sessionService.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(updatedSession))
}

func (c *SessionController) Delete(ctx *web.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid id"))
		return
	}

	if err := c.sessionService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Session deleted successfully"}))
}

func (c *SessionController) DeleteByKeyID(ctx *web.Context) {
	keyIDStr := ctx.Query("key_id")
	if keyIDStr == "" {
		ctx.JSON(http.StatusBadRequest, models.Fail("key_id is required"))
		return
	}

	if err := c.sessionService.DeleteByKeyID(keyIDStr); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Sessions deleted successfully"}))
}

func (c *SessionController) GetConfig(ctx *web.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("invalid id"))
		return
	}

	session, err := c.sessionService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Session not found"))
		return
	}

	config, err := session.GetConfig()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(config))
}
