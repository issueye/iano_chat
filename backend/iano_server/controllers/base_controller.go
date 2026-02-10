package controllers

import (
	"iano_server/models"
	web "iano_web"
)

type BaseController struct{}

func (bc *BaseController) HealthCheck(ctx *web.Context) {
	ctx.JSON(200, models.Success(map[string]string{
		"status": "ok",
	}))
}
