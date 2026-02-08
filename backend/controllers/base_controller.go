package controllers

import (
	"iano_chat/models"
	"iano_chat/utils"
	"net/http"
)

type BaseController struct{}

func (bc *BaseController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.JSONResponse(w, http.StatusOK, models.Success(map[string]string{
		"status": "ok",
	}))
}
