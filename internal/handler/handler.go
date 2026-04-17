package handler

import (
	"net/http"
	"taskforge/db/sqlc"
	"taskforge/internal/api"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Handler struct {
	repo *sqlc.Queries
}

func NewHandler(repository *sqlc.Queries) http.Handler {
	h := &Handler{repo: repository}

	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Get("/status", h.GetStatus)
	r.Post("/register", h.CreateUser)
	
	return r
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	api.SendJSON(w, api.Response{Data: "ok"}, http.StatusOK)
}
