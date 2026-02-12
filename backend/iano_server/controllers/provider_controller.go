package controllers

import (
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
	"net/http"
)

type ProviderController struct {
	providerService *services.ProviderService
}

func NewProviderController(providerService *services.ProviderService) *ProviderController {
	return &ProviderController{providerService: providerService}
}

type CreateProviderRequest struct {
	Name        string  `json:"name" binding:"required" example:"OpenAI"`
	BaseUrl     string  `json:"base_url" binding:"required" example:"https://api.openai.com/v1"`
	ApiKey      string  `json:"api_key" binding:"required" example:"sk-xxx"`
	Model       string  `json:"model" binding:"required" example:"gpt-4"`
	Temperature float32 `json:"temperature" example:"0.7"`
	MaxTokens   int     `json:"max_tokens" example:"2000"`
	IsDefault   bool    `json:"is_default" example:"true"`
}

type UpdateProviderRequest struct {
	Name        *string  `json:"name,omitempty" example:"OpenAI"`
	BaseUrl     *string  `json:"base_url,omitempty" example:"https://api.openai.com/v1"`
	ApiKey      *string  `json:"api_key,omitempty" example:"sk-xxx"`
	Model       *string  `json:"model,omitempty" example:"gpt-4"`
	Temperature *float32 `json:"temperature,omitempty" example:"0.7"`
	MaxTokens   *int     `json:"max_tokens,omitempty" example:"2000"`
	IsDefault   *bool    `json:"is_default,omitempty" example:"true"`
}

// Create godoc
// @Summary 创建 Provider
// @Description 创建一个新的模型提供商配置
// @Tags Provider
// @Accept json
// @Produce json
// @Param provider body CreateProviderRequest true "Provider 信息"
// @Success 201 {object} models.Response{data=models.Provider}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/providers [post]
func (c *ProviderController) Create(ctx *web.Context) {
	var req CreateProviderRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	provider := &models.Provider{
		Name:        req.Name,
		BaseUrl:     req.BaseUrl,
		ApiKey:      req.ApiKey,
		Model:       req.Model,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		IsDefault:   req.IsDefault,
	}

	provider.NewID()

	if err := c.providerService.Create(provider); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(provider))
}

// GetByID godoc
// @Summary 获取 Provider 详情
// @Description 根据 ID 获取 Provider 详情
// @Tags Provider
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} models.Response{data=models.Provider}
// @Failure 404 {object} models.Response
// @Router /api/providers/{id} [get]
func (c *ProviderController) GetByID(ctx *web.Context) {
	id := ctx.Param("id")

	provider, err := c.providerService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Provider not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(provider))
}

// GetAll godoc
// @Summary 获取所有 Provider
// @Description 获取所有 Provider 列表
// @Tags Provider
// @Produce json
// @Success 200 {object} models.Response{data=[]models.Provider}
// @Failure 500 {object} models.Response
// @Router /api/providers [get]
func (c *ProviderController) GetAll(ctx *web.Context) {
	providers, err := c.providerService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(providers))
}

// GetDefault godoc
// @Summary 获取默认 Provider
// @Description 获取默认的 Provider 配置
// @Tags Provider
// @Produce json
// @Success 200 {object} models.Response{data=models.Provider}
// @Failure 404 {object} models.Response
// @Router /api/providers/default [get]
func (c *ProviderController) GetDefault(ctx *web.Context) {
	provider, err := c.providerService.GetDefault()
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Default provider not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(provider))
}

// Update godoc
// @Summary 更新 Provider
// @Description 更新 Provider 配置
// @Tags Provider
// @Accept json
// @Produce json
// @Param id path string true "Provider ID"
// @Param provider body UpdateProviderRequest true "更新内容"
// @Success 200 {object} models.Response{data=models.Provider}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/providers/{id} [put]
func (c *ProviderController) Update(ctx *web.Context) {
	id := ctx.Param("id")

	var req UpdateProviderRequest
	if err := ctx.Bind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Fail(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.BaseUrl != nil {
		updates["base_url"] = *req.BaseUrl
	}
	if req.ApiKey != nil {
		updates["api_key"] = *req.ApiKey
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.Temperature != nil {
		updates["temperature"] = models.NullFloat32{
			Float32: *req.Temperature,
			Valid:   *req.Temperature != 0,
		}
	}
	if req.MaxTokens != nil {
		updates["max_tokens"] = models.NullInt{
			Int:   *req.MaxTokens,
			Valid: *req.MaxTokens != 0,
		}
	}
	if req.IsDefault != nil {
		updates["is_default"] = *req.IsDefault
	}

	provider, err := c.providerService.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(provider))
}

// Delete godoc
// @Summary 删除 Provider
// @Description 删除指定 Provider
// @Tags Provider
// @Produce json
// @Param id path string true "Provider ID"
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/providers/{id} [delete]
func (c *ProviderController) Delete(ctx *web.Context) {
	id := ctx.Param("id")

	if err := c.providerService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Provider deleted successfully"}))
}
