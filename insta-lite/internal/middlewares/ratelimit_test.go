package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/huguescodeur/insta-lite/internal/middlewares"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	middlewares.InitRateLimiter(client)
	return mr, client
}

func makeRequest(handler http.Handler, ip string) int {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = ip + ":1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr.Code
}

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	newTestRedis(t)
	mw := middlewares.ConfigurableRateLimiter("test_allow", 3, time.Minute)
	handler := mw(okHandler)

	for i := 1; i <= 3; i++ {
		if code := makeRequest(handler, "10.0.0.1"); code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, code)
		}
	}
}

func TestRateLimiter_BlocksWhenOverLimit(t *testing.T) {
	newTestRedis(t)
	mw := middlewares.ConfigurableRateLimiter("test_block", 3, time.Minute)
	handler := mw(okHandler)

	for i := 0; i < 3; i++ {
		makeRequest(handler, "10.0.0.2")
	}

	if code := makeRequest(handler, "10.0.0.2"); code != http.StatusTooManyRequests {
		t.Errorf("expected 429 after exceeding limit, got %d", code)
	}
}

func TestRateLimiter_DifferentIPs_IndependentCounters(t *testing.T) {
	newTestRedis(t)
	mw := middlewares.ConfigurableRateLimiter("test_ips", 2, time.Minute)
	handler := mw(okHandler)

	// IP A exhausts its limit
	makeRequest(handler, "10.0.0.3")
	makeRequest(handler, "10.0.0.3")
	if code := makeRequest(handler, "10.0.0.3"); code != http.StatusTooManyRequests {
		t.Errorf("IP A: expected 429, got %d", code)
	}

	// IP B still allowed
	if code := makeRequest(handler, "10.0.0.4"); code != http.StatusOK {
		t.Errorf("IP B: expected 200 (independent counter), got %d", code)
	}
}

func TestRateLimiter_TTL_SetOnceNotReset(t *testing.T) {
	mr, _ := newTestRedis(t)
	mw := middlewares.ConfigurableRateLimiter("test_ttl", 2, time.Minute)
	handler := mw(okHandler)
	ip := "10.0.0.5"

	// First request sets TTL
	makeRequest(handler, ip)
	ttl1 := mr.TTL("ratelimit:test_ttl:" + ip)

	// Advance miniredis clock by 10s
	mr.FastForward(10 * time.Second)

	// More requests — should NOT reset TTL
	makeRequest(handler, ip)
	makeRequest(handler, ip)
	ttl2 := mr.TTL("ratelimit:test_ttl:" + ip)

	if ttl2 >= ttl1 {
		t.Errorf("TTL should be decreasing (set once), but ttl1=%v ttl2=%v", ttl1, ttl2)
	}
}

func TestRateLimiter_WindowExpires_AllowsAgain(t *testing.T) {
	mr, _ := newTestRedis(t)
	mw := middlewares.ConfigurableRateLimiter("test_expire", 2, 30*time.Second)
	handler := mw(okHandler)
	ip := "10.0.0.6"

	// Exhaust the limit
	makeRequest(handler, ip)
	makeRequest(handler, ip)
	if code := makeRequest(handler, ip); code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after limit, got %d", code)
	}

	// Advance past the window
	mr.FastForward(31 * time.Second)

	// Should be allowed again
	if code := makeRequest(handler, ip); code != http.StatusOK {
		t.Errorf("expected 200 after window reset, got %d", code)
	}
}

func TestRateLimiter_FailOpen_WhenRedisDown(t *testing.T) {
	mr, _ := newTestRedis(t)
	mw := middlewares.ConfigurableRateLimiter("test_failopen", 2, time.Minute)
	handler := mw(okHandler)

	mr.Close() // simulate Redis down

	// should pass through (fail-open), not return 500
	if code := makeRequest(handler, "10.0.0.7"); code != http.StatusOK {
		t.Errorf("expected fail-open 200 when Redis is down, got %d", code)
	}
}
