package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/solomonczyk/izborator/internal/logger"
)

// RequestLogger middleware для логирования HTTP запросов
func RequestLogger(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Обёртка для ResponseWriter чтобы получить статус код
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Выполняем следующий handler
			next.ServeHTTP(ww, r)

			// Логируем запрос
			log.Info("HTTP request", map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     ww.Status(),
				"duration":   time.Since(start).Milliseconds(),
				"remote_ip":  r.RemoteAddr,
				"user_agent": r.UserAgent(),
			})
		})
	}
}
