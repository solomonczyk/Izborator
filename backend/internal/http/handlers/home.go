package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	appErrors "github.com/solomonczyk/izborator/internal/errors"
	"github.com/solomonczyk/izborator/internal/homeconfig"
	"github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/http/validation"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
)

type HomeHandler struct {
	*BaseHandler
}

func NewHomeHandler(log *logger.Logger, translator *i18n.Translator) *HomeHandler {
	return &HomeHandler{
		BaseHandler: NewBaseHandler(log, translator),
	}
}

type homeHero = homeconfig.Hero

type homeCategoryCard = homeconfig.CategoryCard

type homeModel struct {
	Version       string             `json:"version"`
	TenantID      string             `json:"tenant_id"`
	Locale        string             `json:"locale"`
	Hero          homeHero           `json:"hero"`
	CategoryCards []homeCategoryCard `json:"categoryCards"`
}

type homeMeta struct {
	Version        string `json:"version"`
	TenantID       string `json:"tenant_id"`
	Locale         string `json:"locale"`
	CardsCount     int    `json:"cards_count"`
	ShowTypeToggle bool   `json:"showTypeToggle"`
	ShowCitySelect bool   `json:"showCitySelect"`
	DefaultType    string `json:"defaultType"`
}

func (h *HomeHandler) GetHome(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	tenantID := validation.SanitizeString(r.URL.Query().Get("tenant_id"))
	if tenantID == "" {
		appErr := appErrors.NewValidationError("tenant_id is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	locale := validation.SanitizeString(r.URL.Query().Get("locale"))
	if locale == "" {
		locale = middleware.GetLangFromContext(r.Context())
		if locale == "" {
			locale = "en"
		}
	}

	model, err := buildHomeModel(tenantID, locale)
	if err != nil {
		if errors.Is(err, homeconfig.ErrTenantNotFound) {
			appErr := appErrors.NewNotFound("home config not found for tenant")
			h.RespondAppError(w, r, appErr)
			return
		}
		appErr := appErrors.NewInternalError("Failed to load home config", err)
		h.RespondAppError(w, r, appErr)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60, s-maxage=300, stale-while-revalidate=600")
	h.RespondJSON(w, http.StatusOK, model)

	ms := time.Since(start).Milliseconds()
	cardsCount := len(model.CategoryCards)
	warnMs := parseWarnMsEnv("HOME_MODEL_WARN_MS", 1500)
	fields := map[string]interface{}{
		"event":       "home_model",
		"tenant_id":   tenantID,
		"locale":      locale,
		"cards_count": cardsCount,
		"ms":          ms,
	}
	if warnMs > 0 && ms > int64(warnMs) {
		h.logger.Warn("home model slow response", fields)
	} else {
		h.logger.Info("home model response", fields)
	}
}

func (h *HomeHandler) GetHomeMeta(w http.ResponseWriter, r *http.Request) {
	tenantID := validation.SanitizeString(r.URL.Query().Get("tenant_id"))
	if tenantID == "" {
		appErr := appErrors.NewValidationError("tenant_id is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	locale := validation.SanitizeString(r.URL.Query().Get("locale"))
	if locale == "" {
		locale = middleware.GetLangFromContext(r.Context())
		if locale == "" {
			locale = "en"
		}
	}

	model, err := buildHomeModel(tenantID, locale)
	if err != nil {
		if errors.Is(err, homeconfig.ErrTenantNotFound) {
			appErr := appErrors.NewNotFound("home config not found for tenant")
			h.RespondAppError(w, r, appErr)
			return
		}
		appErr := appErrors.NewInternalError("Failed to load home config", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	meta := homeMeta{
		Version:        model.Version,
		TenantID:       model.TenantID,
		Locale:         model.Locale,
		CardsCount:     len(model.CategoryCards),
		ShowTypeToggle: model.Hero.ShowTypeToggle,
		ShowCitySelect: model.Hero.ShowCitySelect,
		DefaultType:    model.Hero.DefaultType,
	}

	w.Header().Set("Cache-Control", "public, max-age=60, s-maxage=300, stale-while-revalidate=600")
	h.RespondJSON(w, http.StatusOK, meta)
}

func buildHomeModel(tenantID, locale string) (homeModel, error) {
	config, err := homeconfig.Resolve(tenantID, locale)
	if err != nil {
		return homeModel{}, err
	}
	return homeModel{
		Version:       config.Version,
		TenantID:      tenantID,
		Locale:        locale,
		Hero:          config.Hero,
		CategoryCards: config.CategoryCards,
	}, nil
}

func parseWarnMsEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
