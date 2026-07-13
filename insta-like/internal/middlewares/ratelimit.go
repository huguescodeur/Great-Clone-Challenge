package middlewares

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu       sync.Mutex
	visitors = map[string]*visitor{}
)

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	v, ok := visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rate.Every(time.Minute), 10)}
		visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func RateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := strings.Cut(r.RemoteAddr, ":")
		if !getVisitor(ip).Allow() {
			http.Error(w, "Trop de requêtes, réessayez dans une minute", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
