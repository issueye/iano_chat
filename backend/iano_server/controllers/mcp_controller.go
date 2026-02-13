package controllers

import (
	"context"
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
)

type MCPController struct {
	mcpService *services.MCPService
}

func NewMCPController(mcpService *services.MCPService) *MCPController {
	return &MCPController{
		mcpService: mcpService,
	}
}

type CreateMCPServerRequest struct {
	Name      string `json:"name" example:"filesystem"`                                                    // 服务器名称
	Desc      string `json:"desc" example:"本地文件系统访问"`                                                      // 描述
	Transport string `json:"transport" example:"stdio"`                                                    // 传输类型: stdio, sse, http
	Command   string `json:"command" example:"npx"`                                                        // 命令 (用于 stdio)
	Args      string `json:"args" example:"[\"-y\",\"@modelcontextprotocol/server-filesystem\",\"/tmp\"]"` // 命令参数 (JSON 数组)
	Env       string `json:"env" example:"{}"`                                                             // 环境变量 (JSON 对象)
	URL       string `json:"url" example:"http://localhost:8080/sse"`                                      // URL (用于 sse/http)
	Enabled   *bool  `json:"enabled,omitempty" example:"true"`                                             // 是否启用
	Version   string `json:"version,omitempty" example:"1.0.0"`                                            // 版本
	Author    string `json:"author,omitempty" example:"system"`                                            // 作者
	Icon      string `json:"icon,omitempty" example:"folder"`                                              // 图标
}

type UpdateMCPServerRequest struct {
	Name      *string `json:"name,omitempty" example:"filesystem"`
	Desc      *string `json:"desc,omitempty" example:"本地文件系统访问"`
	Transport *string `json:"transport,omitempty" example:"stdio"`
	Command   *string `json:"command,omitempty" example:"npx"`
	Args      *string `json:"args,omitempty" example:"[\"-y\",\"@modelcontextprotocol/server-filesystem\"]"`
	Env       *string `json:"env,omitempty" example:"{}"`
	URL       *string `json:"url,omitempty" example:"http://localhost:8080/sse"`
	Enabled   *bool   `json:"enabled,omitempty" example:"true"`
	Version   *string `json:"version,omitempty" example:"1.0.0"`
	Author    *string `json:"author,omitempty" example:"system"`
	Icon      *string `json:"icon,omitempty" example:"folder"`
}

type ConnectMCPServerRequest struct {
	ServerID string `json:"server_id" example:"server-uuid"` // 服务器 ID
}

type CallMCToolRequest struct {
	ServerID  string                 `json:"server_id" example:"server-uuid"` // 服务器 ID
	ToolName  string                 `json:"tool_name" example:"read_file"`   // 工具名称
	Arguments map[string]interface{} `json:"arguments" example:"{}"`          // 工具参数
}

// CreateServer godoc
// @Summary 创建 MCP 服务器
// @Description 创建一个新的 MCP 服务器配置
// @Tags MCP
// @Accept json
// @Produce json
// @Param server body CreateMCPServerRequest true "MCP 服务器信息"
// @Success 201 {object} models.Response{data=models.MCPServer}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/mcp/servers [post]
func (c *MCPController) CreateServer(ctx *web.Context) {
	var req CreateMCPServerRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	server := &models.MCPServer{
		Name:      req.Name,
		Desc:      req.Desc,
		Transport: models.MCPTransportType(req.Transport),
		Command:   req.Command,
		Args:      req.Args,
		Env:       req.Env,
		URL:       req.URL,
		Version:   req.Version,
		Author:    req.Author,
		Icon:      req.Icon,
		Enabled:   true,
	}
	if req.Enabled != nil {
		server.Enabled = *req.Enabled
	}

	server.NewID()

	if err := c.mcpService.ServerService.Create(server); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(server))
}

// GetServerByID godoc
// @Summary 获取 MCP 服务器详情
// @Description 根据 ID 获取 MCP 服务器详情
// @Tags MCP
// @Produce json
// @Param id path string true "服务器 ID"
// @Success 200 {object} models.Response{data=models.MCPServer}
// @Failure 404 {object} models.Response
// @Router /api/mcp/servers/{id} [get]
func (c *MCPController) GetServerByID(ctx *web.Context) {
	id := ctx.Param("id")
	server, err := c.mcpService.ServerService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Server not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(server))
}

// GetAllServers godoc
// @Summary 获取所有 MCP 服务器
// @Description 获取所有 MCP 服务器列表
// @Tags MCP
// @Produce json
// @Success 200 {object} models.Response{data=[]models.MCPServer}
// @Router /api/mcp/servers [get]
func (c *MCPController) GetAllServers(ctx *web.Context) {
	servers, err := c.mcpService.ServerService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(servers))
}

// UpdateServer godoc
// @Summary 更新 MCP 服务器
// @Description 更新 MCP 服务器配置
// @Tags MCP
// @Accept json
// @Produce json
// @Param id path string true "服务器 ID"
// @Param server body UpdateMCPServerRequest true "MCP 服务器信息"
// @Success 200 {object} models.Response{data=models.MCPServer}
// @Failure 404 {object} models.Response
// @Router /api/mcp/servers/{id} [put]
func (c *MCPController) UpdateServer(ctx *web.Context) {
	id := ctx.Param("id")
	var req UpdateMCPServerRequest
	if err := ctx.BindAndValidate(&req); err != nil {
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
	if req.Transport != nil {
		updates["transport"] = *req.Transport
	}
	if req.Command != nil {
		updates["command"] = *req.Command
	}
	if req.Args != nil {
		updates["args"] = *req.Args
	}
	if req.Env != nil {
		updates["env"] = *req.Env
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Version != nil {
		updates["version"] = *req.Version
	}
	if req.Author != nil {
		updates["author"] = *req.Author
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}

	server, err := c.mcpService.ServerService.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(server))
}

// DeleteServer godoc
// @Summary 删除 MCP 服务器
// @Description 删除 MCP 服务器配置
// @Tags MCP
// @Produce json
// @Param id path string true "服务器 ID"
// @Success 200 {object} models.Response
// @Failure 404 {object} models.Response
// @Router /api/mcp/servers/{id} [delete]
func (c *MCPController) DeleteServer(ctx *web.Context) {
	id := ctx.Param("id")

	if err := c.mcpService.DisconnectServer(id); err != nil {
	}

	if err := c.mcpService.ServerService.Delete(id); err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Server not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(nil))
}

// ConnectServer godoc
// @Summary 连接 MCP 服务器
// @Description 连接到 MCP 服务器
// @Tags MCP
// @Accept json
// @Produce json
// @Param request body ConnectMCPServerRequest true "连接信息"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Router /api/mcp/servers/connect [post]
func (c *MCPController) ConnectServer(ctx *web.Context) {
	var req ConnectMCPServerRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	if err := c.mcpService.ConnectServer(context.Background(), req.ServerID); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(nil))
}

// DisconnectServer godoc
// @Summary 断开 MCP 服务器
// @Description 断开 MCP 服务器连接
// @Tags MCP
// @Produce json
// @Param id path string true "服务器 ID"
// @Success 200 {object} models.Response
// @Router /api/mcp/servers/{id}/disconnect [post]
func (c *MCPController) DisconnectServer(ctx *web.Context) {
	id := ctx.Param("id")
	if err := c.mcpService.DisconnectServer(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(nil))
}

// GetServerTools godoc
// @Summary 获取 MCP 服务器工具列表
// @Description 获取 MCP 服务器提供的工具列表
// @Tags MCP
// @Produce json
// @Param id path string true "服务器 ID"
// @Success 200 {object} models.Response{data=[]models.MCPServerTool}
// @Router /api/mcp/servers/{id}/tools [get]
func (c *MCPController) GetServerTools(ctx *web.Context) {
	id := ctx.Param("id")
	tools, err := c.mcpService.ServerToolService.GetByServerID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(tools))
}

// CallTool godoc
// @Summary 调用 MCP 工具
// @Description 调用 MCP 服务器上的工具
// @Tags MCP
// @Accept json
// @Produce json
// @Param request body CallMCToolRequest true "工具调用信息"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Router /api/mcp/tools/call [post]
func (c *MCPController) CallTool(ctx *web.Context) {
	var req CallMCToolRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	result, err := c.mcpService.CallTool(context.Background(), req.ServerID, req.ToolName, req.Arguments)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(result))
}

// ListTools godoc
// @Summary 列出 MCP 工具
// @Description 列出 MCP 服务器上的工具
// @Tags MCP
// @Produce json
// @Param id path string true "服务器 ID"
// @Success 200 {object} models.Response
// @Router /api/mcp/servers/{id}/list-tools [get]
func (c *MCPController) ListTools(ctx *web.Context) {
	id := ctx.Param("id")
	result, err := c.mcpService.ListTools(context.Background(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(result))
}

// Ping godoc
// @Summary Ping MCP 服务器
// @Description 检查 MCP 服务器连接状态
// @Tags MCP
// @Produce json
// @Param id path string true "服务器 ID"
// @Success 200 {object} models.Response
// @Router /api/mcp/servers/{id}/ping [get]
func (c *MCPController) Ping(ctx *web.Context) {
	id := ctx.Param("id")
	if err := c.mcpService.Ping(context.Background(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(nil))
}
