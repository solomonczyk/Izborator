package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appErrors "github.com/solomonczyk/izborator/internal/errors"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/scrapingstats"
)

// StatsHandler обработчик для статистики парсинга
type StatsHandler struct {
	*BaseHandler
	service *scrapingstats.Service
}

// NewStatsHandler создаёт новый обработчик статистики
func NewStatsHandler(service *scrapingstats.Service, log *logger.Logger, translator *i18n.Translator) *StatsHandler {
	return &StatsHandler{
		BaseHandler: NewBaseHandler(log, translator),
		service:     service,
	}
}

// GetOverallStats получает общую статистику парсинга
// GET /api/v1/stats/overall?days=7
func (h *StatsHandler) GetOverallStats(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7 // По умолчанию 7 дней
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil {
			appErr := appErrors.NewValidationError("Invalid days parameter", err)
			h.RespondAppError(w, r, appErr)
			return
		}
		if d < 1 || d > 365 {
			appErr := appErrors.NewValidationError("Days must be between 1 and 365", nil)
			h.RespondAppError(w, r, appErr)
			return
		}
		days = d
	}

	stats, err := h.service.GetOverallStats(days)
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to get overall stats", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	h.RespondJSON(w, http.StatusOK, stats)
}

// GetShopStats получает статистику по конкретному магазину
// GET /api/v1/stats/shops/{shop_id}?days=7
func (h *StatsHandler) GetShopStats(w http.ResponseWriter, r *http.Request) {
	shopID := chi.URLParam(r, "shop_id")
	if shopID == "" {
		appErr := appErrors.NewBadRequest("Shop ID is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7 // По умолчанию 7 дней
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil {
			appErr := appErrors.NewValidationError("Invalid days parameter", err)
			h.RespondAppError(w, r, appErr)
			return
		}
		if d < 1 || d > 365 {
			appErr := appErrors.NewValidationError("Days must be between 1 and 365", nil)
			h.RespondAppError(w, r, appErr)
			return
		}
		days = d
	}

	stats, err := h.service.GetShopStats(shopID, days)
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to get shop stats", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	h.RespondJSON(w, http.StatusOK, stats)
}

// GetRecentStats получает статистику за последние N дней
// GET /api/v1/stats/recent?days=7
func (h *StatsHandler) GetRecentStats(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 7 // По умолчанию 7 дней
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil {
			appErr := appErrors.NewValidationError("Invalid days parameter", err)
			h.RespondAppError(w, r, appErr)
			return
		}
		if d < 1 || d > 365 {
			appErr := appErrors.NewValidationError("Days must be between 1 and 365", nil)
			h.RespondAppError(w, r, appErr)
			return
		}
		days = d
	}

	stats, err := h.service.GetRecentStats(days)
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to get recent stats", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	h.RespondJSON(w, http.StatusOK, stats)
}
