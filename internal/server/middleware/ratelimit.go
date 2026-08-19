package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu        sync.Mutex
	entries   map[string]*rateLimitEntry
	max       int
	window    time.Duration
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

// extractIP extracts the client IP from RemoteAddr, handling both direct connections
// and proxied requests via X-Forwarded-For and X-Real-IP headers.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (may contain multiple IPs: client, proxy1, proxy2)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		ip := strings.TrimSpace(ips[0])
		if parsedIP, _, err := net.SplitHostPort(ip); err == nil {
			return parsedIP
		}
		return ip
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractIP(r)
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
