package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/oms-saas/oms-saas-go/internal/response"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	times := rl.requests[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= rl.limit {
		rl.requests[key] = filtered
		return false
	}
	rl.requests[key] = append(filtered, now)
	return true
}

func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if !rl.allow(key) {
				response.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, try again later")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
