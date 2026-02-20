package controllers

import (
	"iano_server/models"
	"iano_server/services"
	web "iano_web"
)

type BaseController struct {
	providerService *services.ProviderService
	sessionService  *services.SessionService
	toolService     *services.ToolService
	agentService    *services.AgentService
}

func NewBaseController(
	providerService *services.ProviderService,
	sessionService *services.SessionService,
	toolService *services.ToolService,
	agentService *services.AgentService,
) *BaseController {
	return &BaseController{
		providerService: providerService,
		sessionService:  sessionService,
		toolService:     toolService,
		agentService:    agentService,
	}
}

func (bc *BaseController) HealthCheck(ctx *web.Context) {
	status := "ok"
	dbStatus := "ok"
	providerStatus := "no providers"

	// 检查数据库
	if err := bc.sessionService.HealthCheck(); err != nil {
		dbStatus = "error"
		status = "degraded"
	}

	// 检查 Provider
	providers, err := bc.providerService.GetAll()
	if err != nil || len(providers) == 0 {
		providerStatus = "no providers"
		status = "degraded"
	} else {
		// 检查是否有默认 Provider
		defaultProvider, err := bc.providerService.GetDefault()
		if err != nil {
			providerStatus = "no default provider"
			status = "degraded"
		} else {
			providerStatus = "ready (" + defaultProvider.Name + ")"
		}
	}

	ctx.JSON(200, models.Success(map[string]interface{}{
		"status":         status,
		"database":       dbStatus,
		"provider":       providerStatus,
		"provider_count": len(providers),
	}))
}
