package middleware

import (
	"context"
	"net/http"
	"strings"
	"taskforge/internal/api"
	"taskforge/internal/auth"

	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				api.SendJSON(w, api.Response{Error: "missing authorization header"}, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(header, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				api.SendJSON(w, api.Response{Error: "invalid authorization header"}, http.StatusUnauthorized)
				return
			}

			userID, err := auth.ValidateToken(parts[1], jwtSecret)
			if err != nil {
				api.SendJSON(w, api.Response{Error: "invalid or expired token"}, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}