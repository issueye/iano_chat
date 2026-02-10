package utils

import (
	"encoding/json"
	"iano_server/models"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, statusCode int, data models.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
