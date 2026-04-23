package handler

import (
	"context"
	"net/http"
	"taskforge/internal/api"
	"taskforge/internal/middleware"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func parseUUIDParam(w http.ResponseWriter,r *http.Request) (uuid.UUID, bool) {
	ID := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(ID)
	if err != nil{
		api.SendJSON(w, api.Response{Error:"invalid UUID"}, http.StatusBadRequest)
		return uuid.Nil, false
	}

	return parsed, true

}

func getUserIDFromContext(w http.ResponseWriter, ctx context.Context) (uuid.UUID, bool) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		api.SendJSON(w, api.Response{Error: "unauthorized"}, http.StatusUnauthorized)
		return uuid.Nil, false
	}

	return userID, true
}

func (h *Handler) isTeamMember(ctx context.Context, teamID pgtype.UUID, userID uuid.UUID) (bool, error) {
    members, err := h.repo.GetTeamMembers(ctx, teamID)
    if err != nil {
        return false, err
    }
    currentUser := pgtype.UUID{Bytes: userID, Valid: true}
    for _, m := range members {
        if m.ID == currentUser {
            return true, nil
        }
    }
    return false, nil
}