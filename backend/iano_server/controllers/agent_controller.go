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
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	IsSubAgent   bool   `json:"is_sub_agent"`
	ProviderID   string `json:"provider_id"`
	Model        string `json:"model"`
	Instructions string `json:"instructions"`
	Tools        string `json:"tools"`
}

type UpdateAgentRequest struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	Type         *string `json:"type,omitempty"`
	IsSubAgent   *bool   `json:"is_sub_agent,omitempty"`
	ProviderID   *string `json:"provider_id,omitempty"`
	Model        *string `json:"model,omitempty"`
	Instructions *string `json:"instructions,omitempty"`
	Tools        *string `json:"tools,omitempty"`
}

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

func (c *AgentController) GetByID(ctx *web.Context) {
	id := ctx.Param("id")
	agent, err := c.agentService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Agent not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(agent))
}

func (c *AgentController) GetAll(ctx *web.Context) {
	agents, err := c.agentService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(agents))
}

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

func (c *AgentController) Delete(ctx *web.Context) {
	id := ctx.Param("id")
	if err := c.agentService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Agent deleted successfully"}))
}

func (c *AgentController) AddTool(ctx *web.Context) {
	agentID := ctx.Param("id")
	var req struct {
		ToolID string `json:"tool_id"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	// map
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

func (c *AgentController) RemoveTool(ctx *web.Context) {
	agentID := ctx.Param("id")
	toolName := ctx.Param("tool_name")

	// 获取当前 Agent 配置
	agent, err := c.agentService.GetByID(agentID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Agent not found"))
		return
	}

	// 从 Tools 字符串中移除指定工具
	tools := strings.Split(agent.Tools, ",")
	for i, tool := range tools {
		if strings.TrimSpace(tool) == toolName {
			tools = append(tools[:i], tools[i+1:]...)
			break
		}
	}
	agent.Tools = strings.Join(tools, ",")

	// 更新 Agent 配置
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
