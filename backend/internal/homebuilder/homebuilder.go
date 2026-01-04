package homebuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"sync"

	_ "embed"

	"github.com/solomonczyk/izborator/internal/categorytree"
	"github.com/solomonczyk/izborator/internal/homeconfig"
)

//go:embed featured_v1.json
var rawFeatured []byte

type FeaturedCategorySpec struct {
	CategoryID string `json:"category_id"`
	Priority   string `json:"priority"`
	Order      int    `json:"order"`
	IconKey    string `json:"icon_key,omitempty"`
	Type       string `json:"type,omitempty"`
}

type TenantFeaturedConfig struct {
	FeaturedCategories []FeaturedCategorySpec `json:"featured_categories"`
}

type FeaturedConfig struct {
	Version string                          `json:"version"`
	Tenants map[string]TenantFeaturedConfig `json:"tenants"`
}

type FeaturedCategory struct {
	CategoryID string `json:"category_id"`
	Title      string `json:"title"`
	Href       string `json:"href"`
	Priority   string `json:"priority"`
	Order      int    `json:"order"`
	IconKey    string `json:"icon_key,omitempty"`
}

type SearchPreset struct {
	Label      string `json:"label"`
	Query      string `json:"query"`
	Type       string `json:"type,omitempty"`
	CategoryID string `json:"category_id,omitempty"`
}

type HomeModel struct {
	Version            string             `json:"version"`
	TenantID           string             `json:"tenant_id"`
	Locale             string             `json:"locale"`
	Hero               homeconfig.Hero    `json:"hero"`
	FeaturedCategories []FeaturedCategory `json:"featuredCategories"`
	SearchPresets      []SearchPreset     `json:"searchPresets,omitempty"`
}

var (
	loadOnce sync.Once
	loadErr  error
	config   FeaturedConfig
)

func BuildHomeModel(tenantID, locale string) (HomeModel, error) {
	if tenantID == "" {
		return HomeModel{}, errors.New("tenant_id is required")
	}

	tenantConfig, err := resolveFeaturedConfig(tenantID)
	if err != nil {
		return HomeModel{}, err
	}

	homeConfig, err := homeconfig.Resolve(tenantID, locale)
	if err != nil {
		return HomeModel{}, err
	}

	tree, err := categorytree.Load()
	if err != nil {
		return HomeModel{}, err
	}

	nodeIndex := make(map[string]categorytree.Node)
	indexNodes(tree.Categories, nodeIndex)

	featured := make([]FeaturedCategory, 0, len(tenantConfig.FeaturedCategories))
	for _, spec := range tenantConfig.FeaturedCategories {
		node, ok := nodeIndex[spec.CategoryID]
		if !ok {
			continue
		}
		featured = append(featured, FeaturedCategory{
			CategoryID: spec.CategoryID,
			Title:      node.Title,
			Href:       buildHref(spec),
			Priority:   spec.Priority,
			Order:      spec.Order,
			IconKey:    spec.IconKey,
		})
	}

	sort.SliceStable(featured, func(i, j int) bool {
		return featured[i].Order < featured[j].Order
	})

	return HomeModel{
		Version:            "2",
		TenantID:           tenantID,
		Locale:             locale,
		Hero:               homeConfig.Hero,
		FeaturedCategories: featured,
	}, nil
}

func resolveFeaturedConfig(tenantID string) (TenantFeaturedConfig, error) {
	loadOnce.Do(load)
	if loadErr != nil {
		return TenantFeaturedConfig{}, loadErr
	}
	tenantConfig, ok := config.Tenants[tenantID]
	if !ok {
		return TenantFeaturedConfig{}, homeconfig.ErrTenantNotFound
	}
	return tenantConfig, nil
}

func load() {
	if err := json.Unmarshal(rawFeatured, &config); err != nil {
		loadErr = fmt.Errorf("failed to parse featured config: %w", err)
	}
}

func indexNodes(nodes []categorytree.Node, index map[string]categorytree.Node) {
	for _, node := range nodes {
		index[node.ID] = node
		if len(node.Children) > 0 {
			indexNodes(node.Children, index)
		}
	}
}

func buildHref(spec FeaturedCategorySpec) string {
	escaped := url.QueryEscape(spec.CategoryID)
	if spec.Type != "" && spec.Type != "all" {
		return "/catalog?type=" + url.QueryEscape(spec.Type) + "&category=" + escaped
	}
	return "/catalog?category=" + escaped
}
