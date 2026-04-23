package handler

import (
	"net/http"
	"taskforge/db/sqlc"
	"taskforge/internal/api"
	"time"

	customMiddleware "taskforge/internal/middleware"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	repo        *sqlc.Queries
	jwtSecret   string
	redisClient *redis.Client
}

func NewHandler(repository *sqlc.Queries, jwtSecret string, redisClient *redis.Client) http.Handler {
	h := &Handler{repo: repository, jwtSecret: jwtSecret, redisClient: redisClient}

	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(customMiddleware.RateLimiter(redisClient, 100, time.Minute))

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

		r.Post("/{id}/tasks", h.CreateTask)
		r.Get("/{id}/tasks", h.ListTasks)
		r.Get("/{id}/tasks/{taskID}", h.GetTaskByID)
		r.Patch("/{id}/tasks/{taskID}/status", h.UpdateTaskStatus)
		r.Patch("/{id}/tasks/{taskID}/assign", h.AssignTask)
		r.Delete("/{id}/tasks/{taskID}", h.RemoveTask)
	})
	
	return r
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	api.SendJSON(w, api.Response{Data: "ok"}, http.StatusOK)
}
