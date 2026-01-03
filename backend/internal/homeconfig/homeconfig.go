package homeconfig

import (
	"encoding/json"
	"errors"
	"fmt"
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

type TenantConfig struct {
	Version       string         `json:"version"`
	Hero          Hero           `json:"hero"`
	CategoryCards []CategoryCard `json:"categoryCards"`
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
