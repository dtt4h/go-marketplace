package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

// RateLimiter provides IP-based rate limiting.
// Uses Redis when available (atomic INCR + EXPIRE), falls back to in-memory.
type RateLimiter struct {
	rdb    *redis.Client
	log    *slog.Logger
	mu     sync.Mutex
	mem    map[string]*rateLimitEntry
	max    int
	window time.Duration
}

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

func NewRateLimiter(rdb *redis.Client, max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rdb:    rdb,
		log:    slog.Default(),
		mem:    make(map[string]*rateLimitEntry),
		max:    max,
		window: window,
	}
	if rdb == nil {
		go rl.cleanup()
	}
	return rl
}

// NewRateLimiterWithLog creates a RateLimiter with a custom logger.
func NewRateLimiterWithLog(rdb *redis.Client, log *slog.Logger, max int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rdb:    rdb,
		log:    log,
		mem:    make(map[string]*rateLimitEntry),
		max:    max,
		window: window,
	}
	if rdb == nil {
		go rl.cleanup()
	}
	return rl
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
	if rl.rdb != nil {
		return rl.allowRedis(ctx, key)
	}
	return rl.allowMemory(key)
}

func (rl *RateLimiter) allowRedis(ctx context.Context, key string) bool {
	redisKey := fmt.Sprintf("rl:%s", key)

	count, err := rl.rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		// Redis error — fail open (allow request) but log it
		rl.log.Warn("redis rate limiter error, falling back to in-memory",
			slog.String("key", key), slog.String("error", err.Error()))
		return rl.allowMemory(key)
	}

	if count == 1 {
		rl.rdb.Expire(ctx, redisKey, rl.window)
	}

	return count <= int64(rl.max)
}

func (rl *RateLimiter) allowMemory(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.mem[key]
	if !exists || now.Sub(entry.windowStart) >= rl.window {
		rl.mem[key] = &rateLimitEntry{count: 1, windowStart: now}
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
		for key, entry := range rl.mem {
			if now.Sub(entry.windowStart) >= rl.window {
				delete(rl.mem, key)
			}
		}
		rl.mu.Unlock()
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		ip := strings.TrimSpace(ips[0])
		if parsedIP, _, err := net.SplitHostPort(ip); err == nil {
			return parsedIP
		}
		return ip
	}

	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}

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
			if !rl.Allow(r.Context(), key) {
				w.Header().Set("Retry-After", "60")
				httputil.ErrorWithDetails(w, http.StatusTooManyRequests, "RATE_LIMITED",
					"too many requests, please try again later", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
