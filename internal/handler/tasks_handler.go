package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"taskforge/db/sqlc"
	"taskforge/internal/api"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var body CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to decode body", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	if body.Title == "" {
		api.SendJSON(w, api.Response{Error: "title is required"}, http.StatusBadRequest)
		return
	}

	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	isMember, err := h.isTeamMember(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true}, userID)
	if err != nil {
		slog.Error("failed to check membership", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}
	if !isMember {
		api.SendJSON(w, api.Response{Error: "forbidden"}, http.StatusForbidden)
		return
	}

	task, err := h.repo.CreateTask(r.Context(), sqlc.CreateTaskParams{
		Title:       body.Title,
		Description: pgtype.Text{String: body.Description, Valid: true},
		TeamID:      pgtype.UUID{Bytes: teamID, Valid: true},
	})
	if err != nil {
		slog.Error("failed to create task", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	h.redisClient.Del(r.Context(), "tasks:"+teamID.String())

	api.SendJSON(w, api.Response{Data: task}, http.StatusCreated)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	isMember, err := h.isTeamMember(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true}, userID)
	if err != nil {
		slog.Error("failed to check membership", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if !isMember {
		api.SendJSON(w, api.Response{Error: "forbidden"}, http.StatusForbidden)
		return
	}

	cacheKey := "tasks:" + teamID.String()
	cached, err := h.redisClient.Get(r.Context(), cacheKey).Result()
	if err == nil {
		var tasks []sqlc.Task
		json.Unmarshal([]byte(cached), &tasks)
		api.SendJSON(w, api.Response{Data: tasks}, http.StatusOK)
		return
	}

	tasks, err := h.repo.ListTasks(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true})
	if err != nil {
		slog.Error("failed to get tasks list", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	tasksJSON, _ := json.Marshal(tasks)
	h.redisClient.Set(r.Context(), cacheKey, tasksJSON, 5*time.Minute)

	api.SendJSON(w, api.Response{Data: tasks}, http.StatusOK)
}

func (h *Handler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	parsed, err := uuid.Parse(taskID)
	if err != nil {
		api.SendJSON(w, api.Response{Error: "invalid task id"}, http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	isMember, err := h.isTeamMember(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true}, userID)
	if err != nil {
		slog.Error("failed to check membership", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if !isMember {
		api.SendJSON(w, api.Response{Error: "forbidden"}, http.StatusForbidden)
		return
	}

	task, err := h.repo.GetTaskByID(r.Context(), pgtype.UUID{Bytes: parsed, Valid: true})
	if err != nil {
		slog.Error("failed to get task by id", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	api.SendJSON(w, api.Response{Data: task}, http.StatusOK)
}

type UpdateTaskStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	parsed, err := uuid.Parse(taskID)
	if err != nil {
		api.SendJSON(w, api.Response{Error: "invalid task id"}, http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	isMember, err := h.isTeamMember(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true}, userID)
	if err != nil {
		slog.Error("failed to check membership", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if !isMember {
		api.SendJSON(w, api.Response{Error: "forbidden"}, http.StatusForbidden)
		return
	}

	var body UpdateTaskStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to decode body", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid body request"}, http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"done":        true,
	}

	if !validStatuses[body.Status] {
		api.SendJSON(w, api.Response{Error: "invalid status"}, http.StatusBadRequest)
		return
	}

	err = h.repo.UpdateTaskStatus(r.Context(), sqlc.UpdateTaskStatusParams{
		Status: body.Status,
		ID:     pgtype.UUID{Bytes: parsed, Valid: true},
	})
	if err != nil {
		slog.Error("failed to update task", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	h.redisClient.Del(r.Context(), "tasks:"+teamID.String())
	api.SendJSON(w, api.Response{Data: "Task successfully updated"}, http.StatusOK)
}

type AssignTaskRequest struct {
	Email string `json:"email"`
}

func (h *Handler) AssignTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	parsed, err := uuid.Parse(taskID)
	if err != nil {
		api.SendJSON(w, api.Response{Error: "invalid task id"}, http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	isMember, err := h.isTeamMember(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true}, userID)
	if err != nil {
		slog.Error("failed to check membership", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if !isMember {
		api.SendJSON(w, api.Response{Error: "forbidden"}, http.StatusForbidden)
		return
	}

	var body AssignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to decode body", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid body request"}, http.StatusBadRequest)
		return
	}

	targetUser, err := h.repo.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.SendJSON(w, api.Response{Error: "user not found"}, http.StatusNotFound)
			return
		}
		slog.Error("failed to get user ID", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	isMember, err = h.isTeamMember(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true}, targetUser.ID.Bytes)
	if err != nil {
		slog.Error("failed to check membership", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if !isMember {
		api.SendJSON(w, api.Response{Error: "target user is not a member of this team"}, http.StatusBadRequest)
		return
	}

	err = h.repo.AssignTask(r.Context(), sqlc.AssignTaskParams{
		UserID: targetUser.ID,
		ID:     pgtype.UUID{Bytes: parsed, Valid: true},
	})
	if err != nil {
		slog.Error("failed to assign task", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	h.redisClient.Del(r.Context(), "tasks:"+teamID.String())
	api.SendJSON(w, api.Response{Data: "task assigned successfully"}, http.StatusOK)
}

func (h *Handler) RemoveTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	parsed, err := uuid.Parse(taskID)
	if err != nil {
		api.SendJSON(w, api.Response{Error: "invalid task id"}, http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	isMember, err := h.isTeamMember(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true}, userID)
	if err != nil {
		slog.Error("failed to check membership", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if !isMember {
		api.SendJSON(w, api.Response{Error: "forbidden"}, http.StatusForbidden)
		return
	}

	if err := h.repo.RemoveTask(r.Context(), pgtype.UUID{Bytes: parsed, Valid: true}); err != nil {
		slog.Error("failed to remove task", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	h.redisClient.Del(r.Context(), "tasks:"+teamID.String())
	api.SendJSON(w, api.Response{Data: "task successfully removed"}, http.StatusOK)
}
