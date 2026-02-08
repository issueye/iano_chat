package middleware

import (
	"iano_chat/models"
	"iano_chat/utils"
	"net/http"
	"strings"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.JSONResponse(w, http.StatusUnauthorized, models.Error(401, "authorization header required"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.JSONResponse(w, http.StatusUnauthorized, models.Error(401, "invalid authorization header format"))
			return
		}

		token := parts[1]
		if token == "" {
			utils.JSONResponse(w, http.StatusUnauthorized, models.Error(401, "token required"))
			return
		}

		next(w, r)
	}
}
