package routes

import (
	"iano_chat/controllers"
	"iano_chat/middleware"
	"net/http"
	"strings"
)

func SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	baseController := controllers.BaseController{}

	mux.HandleFunc("/health", baseController.HealthCheck)

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
		case http.MethodPost:
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if !strings.HasPrefix(path, "/api/users/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	})

	handler := middleware.CORS(mux)
	handler = middleware.Logger(handler)

	return handler
}
