package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// contextKey для хранения trace ID в контексте
type contextKey string

const traceIDKey contextKey = "trace_id"

// TraceID middleware для добавления уникального ID каждому request'у
func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверить есть ли trace ID в заголовке
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Добавить trace ID в заголовок ответа
		w.Header().Set("X-Trace-ID", traceID)

		// Добавить trace ID в контекст
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTraceID извлекает trace ID из контекста
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}
