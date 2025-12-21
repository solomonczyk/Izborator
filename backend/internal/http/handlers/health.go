package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthHandler обработчик для health check
type HealthHandler struct{}

// NewHealthHandler создаёт новый обработчик health check
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check обрабатывает health check
// GET /api/health
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}

