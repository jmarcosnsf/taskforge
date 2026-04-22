package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"taskforge/db/sqlc"
	"taskforge/internal/api"
	"taskforge/internal/auth"

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
	if body.Name == "" || body.Email == "" || body.Password == "" {
    api.SendJSON(w, api.Response{Error: "invalid request body (name,email and password required)"}, http.StatusBadRequest)
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

type LoginRequest struct{
	Email string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) LoginUser (w http.ResponseWriter, r *http.Request) {
	var body LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err !=nil {
		slog.Error("failed to decode body", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		slog.Error("invalid email or password", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid email or password"}, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil{
		slog.Error("invalid email or password", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid email or password"}, http.StatusUnauthorized)
		return
	}
 
	token, err := auth.GenerateToken(user.ID, h.jwtSecret)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		api.SendJSON(w,api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	api.SendJSON(w, api.Response{Data: token}, http.StatusOK)
}

