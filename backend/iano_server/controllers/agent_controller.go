package controllers

import (
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
	"strings"
)

type AgentController struct {
	agentService        *services.AgentService
	agentRuntimeService *services.AgentRuntimeService
}

func NewAgentController(agentService *services.AgentService, runtimeService *services.AgentRuntimeService) *AgentController {
	return &AgentController{
		agentService:        agentService,
		agentRuntimeService: runtimeService,
	}
}

type CreateAgentRequest struct {
	Name         string `json:"name" example:"助手"`
	Description  string `json:"description" example:"通用助手"`
	Type         string `json:"type" example:"main"`
	IsSubAgent   bool   `json:"is_sub_agent" example:"false"`
	ProviderID   string `json:"provider_id" example:"provider-001"`
	Model        string `json:"model" example:"gpt-4"`
	Instructions string `json:"instructions" example:"你是一个智能助手"`
	Tools        string `json:"tools" example:"file_read,file_write"`
}

type UpdateAgentRequest struct {
	Name         *string `json:"name,omitempty" example:"助手"`
	Description  *string `json:"description,omitempty" example:"通用助手"`
	Type         *string `json:"type,omitempty" example:"main"`
	IsSubAgent   *bool   `json:"is_sub_agent,omitempty" example:"false"`
	ProviderID   *string `json:"provider_id,omitempty" example:"provider-001"`
	Model        *string `json:"model,omitempty" example:"gpt-4"`
	Instructions *string `json:"instructions,omitempty" example:"你是一个智能助手"`
	Tools        *string `json:"tools,omitempty" example:"file_read,file_write"`
}

// Create godoc
// @Summary 创建 Agent
// @Description 创建一个新的 Agent
// @Tags Agent
// @Accept json
// @Produce json
// @Param agent body CreateAgentRequest true "Agent 信息"
// @Success 201 {object} models.Response{data=models.Agent}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/agents [post]
func (c *AgentController) Create(ctx *web.Context) {
	var req CreateAgentRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

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
	if err := c.agentService.Create(agent); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(agent))
}

// GetByID godoc
// @Summary 获取 Agent 详情
// @Description 根据 ID 获取 Agent 详情
// @Tags Agent
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} models.Response{data=models.Agent}
// @Failure 404 {object} models.Response
// @Router /api/agents/{id} [get]
func (c *AgentController) GetByID(ctx *web.Context) {
	id := ctx.Param("id")
	agent, err := c.agentService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Agent not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(agent))
}

// GetAll godoc
// @Summary 获取所有 Agent
// @Description 获取所有 Agent 列表
// @Tags Agent
// @Produce json
// @Success 200 {object} models.Response{data=[]models.Agent}
// @Failure 500 {object} models.Response
// @Router /api/agents [get]
func (c *AgentController) GetAll(ctx *web.Context) {
	agents, err := c.agentService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(agents))
}

// GetByType godoc
// @Summary 按类型获取 Agent
// @Description 根据类型获取 Agent 列表
// @Tags Agent
// @Produce json
// @Param type query string false "Agent 类型 (main/sub/custom)"
// @Success 200 {object} models.Response{data=[]models.Agent}
// @Failure 500 {object} models.Response
// @Router /api/agents/type [get]
func (c *AgentController) GetByType(ctx *web.Context) {
	agentType := ctx.Query("type")
	if agentType == "" {
		c.GetAll(ctx)
		return
	}

	agents, err := c.agentService.GetByType(models.AgentType(agentType))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(agents))
}

// Update godoc
// @Summary 更新 Agent
// @Description 更新 Agent 信息
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param agent body UpdateAgentRequest true "更新内容"
// @Success 200 {object} models.Response{data=models.Agent}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/agents/{id} [put]
func (c *AgentController) Update(ctx *web.Context) {
	id := ctx.Param("id")
	var req UpdateAgentRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.IsSubAgent != nil {
		updates["is_sub_agent"] = *req.IsSubAgent
	}
	if req.ProviderID != nil {
		updates["provider_id"] = *req.ProviderID
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.Instructions != nil {
		updates["instructions"] = *req.Instructions
	}
	if req.Tools != nil {
		updates["tools"] = *req.Tools
	}

	agent, err := c.agentService.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(agent))
}

// Delete godoc
// @Summary 删除 Agent
// @Description 删除指定 Agent
// @Tags Agent
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/agents/{id} [delete]
func (c *AgentController) Delete(ctx *web.Context) {
	id := ctx.Param("id")
	if err := c.agentService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Agent deleted successfully"}))
}

// AddTool godoc
// @Summary 为 Agent 添加工具
// @Description 为指定 Agent 添加工具
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param tool body map[string]string true "工具 ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/agents/{id}/tools [post]
func (c *AgentController) AddTool(ctx *web.Context) {
	agentID := ctx.Param("id")
	var req struct {
		ToolID string `json:"tool_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	updates := map[string]interface{}{
		"tools": req.ToolID,
	}

	_, err := c.agentService.Update(agentID, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Tool added to agent successfully"}))
}

// RemoveTool godoc
// @Summary 移除 Agent 工具
// @Description 从 Agent 中移除指定工具
// @Tags Agent
// @Produce json
// @Param id path string true "Agent ID"
// @Param tool_name path string true "工具名称"
// @Success 200 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/agents/{id}/tools/{tool_name} [delete]
func (c *AgentController) RemoveTool(ctx *web.Context) {
	agentID := ctx.Param("id")
	toolName := ctx.Param("tool_name")

	agent, err := c.agentService.GetByID(agentID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Agent not found"))
		return
	}

	tools := strings.Split(agent.Tools, ",")
	for i, tool := range tools {
		if strings.TrimSpace(tool) == toolName {
			tools = append(tools[:i], tools[i+1:]...)
			break
		}
	}
	agent.Tools = strings.Join(tools, ",")

	updates := map[string]interface{}{
		"tools": agent.Tools,
	}
	_, err = c.agentService.Update(agentID, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Tool removed from agent successfully"}))
}
