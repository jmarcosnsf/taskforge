package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"taskforge/db/sqlc"
	"taskforge/internal/api"

	"golang.org/x/crypto/bcrypt"
)

type CreateUserRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to decode body", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	newUser, err := h.repo.CreateUser(r.Context(), sqlc.CreateUserParams{
		Name: body.Name,
		Email: body.Email,
		Password: string(hash),
	})
	
	if err != nil {
		slog.Error("user already exists", "error", err.Error())
		api.SendJSON(w, api.Response{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	
	newUser.Password = ""
	api.SendJSON(w, api.Response{Data: newUser}, http.StatusCreated)
}
