package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/solomonczyk/izborator/internal/cities"
	appErrors "github.com/solomonczyk/izborator/internal/errors"
	httpMiddleware "github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
)

// CitiesHandler обработчик для работы с городами
type CitiesHandler struct {
	service    *cities.Service
	logger     *logger.Logger
	translator *i18n.Translator
}

// NewCitiesHandler создаёт новый обработчик городов
func NewCitiesHandler(service *cities.Service, log *logger.Logger, translator *i18n.Translator) *CitiesHandler {
	return &CitiesHandler{
		service:    service,
		logger:     log,
		translator: translator,
	}
}

// GetAllActive обрабатывает получение всех активных городов
// GET /api/v1/cities
func (h *CitiesHandler) GetAllActive(w http.ResponseWriter, r *http.Request) {
	citiesList, err := h.service.GetAllActive()
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to load cities", err)
		h.respondAppError(w, r, appErr)
		return
	}

	// Преобразуем в JSON
	result := make([]CityResponse, 0, len(citiesList))
	for _, city := range citiesList {
		result = append(result, CityResponse{
			ID:        city.ID,
			Slug:      city.Slug,
			NameSr:    city.NameSr,
			RegionSr:  city.RegionSr, // *string - правильно
			SortOrder: city.SortOrder,
			IsActive:  city.IsActive,
		})
	}

	h.respondJSON(w, http.StatusOK, result)
}

// CityResponse структура ответа для города
type CityResponse struct {
	ID        string  `json:"id"`
	Slug      string  `json:"slug"`
	NameSr    string  `json:"name_sr"`
	RegionSr  *string `json:"region_sr,omitempty"`
	SortOrder int     `json:"sort_order"`
	IsActive  bool    `json:"is_active"`
}

// respondJSON отправляет JSON ответ
func (h *CitiesHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error": err,
		})
	}
}

// respondAppError отправляет JSON ошибку из AppError
func (h *CitiesHandler) respondAppError(w http.ResponseWriter, r *http.Request, err *appErrors.AppError) {
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
