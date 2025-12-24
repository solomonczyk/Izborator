package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appErrors "github.com/solomonczyk/izborator/internal/errors"
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
		d, err := strconv.Atoi(daysStr)
		if err != nil {
			appErr := appErrors.NewValidationError("Invalid days parameter", err)
			h.respondAppError(w, r, appErr)
			return
		}
		if d < 1 || d > 365 {
			appErr := appErrors.NewValidationError("Days must be between 1 and 365", nil)
			h.respondAppError(w, r, appErr)
			return
		}
		days = d
	}

	stats, err := h.service.GetOverallStats(days)
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to get overall stats", err)
		h.respondAppError(w, r, appErr)
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// GetShopStats получает статистику по конкретному магазину
// GET /api/v1/stats/shops/{shop_id}?days=7
func (h *StatsHandler) GetShopStats(w http.ResponseWriter, r *http.Request) {
	shopID := chi.URLParam(r, "shop_id")
	if shopID == "" {
		appErr := appErrors.NewBadRequest("Shop ID is required", nil)
		h.respondAppError(w, r, appErr)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7 // По умолчанию 7 дней
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err != nil {
			appErr := appErrors.NewValidationError("Invalid days parameter", err)
			h.respondAppError(w, r, appErr)
			return
		}
		if d < 1 || d > 365 {
			appErr := appErrors.NewValidationError("Days must be between 1 and 365", nil)
			h.respondAppError(w, r, appErr)
			return
		}
		days = d
	}

	stats, err := h.service.GetShopStats(shopID, days)
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to get shop stats", err)
		h.respondAppError(w, r, appErr)
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
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
			h.respondAppError(w, r, appErr)
			return
		}
		if d < 1 || d > 365 {
			appErr := appErrors.NewValidationError("Days must be between 1 and 365", nil)
			h.respondAppError(w, r, appErr)
			return
		}
		days = d
	}

	stats, err := h.service.GetRecentStats(days)
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to get recent stats", err)
		h.respondAppError(w, r, appErr)
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// respondJSON отправляет JSON ответ
func (h *StatsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error": err,
		})
	}
}

// respondAppError отправляет JSON ошибку из AppError
func (h *StatsHandler) respondAppError(w http.ResponseWriter, r *http.Request, err *appErrors.AppError) {
	lang := httpMiddleware.GetLangFromContext(r.Context())

	// Пытаемся получить локализованное сообщение
	messageKey := "api.errors." + err.Code
	message := h.translator.T(lang, messageKey)
	if message == messageKey || message == "" {
		message = h.translator.T("en", messageKey)
	}
	if message == "" {
		message = err.Message
	}

	// Логируем оригинальную ошибку для отладки
	if err.Err != nil {
		h.logger.Error("App error occurred", map[string]interface{}{
			"code":    err.Code,
			"message": err.Message,
			"error":   err.Err.Error(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    err.Code,
			"message": message,
		},
	})
}
