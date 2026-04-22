package handler

import (
	"net/http"
	"taskforge/db/sqlc"
	"taskforge/internal/api"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	customMiddleware "taskforge/internal/middleware"
)

type Handler struct {
	repo *sqlc.Queries
	jwtSecret string
}

func NewHandler(repository *sqlc.Queries, jwtSecret string) http.Handler {
	h := &Handler{repo: repository, jwtSecret: jwtSecret}

	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Get("/status", h.GetStatus)
	r.Post("/register", h.CreateUser)
	r.Post("/login", h.LoginUser)

	r.Route("/teams", func(r chi.Router){
		r.Use(customMiddleware.AuthMiddleware(jwtSecret))
		r.Post("/", h.CreateTeam)
		r.Get("/" , h.GetTeamsByUser)
		r.Get("/{id}", h.GetTeamByID)
		r.Post("/{id}/members", h.AddTeamMember)
		r.Delete("/{id}/members", h.RemoveTeamMember)
		r.Get("/{id}/members", h.GetTeamMembers)
	})
	
	return r
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	api.SendJSON(w, api.Response{Data: "ok"}, http.StatusOK)
}
