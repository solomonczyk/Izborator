package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	httpMiddleware "github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/scrapingstats"
)

// StatsHandler обработчик для статистики парсинга
type StatsHandler struct {
	service    *scrapingstats.Service
	logger     *logger.Logger
	translator *i18n.Translator
}

// NewStatsHandler создаёт новый обработчик статистики
func NewStatsHandler(service *scrapingstats.Service, log *logger.Logger, translator *i18n.Translator) *StatsHandler {
	return &StatsHandler{
		service:    service,
		logger:     log,
		translator: translator,
	}
}

// GetOverallStats получает общую статистику парсинга
// GET /api/v1/stats/overall?days=7
func (h *StatsHandler) GetOverallStats(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7 // По умолчанию 7 дней
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	stats, err := h.service.GetOverallStats(days)
	if err != nil {
		h.logger.Error("GetOverallStats failed", map[string]interface{}{
			"error": err.Error(),
			"days":  days,
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.overall_stats_failed")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// GetShopStats получает статистику по конкретному магазину
// GET /api/v1/stats/shops/{shop_id}?days=7
func (h *StatsHandler) GetShopStats(w http.ResponseWriter, r *http.Request) {
	shopID := chi.URLParam(r, "shop_id")
	if shopID == "" {
		h.respondError(w, r, http.StatusBadRequest, "api.errors.shop_id_required")
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7 // По умолчанию 7 дней
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	stats, err := h.service.GetShopStats(shopID, days)
	if err != nil {
		h.logger.Error("GetShopStats failed", map[string]interface{}{
			"error":  err.Error(),
			"shop_id": shopID,
			"days":    days,
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.shop_stats_failed")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// GetRecentStats получает последние записи статистики
// GET /api/v1/stats/recent?limit=20
func (h *StatsHandler) GetRecentStats(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20 // По умолчанию 20 записей
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	stats, err := h.service.GetRecentStats(limit)
	if err != nil {
		h.logger.Error("GetRecentStats failed", map[string]interface{}{
			"error": err.Error(),
			"limit": limit,
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.recent_stats_failed")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats,
		"count": len(stats),
	})
}

// respondJSON отправляет JSON ответ
func (h *StatsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error": err,
		})
	}
}

// respondError отправляет JSON ошибку
func (h *StatsHandler) respondError(w http.ResponseWriter, r *http.Request, status int, key string) {
	lang := httpMiddleware.GetLangFromContext(r.Context())
	message := h.translator.T(lang, key)
	if message == key || message == "" {
		message = h.translator.T("en", key)
		if message == "" {
			message = key
		}
	}
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

