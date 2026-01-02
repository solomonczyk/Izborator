package handlers

import (
	"net/http"

	"github.com/solomonczyk/izborator/internal/cities"
	appErrors "github.com/solomonczyk/izborator/internal/errors"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/http/validation"
)

// CitiesHandler обработчик для работы с городами
type CitiesHandler struct {
	*BaseHandler
	service *cities.Service
}

// NewCitiesHandler создаёт новый обработчик городов
func NewCitiesHandler(service *cities.Service, log *logger.Logger, translator *i18n.Translator) *CitiesHandler {
	return &CitiesHandler{
		BaseHandler: NewBaseHandler(log, translator),
		service:     service,
	}
}

// GetAllActive обрабатывает получение всех активных городов
// GET /api/v1/cities
func (h *CitiesHandler) GetAllActive(w http.ResponseWriter, r *http.Request) {
	tenantID := validation.SanitizeString(r.URL.Query().Get("tenant_id"))
	if tenantID == "" {
		appErr := appErrors.NewValidationError("tenant_id is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	citiesList, err := h.service.GetAllActive()
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to load cities", err)
		h.RespondAppError(w, r, appErr)
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

	h.RespondJSON(w, http.StatusOK, result)
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
