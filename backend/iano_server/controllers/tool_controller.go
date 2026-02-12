package controllers

import (
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
)

type ToolController struct {
	toolService *services.ToolService
}

func NewToolController(toolService *services.ToolService) *ToolController {
	return &ToolController{
		toolService: toolService,
	}
}

type CreateToolRequest struct {
	Name       string `json:"name" example:"file_read"`
	Desc       string `json:"desc" example:"读取文件内容"`
	Returns    string `json:"returns" example:"文件内容字符串"`
	Example    string `json:"example,omitempty" example:"读取 /path/to/file.txt"`
	Type       string `json:"type" example:"builtin"`
	Config     string `json:"config,omitempty" example:"{}"`
	Parameters string `json:"parameters,omitempty" example:"{\"path\": \"string\"}"`
	Version    string `json:"version,omitempty" example:"1.0.0"`
	Author     string `json:"author,omitempty" example:"system"`
}

type UpdateToolRequest struct {
	Name       *string `json:"name,omitempty" example:"file_read"`
	Desc       *string `json:"desc,omitempty" example:"读取文件内容"`
	Returns    *string `json:"returns,omitempty" example:"文件内容字符串"`
	Example    *string `json:"example,omitempty" example:"读取 /path/to/file.txt"`
	Type       *string `json:"type,omitempty" example:"builtin"`
	Config     *string `json:"config,omitempty" example:"{}"`
	Parameters *string `json:"parameters,omitempty" example:"{\"path\": \"string\"}"`
	Version    *string `json:"version,omitempty" example:"1.0.0"`
	Author     *string `json:"author,omitempty" example:"system"`
}

// Create godoc
// @Summary 创建工具
// @Description 创建一个新的工具定义
// @Tags Tool
// @Accept json
// @Produce json
// @Param tool body CreateToolRequest true "工具信息"
// @Success 201 {object} models.Response{data=models.Tool}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/tools [post]
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

	tool.NewID()

	if err := c.toolService.Create(tool); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(tool))
}

// GetByID godoc
// @Summary 获取工具详情
// @Description 根据 ID 获取工具详情
// @Tags Tool
// @Produce json
// @Param id path string true "工具 ID"
// @Success 200 {object} models.Response{data=models.Tool}
// @Failure 404 {object} models.Response
// @Router /api/tools/{id} [get]
func (c *ToolController) GetByID(ctx *web.Context) {
	id := ctx.Param("id")
	tool, err := c.toolService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Tool not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(tool))
}

// GetAll godoc
// @Summary 获取所有工具
// @Description 获取所有工具列表
// @Tags Tool
// @Produce json
// @Success 200 {object} models.Response{data=[]models.Tool}
// @Failure 500 {object} models.Response
// @Router /api/tools [get]
func (c *ToolController) GetAll(ctx *web.Context) {
	tools, err := c.toolService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(tools))
}

// GetByType godoc
// @Summary 按类型获取工具
// @Description 根据类型获取工具列表
// @Tags Tool
// @Produce json
// @Param type query string false "工具类型 (builtin/custom/mcp)"
// @Success 200 {object} models.Response{data=[]models.Tool}
// @Failure 500 {object} models.Response
// @Router /api/tools/type [get]
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

// GetByStatus godoc
// @Summary 按状态获取工具
// @Description 根据状态获取工具列表
// @Tags Tool
// @Produce json
// @Param status query string false "工具状态 (active/inactive/deprecated)"
// @Success 200 {object} models.Response{data=[]models.Tool}
// @Failure 500 {object} models.Response
// @Router /api/tools/status [get]
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

// Update godoc
// @Summary 更新工具
// @Description 更新工具信息
// @Tags Tool
// @Accept json
// @Produce json
// @Param id path string true "工具 ID"
// @Param tool body UpdateToolRequest true "更新内容"
// @Success 200 {object} models.Response{data=models.Tool}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/tools/{id} [put]
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

// UpdateConfig godoc
// @Summary 更新工具配置
// @Description 更新指定工具的配置
// @Tags Tool
// @Accept json
// @Produce json
// @Param id path string true "工具 ID"
// @Param config body map[string]string true "配置信息"
// @Success 200 {object} models.Response{data=models.Tool}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/tools/{id}/config [put]
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

// Delete godoc
// @Summary 删除工具
// @Description 删除指定工具
// @Tags Tool
// @Produce json
// @Param id path string true "工具 ID"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/tools/{id} [delete]
func (c *ToolController) Delete(ctx *web.Context) {
	id := ctx.Param("id")
	if err := c.toolService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Tool deleted successfully"}))
}

// Test godoc
// @Summary 测试工具
// @Description 测试工具定义是否正确加载
// @Tags Tool
// @Produce json
// @Param id path string true "工具 ID"
// @Success 200 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/tools/{id}/test [get]
func (c *ToolController) Test(ctx *web.Context) {
	id := ctx.Param("id")
	tool, err := c.toolService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Tool not found"))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(map[string]interface{}{
		"tool":    tool,
		"message": "Tool definition loaded successfully",
	}))
}
