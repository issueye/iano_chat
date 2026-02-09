package controllers

import (
	"iano_chat/models"
	"iano_chat/pkg/web"
	"iano_chat/services"
	"net/http"
)

type ToolController struct {
	toolService *services.ToolService
}

func NewToolController(toolService *services.ToolService) *ToolController {
	return &ToolController{toolService: toolService}
}

type CreateToolRequest struct {
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Returns    string `json:"returns"`
	Example    string `json:"example,omitempty"`
	Type       string `json:"type"`
	Config     string `json:"config,omitempty"`
	Parameters string `json:"parameters,omitempty"`
	Version    string `json:"version,omitempty"`
	Author     string `json:"author,omitempty"`
}

type UpdateToolRequest struct {
	Name       *string `json:"name,omitempty"`
	Desc       *string `json:"desc,omitempty"`
	Returns    *string `json:"returns,omitempty"`
	Example    *string `json:"example,omitempty"`
	Type       *string `json:"type,omitempty"`
	Config     *string `json:"config,omitempty"`
	Parameters *string `json:"parameters,omitempty"`
	Version    *string `json:"version,omitempty"`
	Author     *string `json:"author,omitempty"`
}

func (c *ToolController) Create(ctx *web.Context) {
	var req CreateToolRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	tool := &models.Tool{
		Name:       req.Name,
		Desc:       req.Desc,
		Returns:    req.Returns,
		Example:    req.Example,
		Type:       models.ToolType(req.Type),
		Config:     req.Config,
		Parameters: req.Parameters,
		Version:    req.Version,
		Author:     req.Author,
	}

	if err := c.toolService.Create(tool); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(tool))
}

func (c *ToolController) GetByID(ctx *web.Context) {
	id := ctx.Param("id")
	tool, err := c.toolService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Tool not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(tool))
}

func (c *ToolController) GetAll(ctx *web.Context) {
	tools, err := c.toolService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(tools))
}

func (c *ToolController) GetByType(ctx *web.Context) {
	toolType := ctx.Query("type")
	if toolType == "" {
		c.GetAll(ctx)
		return
	}

	tools, err := c.toolService.GetByType(models.ToolType(toolType))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(tools))
}

func (c *ToolController) GetByStatus(ctx *web.Context) {
	status := ctx.Query("status")
	if status == "" {
		c.GetAll(ctx)
		return
	}

	tools, err := c.toolService.GetByStatus(models.ToolStatus(status))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(tools))
}

func (c *ToolController) Update(ctx *web.Context) {
	id := ctx.Param("id")
	var req UpdateToolRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Desc != nil {
		updates["desc"] = *req.Desc
	}
	if req.Returns != nil {
		updates["returns"] = *req.Returns
	}
	if req.Example != nil {
		updates["example"] = *req.Example
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Config != nil {
		updates["config"] = *req.Config
	}
	if req.Parameters != nil {
		updates["parameters"] = *req.Parameters
	}
	if req.Version != nil {
		updates["version"] = *req.Version
	}
	if req.Author != nil {
		updates["author"] = *req.Author
	}

	tool, err := c.toolService.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(tool))
}

func (c *ToolController) UpdateConfig(ctx *web.Context) {
	id := ctx.Param("id")
	var req struct {
		Config string `json:"config"`
	}
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	tool, err := c.toolService.UpdateConfig(id, req.Config)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(tool))
}

func (c *ToolController) Delete(ctx *web.Context) {
	id := ctx.Param("id")
	if err := c.toolService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Tool deleted successfully"}))
}
