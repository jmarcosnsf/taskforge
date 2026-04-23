package middleware

import (
	"net/http"
	"taskforge/internal/api"
	"time"

	"github.com/redis/go-redis/v9"
)

func RateLimiter(redisClient *redis.Client, maxRequests int64, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			key := "rate:" + ip

			count, err := redisClient.Incr(r.Context(), key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				redisClient.Expire(r.Context(), key, window)
			}

			if count > maxRequests {
				api.SendJSON(w, api.Response{Error: "too many requests"}, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}