package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/solomonczyk/izborator/internal/cities"
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
		h.logger.Error("GetAllActive cities failed", map[string]interface{}{
			"error": err.Error(),
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.cities_load_failed")
		return
	}

	// Преобразуем в JSON
	result := make([]CityResponse, 0, len(citiesList))
	for _, city := range citiesList {
		result = append(result, CityResponse{
			ID:        city.ID,
			Slug:      city.Slug,
			NameSr:    city.NameSr,
			RegionSr:  city.RegionSr,
			SortOrder: city.SortOrder,
			IsActive:  city.IsActive,
		})
	}

	h.respondJSON(w, http.StatusOK, result)
}

// CityResponse ответ с данными города
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error": err,
		})
	}
}

// respondError отправляет JSON ошибку
func (h *CitiesHandler) respondError(w http.ResponseWriter, r *http.Request, status int, key string) {
	lang := httpMiddleware.GetLangFromContext(r.Context())
	message := h.translator.T(lang, key)
	if message == key || message == "" {
		// fallback на английский
		message = h.translator.T("en", key)
		if message == "" {
			message = key
		}
	}
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
