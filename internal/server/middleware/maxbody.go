package middleware

import (
	"net/http"

	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

// MaxBodySize returns a middleware that limits the request body size.
// maxBytes is the maximum allowed size in bytes (e.g. 10 << 20 for 10MB).
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				httputil.Error(w, http.StatusRequestEntityTooLarge,
					"PAYLOAD_TOO_LARGE", "request body too large")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
