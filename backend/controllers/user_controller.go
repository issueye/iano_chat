package controllers

import (
	"encoding/json"
	"iano_chat/models"
	"iano_chat/utils"
	"net/http"
	"strconv"
	"strings"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) GetUsers(w http.ResponseWriter, r *http.Request) {
	users := []models.UserResponse{
		{
			ID:       1,
			Username: "user1",
			Email:    "user1@example.com",
		},
		{
			ID:       2,
			Username: "user2",
			Email:    "user2@example.com",
		},
	}
	utils.JSONResponse(w, http.StatusOK, models.Success(users))
}

func (uc *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, models.Error(400, "invalid user id"))
		return
	}

	user := models.UserResponse{
		ID:       id,
		Username: "user" + idStr,
		Email:    "user" + idStr + "@example.com",
	}
	utils.JSONResponse(w, http.StatusOK, models.Success(user))
}

func (uc *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, models.Error(400, "invalid request body"))
		return
	}

	if user.Username == "" || user.Email == "" {
		utils.JSONResponse(w, http.StatusBadRequest, models.Error(400, "username and email are required"))
		return
	}

	user.ID = 1
	utils.JSONResponse(w, http.StatusCreated, models.Success(user.ToResponse()))
}

func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, models.Error(400, "invalid user id"))
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, models.Error(400, "invalid request body"))
		return
	}

	user.ID = id
	utils.JSONResponse(w, http.StatusOK, models.Success(user.ToResponse()))
}

func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, models.Error(400, "invalid user id"))
		return
	}

	utils.JSONResponse(w, http.StatusOK, models.Success(nil))
}
