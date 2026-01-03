package homeconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	_ "embed"
)

//go:embed home_config_v1.json
var rawConfig []byte

var ErrTenantNotFound = errors.New("tenant not found")

type Hero struct {
	Title             string `json:"title"`
	Subtitle          string `json:"subtitle,omitempty"`
	SearchPlaceholder string `json:"searchPlaceholder"`
	ShowTypeToggle    bool   `json:"showTypeToggle"`
	ShowCitySelect    bool   `json:"showCitySelect"`
	DefaultType       string `json:"defaultType"`
}

type CategoryCard struct {
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

type TenantLocaleConfig struct {
	Hero          *Hero          `json:"hero,omitempty"`
	CategoryCards []CategoryCard `json:"categoryCards,omitempty"`
}

type TenantConfig struct {
	Version       string         `json:"version"`
	Hero          Hero           `json:"hero"`
	CategoryCards []CategoryCard `json:"categoryCards"`
	Locales       map[string]TenantLocaleConfig `json:"locales,omitempty"`
}

type Config map[string]TenantConfig

var (
	loadedConfig Config
	loadOnce     sync.Once
	loadErr      error
)

func loadConfig() {
	if err := json.Unmarshal(rawConfig, &loadedConfig); err != nil {
		loadErr = fmt.Errorf("failed to parse home config: %w", err)
	}
}

func Get(tenantID string) (TenantConfig, error) {
	loadOnce.Do(loadConfig)
	if loadErr != nil {
		return TenantConfig{}, loadErr
	}
	tenantConfig, ok := loadedConfig[tenantID]
	if !ok {
		return TenantConfig{}, ErrTenantNotFound
	}
	return tenantConfig, nil
}

func Resolve(tenantID, locale string) (TenantConfig, error) {
	config, err := Get(tenantID)
	if err != nil {
		return TenantConfig{}, err
	}
	if locale == "" || len(config.Locales) == 0 {
		return config, nil
	}

	normalized := strings.ToLower(locale)
	if override, ok := config.Locales[normalized]; ok {
		return applyLocale(config, override), nil
	}
	if parts := strings.SplitN(normalized, "-", 2); len(parts) > 1 {
		if override, ok := config.Locales[parts[0]]; ok {
			return applyLocale(config, override), nil
		}
	}

	return config, nil
}

func applyLocale(config TenantConfig, override TenantLocaleConfig) TenantConfig {
	if override.Hero != nil {
		config.Hero = *override.Hero
	}
	if len(override.CategoryCards) > 0 {
		config.CategoryCards = override.CategoryCards
	}
	return config
}
