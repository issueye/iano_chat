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
	userController := controllers.NewUserController()

	mux.HandleFunc("/health", baseController.HealthCheck)

	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userController.GetUsers(w, r)
		case http.MethodPost:
			userController.CreateUser(w, r)
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

		switch r.Method {
		case http.MethodGet:
			userController.GetUser(w, r)
		case http.MethodPut:
			userController.UpdateUser(w, r)
		case http.MethodDelete:
			userController.DeleteUser(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	handler := middleware.CORS(mux)
	handler = middleware.Logger(handler)

	return handler
}
