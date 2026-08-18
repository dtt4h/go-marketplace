package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type rateLimitEntry struct {
	count    int
	windowStart time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	entries    map[string]*rateLimitEntry
	max        int
	window     time.Duration
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		max:     max,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]
	if !exists || now.Sub(entry.windowStart) >= rl.window {
		rl.entries[key] = &rateLimitEntry{count: 1, windowStart: now}
		return true
	}

	entry.count++
	return entry.count <= rl.max
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			if now.Sub(entry.windowStart) >= rl.window {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if !rl.Allow(key) {
				w.Header().Set("Retry-After", "60")
				httputil.ErrorWithDetails(w, http.StatusTooManyRequests, "RATE_LIMITED",
					"too many requests, please try again later", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
