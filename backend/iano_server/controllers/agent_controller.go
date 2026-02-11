package controllers

import (
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
)

type AgentController struct {
	agentService        *services.AgentService
	agentManagerService *services.AgentManagerService
}

func NewAgentController(agentService *services.AgentService, agentManagerService *services.AgentManagerService) *AgentController {
	return &AgentController{
		agentService:        agentService,
		agentManagerService: agentManagerService,
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

	agent, err := c.agentManagerService.CreateAgent(ctx.Request.Context(), &services.CreateAgentRequest{
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		IsSubAgent:   req.IsSubAgent,
		ProviderID:   req.ProviderID,
		Model:        req.Model,
		Instructions: req.Instructions,
		Tools:        req.Tools,
	})
	if err != nil {
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

	if err := c.agentManagerService.ReloadAgent(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusOK, models.Success(map[string]interface{}{
			"agent":   agent,
			"warning": "Agent updated but failed to reload instance: " + err.Error(),
		}))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(agent))
}

func (c *AgentController) Delete(ctx *web.Context) {
	id := ctx.Param("id")
	if err := c.agentManagerService.DeleteAgent(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Agent deleted successfully"}))
}

func (c *AgentController) GetInstanceInfo(ctx *web.Context) {
	id := ctx.Param("id")
	info, err := c.agentManagerService.GetAgentInfo(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(info))
}

func (c *AgentController) ListInstances(ctx *web.Context) {
	instances := c.agentManagerService.ListAgentInstances()
	ctx.JSON(http.StatusOK, models.Success(instances))
}

func (c *AgentController) Reload(ctx *web.Context) {
	id := ctx.Param("id")
	if err := c.agentManagerService.ReloadAgent(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Agent reloaded successfully"}))
}

func (c *AgentController) GetStats(ctx *web.Context) {
	stats := c.agentManagerService.GetManagerStats()
	ctx.JSON(http.StatusOK, models.Success(stats))
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

	if err := c.agentManagerService.AddToolToAgent(ctx.Request.Context(), agentID, req.ToolID); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Tool added to agent successfully"}))
}

func (c *AgentController) RemoveTool(ctx *web.Context) {
	agentID := ctx.Param("id")
	toolName := ctx.Param("tool_name")

	if err := c.agentManagerService.RemoveToolFromAgent(agentID, toolName); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Tool removed from agent successfully"}))
}
