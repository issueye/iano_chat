package controllers

import (
	"encoding/json"
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
)

type SessionController struct {
	sessionService *services.SessionService
}

func NewSessionController(sessionService *services.SessionService) *SessionController {
	return &SessionController{sessionService: sessionService}
}

type CreateSessionRequest struct {
	Title string `json:"title" example:"新会话"`
}

type UpdateSessionRequest struct {
	Title  *string               `json:"title,omitempty" example:"我的会话"`
	Status *string               `json:"status,omitempty" example:"active"`
	Config *models.SessionConfig `json:"config,omitempty"`
}

type UpdateSessionConfigRequest struct {
	Config *models.SessionConfig `json:"config,omitempty"`
}

// Create godoc
// @Summary 创建会话
// @Description 创建一个新的会话
// @Tags Session
// @Accept json
// @Produce json
// @Param session body CreateSessionRequest true "会话信息"
// @Success 201 {object} models.Response{data=models.Session}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/sessions [post]
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

// GetByID godoc
// @Summary 获取会话详情
// @Description 根据 ID 获取会话详情
// @Tags Session
// @Produce json
// @Param id path string true "会话 ID"
// @Success 200 {object} models.Response{data=models.Session}
// @Failure 404 {object} models.Response
// @Router /api/sessions/{id} [get]
func (c *SessionController) GetByID(ctx *web.Context) {
	id := ctx.Param("id")

	session, err := c.sessionService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Session not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(session))
}

// GetAll godoc
// @Summary 获取所有会话
// @Description 获取所有会话列表
// @Tags Session
// @Produce json
// @Success 200 {object} models.Response{data=[]models.Session}
// @Failure 500 {object} models.Response
// @Router /api/sessions [get]
func (c *SessionController) GetAll(ctx *web.Context) {
	sessions, err := c.sessionService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(sessions))
}

// GetByStatus godoc
// @Summary 按状态获取会话
// @Description 根据状态获取会话列表
// @Tags Session
// @Produce json
// @Param status query string false "会话状态 (active/paused/completed/archived)"
// @Success 200 {object} models.Response{data=[]models.Session}
// @Failure 500 {object} models.Response
// @Router /api/sessions/status [get]
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

// Update godoc
// @Summary 更新会话
// @Description 更新会话信息
// @Tags Session
// @Accept json
// @Produce json
// @Param id path string true "会话 ID"
// @Param session body UpdateSessionRequest true "更新内容"
// @Success 200 {object} models.Response{data=models.Session}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/sessions/{id} [put]
func (c *SessionController) Update(ctx *web.Context) {
	id := ctx.Param("id")

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

// UpdateConfig godoc
// @Summary 更新会话配置
// @Description 更新指定会话的配置
// @Tags Session
// @Accept json
// @Produce json
// @Param id path string true "会话 ID"
// @Param config body UpdateSessionConfigRequest true "配置信息"
// @Success 200 {object} models.Response{data=models.Session}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/sessions/{id}/config [put]
func (c *SessionController) UpdateConfig(ctx *web.Context) {
	id := ctx.Param("id")

	var req UpdateSessionConfigRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	if req.Config == nil {
		ctx.JSON(http.StatusBadRequest, models.Fail("config is required"))
		return
	}

	session, err := c.sessionService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Session not found"))
		return
	}

	currentConfig, err := session.GetConfig()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	if req.Config.ModelID != 0 {
		currentConfig.ModelID = req.Config.ModelID
	}
	if req.Config.SystemPrompt != "" {
		currentConfig.SystemPrompt = req.Config.SystemPrompt
	}
	if req.Config.Temperature > 0 {
		currentConfig.Temperature = req.Config.Temperature
	}
	if req.Config.MaxTokens > 0 {
		currentConfig.MaxTokens = req.Config.MaxTokens
	}
	if req.Config.EnableTools {
		currentConfig.EnableTools = req.Config.EnableTools
	}
	if req.Config.EnableSummary {
		currentConfig.EnableSummary = req.Config.EnableSummary
	}
	if req.Config.EnableRateLimit {
		currentConfig.EnableRateLimit = req.Config.EnableRateLimit
	}
	if req.Config.RateLimitRPM > 0 {
		currentConfig.RateLimitRPM = req.Config.RateLimitRPM
	}
	if req.Config.KeepRounds > 0 {
		currentConfig.KeepRounds = req.Config.KeepRounds
	}
	if len(req.Config.SelectedTools) > 0 {
		currentConfig.SelectedTools = req.Config.SelectedTools
	}

	if err := session.SetConfig(currentConfig); err != nil {
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

// Delete godoc
// @Summary 删除会话
// @Description 删除指定会话
// @Tags Session
// @Produce json
// @Param id path string true "会话 ID"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/sessions/{id} [delete]
func (c *SessionController) Delete(ctx *web.Context) {
	id := ctx.Param("id")

	if err := c.sessionService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Session deleted successfully"}))
}

// GetConfig godoc
// @Summary 获取会话配置
// @Description 获取指定会话的配置
// @Tags Session
// @Produce json
// @Param id path string true "会话 ID"
// @Success 200 {object} models.Response{data=models.SessionConfig}
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/sessions/{id}/config [get]
func (c *SessionController) GetConfig(ctx *web.Context) {
	id := ctx.Param("id")

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
