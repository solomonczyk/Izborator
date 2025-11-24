package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/solomonczyk/izborator/internal/logger"
)

// Recovery middleware для обработки паник
func Recovery(log *logger.Logger) func(next http.Handler) http.Handler {
	return middleware.Recoverer
}

