package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"taskforge/db/sqlc"
	"taskforge/internal/api"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var body CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to decode body", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	if body.Name == "" {
		api.SendJSON(w, api.Response{Error: "name is required"}, http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	newTeam, err := h.repo.CreateTeam(r.Context(), sqlc.CreateTeamParams{
		Name:        body.Name,
		Description: pgtype.Text{String: body.Description, Valid: true},
		OwnerID:     pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		slog.Error("failed to create team", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	err = h.repo.AddTeamMember(r.Context(), sqlc.AddTeamMemberParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		TeamID: newTeam.ID,
	})
	if err != nil {
		slog.Error("failed to add owner on created team", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	api.SendJSON(w, api.Response{Data: newTeam}, http.StatusCreated)
}

func (h *Handler) GetTeamsByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	teams, err := h.repo.GetTeamsByUser(r.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		slog.Error("failed to get teams", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	api.SendJSON(w, api.Response{Data: teams}, http.StatusOK)
}

func (h *Handler) GetTeamByID(w http.ResponseWriter, r *http.Request) {
	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	team, err := h.repo.GetTeamByID(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.SendJSON(w, api.Response{Error: "team not found"}, http.StatusNotFound)
			return
		}
		slog.Error("failed to get team", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	api.SendJSON(w, api.Response{Data: team}, http.StatusOK)
}

type AddTeamMemberRequest struct {
	Email string `json:"email"`
}

func (h *Handler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	team, err := h.repo.GetTeamByID(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.SendJSON(w, api.Response{Error: "team not found"}, http.StatusNotFound)
			return
		}
		slog.Error("failed to get team", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if team.OwnerID.Bytes != uuid.UUID(userID) {
		api.SendJSON(w, api.Response{Error: "forbidden"}, http.StatusForbidden)
		return
	}

	var body AddTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to decode body", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.SendJSON(w, api.Response{Error: "user not found"}, http.StatusNotFound)
			return
		}
		slog.Error("failed to get user", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	err = h.repo.AddTeamMember(r.Context(), sqlc.AddTeamMemberParams{
		UserID: user.ID,
		TeamID: pgtype.UUID{Bytes: teamID, Valid: true},
	})
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			api.SendJSON(w, api.Response{Error: "user is already a member of this team"}, http.StatusConflict)
			return
		}
		slog.Error("failed to add member on team", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	h.redisClient.Del(r.Context(), "members:"+teamID.String())
	
	api.SendJSON(w, api.Response{Data: "member successfully added to team"}, http.StatusOK)
}

type RemoveTeamMemberRequest struct {
	Email string `json:"email"`
}

func (h *Handler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(w, r.Context())
	if !ok {
		return
	}

	teamID, ok := parseUUIDParam(w, r)
	if !ok {
		return
	}

	team, err := h.repo.GetTeamByID(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.SendJSON(w, api.Response{Error: "team not found"}, http.StatusNotFound)
			return
		}
		slog.Error("failed to get team", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if team.OwnerID.Bytes != uuid.UUID(userID) {
		api.SendJSON(w, api.Response{Error: "forbidden"}, http.StatusForbidden)
		return
	}

	var body RemoveTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("failed to decode body", "error", err)
		api.SendJSON(w, api.Response{Error: "invalid request body"}, http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.SendJSON(w, api.Response{Error: "user not found"}, http.StatusNotFound)
			return
		}
		slog.Error("failed to get user", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if user.ID == team.OwnerID {
		api.SendJSON(w, api.Response{Error: "owner cannot be removed from team"}, http.StatusBadRequest)
		return
	}

	err = h.repo.RemoveTeamMember(r.Context(), sqlc.RemoveTeamMemberParams{
		UserID: user.ID,
		TeamID: team.ID,
	})
	if err != nil {
		slog.Error("failed to delete team member", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	h.redisClient.Del(r.Context(), "members:"+teamID.String())
	api.SendJSON(w, api.Response{Data: "member successfully removed"}, http.StatusOK)
}

func (h *Handler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
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

	cacheKey := "members:" + teamID.String()
	cached, err := h.redisClient.Get(r.Context(), cacheKey).Result()
	if err == nil {
		var members []sqlc.User
		json.Unmarshal([]byte(cached), &members)
		api.SendJSON(w, api.Response{Data: members}, http.StatusOK)
		return
	}

	members, err := h.repo.GetTeamMembers(r.Context(), pgtype.UUID{Bytes: teamID, Valid: true})
	if err != nil {
		slog.Error("failed to get team members", "error", err)
		api.SendJSON(w, api.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	for i := range members {
		members[i].Password = ""
	}

	membersJSON, _ := json.Marshal(members)
	h.redisClient.Set(r.Context(), cacheKey, membersJSON, 5*time.Minute)

	api.SendJSON(w, api.Response{Data: members}, http.StatusOK)
}
