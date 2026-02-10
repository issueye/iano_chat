package controllers

import (
	"iano_server/models"
	"iano_server/pkg/web"
	"iano_server/services"
	"net/http"
)

type ProviderController struct {
	providerService *services.ProviderService
}

func NewProviderController(providerService *services.ProviderService) *ProviderController {
	return &ProviderController{providerService: providerService}
}

type CreateProviderRequest struct {
	Name        string  `json:"name" binding:"required"`
	BaseUrl     string  `json:"base_url" binding:"required"`
	ApiKey      string  `json:"api_key" binding:"required"`
	Model       string  `json:"model" binding:"required"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

type UpdateProviderRequest struct {
	Name        *string  `json:"name,omitempty"`
	BaseUrl     *string  `json:"base_url,omitempty"`
	ApiKey      *string  `json:"api_key,omitempty"`
	Model       *string  `json:"model,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
}

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
	}

	if err := c.providerService.Create(provider); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusCreated, models.Success(provider))
}

func (c *ProviderController) GetByID(ctx *web.Context) {
	id := ctx.Param("id")

	provider, err := c.providerService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, models.Fail("Provider not found"))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(provider))
}

func (c *ProviderController) GetAll(ctx *web.Context) {
	providers, err := c.providerService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(providers))
}

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
		updates["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		updates["max_tokens"] = *req.MaxTokens
	}

	provider, err := c.providerService.Update(id, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, models.Success(provider))
}

func (c *ProviderController) Delete(ctx *web.Context) {
	id := ctx.Param("id")

	if err := c.providerService.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Fail(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, models.Success(map[string]string{"message": "Provider deleted successfully"}))
}
