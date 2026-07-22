package middlewares

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRateLimiter(client *redis.Client) {
	rdb = client
}

func ConfigurableRateLimiter(ruleName string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, found := strings.Cut(r.RemoteAddr, ":")
			if !found {
				ip = r.RemoteAddr
			}
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ips := strings.Split(xff, ",")
				ip = strings.TrimSpace(ips[0])
			}

			key := "ratelimit:" + ruleName + ":" + ip

			count, err := rdb.Incr(r.Context(), key).Result()
			if err != nil {

				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {

				rdb.Expire(context.Background(), key, window)
			}

			if int(count) > limit {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "Trop de requêtes. Veuillez ralentir."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
