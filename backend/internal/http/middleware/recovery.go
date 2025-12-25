package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/solomonczyk/izborator/internal/logger"
)

// Recovery catches panics, logs them, and returns 500 instead of crashing.
func Recovery(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", map[string]interface{}{
						"error":      rec,
						"stacktrace": string(debug.Stack()),
						"request_id": middleware.GetReqID(r.Context()),
						"method":     r.Method,
						"path":       r.URL.Path,
					})

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
