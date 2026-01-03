package handlers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	appErrors "github.com/solomonczyk/izborator/internal/errors"
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

type homeHero struct {
	Title             string `json:"title"`
	Subtitle          string `json:"subtitle,omitempty"`
	SearchPlaceholder string `json:"searchPlaceholder"`
	ShowTypeToggle    bool   `json:"showTypeToggle"`
	ShowCitySelect    bool   `json:"showCitySelect"`
	DefaultType       string `json:"defaultType"`
}

type homeCategoryCard struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Hint        string `json:"hint,omitempty"`
	IconKey     string `json:"icon_key,omitempty"`
	Href        string `json:"href"`
	Priority    string `json:"priority,omitempty"`
	Weight      int    `json:"weight,omitempty"`
	Domain      string `json:"domain,omitempty"`
	AnalyticsID string `json:"analytics_id,omitempty"`
}

type homeModel struct {
	Version       string             `json:"version"`
	TenantID      string             `json:"tenant_id"`
	Locale        string             `json:"locale"`
	Hero          homeHero           `json:"hero"`
	CategoryCards []homeCategoryCard `json:"categoryCards"`
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

	model := buildHomeModel(tenantID, locale)
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

func buildHomeModel(tenantID, locale string) homeModel {
	return homeModel{
		Version:  "1",
		TenantID: tenantID,
		Locale:   locale,
		Hero: homeHero{
			Title:             "Find products and services",
			Subtitle:          "Compare offers in one place",
			SearchPlaceholder: "What are you looking for?",
			ShowTypeToggle:    true,
			ShowCitySelect:    false,
			DefaultType:       "all",
		},
		CategoryCards: []homeCategoryCard{
			{
				ID:       "electronics",
				Title:    "Electronics",
				Hint:     "Phones, laptops, gadgets",
				Href:     "/catalog?type=good&category=electronics",
				Priority: "primary",
				Weight:   10,
				Domain:   "good",
			},
			{
				ID:       "food",
				Title:    "Food and drinks",
				Hint:     "Groceries and delivery",
				Href:     "/catalog?type=good&category=food",
				Priority: "primary",
				Weight:   9,
				Domain:   "good",
			},
			{
				ID:     "fashion",
				Title:  "Fashion",
				Hint:   "Clothes and shoes",
				Href:   "/catalog?type=good&category=fashion",
				Weight: 8,
				Domain: "good",
			},
			{
				ID:     "home",
				Title:  "Home and garden",
				Hint:   "Furniture and decor",
				Href:   "/catalog?type=good&category=home",
				Weight: 7,
				Domain: "good",
			},
			{
				ID:     "sport",
				Title:  "Sport and leisure",
				Hint:   "Outdoor and fitness",
				Href:   "/catalog?type=good&category=sport",
				Weight: 6,
				Domain: "good",
			},
			{
				ID:     "auto",
				Title:  "Auto",
				Hint:   "Cars and accessories",
				Href:   "/catalog?type=good&category=auto",
				Weight: 5,
				Domain: "good",
			},
			{
				ID:     "services",
				Title:  "Services",
				Hint:   "Repair, beauty, events",
				Href:   "/catalog?type=service",
				Weight: 4,
				Domain: "service",
			},
			{
				ID:     "finance",
				Title:  "Finance",
				Hint:   "Insurance and banking",
				Href:   "/catalog?type=service&category=finance",
				Weight: 3,
				Domain: "service",
			},
		},
	}
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
