package middleware

import (
	"net/http"
	"strings"

	"github.com/dtt4h/go-marketplace/internal/config"
)

func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	allowedOrigins := map[string]bool{}
	if cfg.DevMode {
		allowedOrigins["http://localhost:5173"] = true
		allowedOrigins["http://localhost:3000"] = true
	}
	for _, o := range strings.Split(cfg.Server.CORSOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowedOrigins[o] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
