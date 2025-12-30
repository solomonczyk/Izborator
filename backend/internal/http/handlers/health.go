package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/solomonczyk/izborator/internal/http/response"
	"github.com/solomonczyk/izborator/internal/logger"
)

// HealthHandler обработчик для health check
type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
	log   *logger.Logger
}

// NewHealthHandler создаёт новый обработчик health check
func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client, log *logger.Logger) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
		log:   log,
	}
}

// Check обрабатывает health check (основной endpoint)
// GET /api/health
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	}

	if err := response.WriteSuccess(w, status); err != nil {
		h.log.Error("Failed to write health check response", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// Alive обрабатывает liveness probe (быстрая проверка что сервис живой)
// GET /api/health/live
func (h *HealthHandler) Alive(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now().Unix(),
	}

	if err := response.WriteSuccess(w, status); err != nil {
		h.log.Error("Failed to write liveness response", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// Ready обрабатывает readiness probe (проверка готовности к принятию трафика)
// GET /api/health/ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checks := map[string]bool{
		"database": h.checkDatabase(ctx),
		"redis":    h.checkRedis(ctx),
	}

	allReady := true
	for _, ready := range checks {
		if !ready {
			allReady = false
			break
		}
	}

	status := http.StatusOK
	if !allReady {
		status = http.StatusServiceUnavailable
	}

	result := map[string]interface{}{
		"ready":     allReady,
		"checks":    checks,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(result)
}

// Full обрабатывает полную проверку здоровья системы
// GET /api/health/full
func (h *HealthHandler) Full(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	defer func() {
		h.log.Info("Full health check completed", map[string]interface{}{
			"duration_ms": time.Since(start).Milliseconds(),
		})
	}()

	dbCheck := h.checkDatabaseFull(ctx)
	redisCheck := h.checkRedisFull(ctx)

	allHealthy := dbCheck["healthy"].(bool) && redisCheck["healthy"].(bool)
	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	result := map[string]interface{}{
		"status":    "ok",
		"healthy":   allHealthy,
		"timestamp": time.Now().Unix(),
		"components": map[string]interface{}{
			"database": dbCheck,
			"redis":    redisCheck,
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(result)
}

// checkDatabase проверяет доступность БД
func (h *HealthHandler) checkDatabase(ctx context.Context) bool {
	if h.db == nil {
		return false
	}

	return h.db.Ping(ctx) == nil
}

// checkDatabaseFull проверяет БД с деталями
func (h *HealthHandler) checkDatabaseFull(ctx context.Context) map[string]interface{} {
	start := time.Now()
	result := map[string]interface{}{
		"healthy":    false,
		"latency_ms": int64(0),
		"error":      "unknown",
	}

	if h.db == nil {
		result["error"] = "database client not initialized"
		return result
	}

	err := h.db.Ping(ctx)
	latency := time.Since(start).Milliseconds()
	result["latency_ms"] = latency

	if err != nil {
		result["error"] = err.Error()
		h.log.Error("Database health check failed", map[string]interface{}{
			"error": err.Error(),
		})
		return result
	}

	// Попробовать простой запрос
	testCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var testVal int
	queryErr := h.db.QueryRow(testCtx, "SELECT 1").Scan(&testVal)
	if queryErr != nil {
		result["error"] = queryErr.Error()
		h.log.Error("Database query test failed", map[string]interface{}{
			"error": queryErr.Error(),
		})
		return result
	}

	result["healthy"] = true
	result["error"] = nil
	return result
}

// checkRedis проверяет доступность Redis
func (h *HealthHandler) checkRedis(ctx context.Context) bool {
	if h.redis == nil {
		return false
	}

	err := h.redis.Ping(ctx).Err()
	return err == nil
}

// checkRedisFull проверяет Redis с деталями
func (h *HealthHandler) checkRedisFull(ctx context.Context) map[string]interface{} {
	start := time.Now()
	result := map[string]interface{}{
		"healthy":    false,
		"latency_ms": int64(0),
		"error":      "unknown",
	}

	if h.redis == nil {
		result["error"] = "redis client not initialized"
		return result
	}

	err := h.redis.Ping(ctx).Err()
	latency := time.Since(start).Milliseconds()
	result["latency_ms"] = latency

	if err != nil {
		result["error"] = err.Error()
		h.log.Error("Redis health check failed", map[string]interface{}{
			"error": err.Error(),
		})
		return result
	}

	result["healthy"] = true
	result["error"] = nil
	return result
}
