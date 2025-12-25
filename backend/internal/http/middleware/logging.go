package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/solomonczyk/izborator/internal/logger"
)

// RequestLogger logs structured HTTP access events with request id and client IP.
func RequestLogger(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			log.Info("HTTP request", map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      ww.Status(),
				"duration_ms": time.Since(start).Milliseconds(),
				"request_id":  middleware.GetReqID(r.Context()),
				"host":        r.Host,
				"remote_ip":   realIP(r),
				"user_agent":  r.UserAgent(),
				"referer":     r.Referer(),
			})
		})
	}
}

// realIP returns best-effort client IP using headers or RemoteAddr.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	if rip := strings.TrimSpace(r.Header.Get("X-Real-IP")); rip != "" {
		return rip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	return r.RemoteAddr
}
