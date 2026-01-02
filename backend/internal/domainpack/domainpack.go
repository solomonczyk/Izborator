package domainpack

import (
	"encoding/json"
	"fmt"
	"sync"

	_ "embed"
)

//go:embed domain_pack_v1.json
var rawConfig []byte

type FacetDefinition struct {
	SemanticType string `json:"semantic_type"`
	FacetType    string `json:"facet_type"`
	Values       []string `json:"values,omitempty"`
}

type DomainPack struct {
	Facets []FacetDefinition `json:"facets"`
}

type PackConfig struct {
	Goods    DomainPack `json:"goods"`
	Services DomainPack `json:"services"`
}

var (
	loadedConfig PackConfig
	loadOnce     sync.Once
	loadErr      error
)

func loadConfig() {
	if err := json.Unmarshal(rawConfig, &loadedConfig); err != nil {
		loadErr = fmt.Errorf("failed to parse domain pack config: %w", err)
	}
}

func Facets(domain string) ([]FacetDefinition, error) {
	loadOnce.Do(loadConfig)
	if loadErr != nil {
		return nil, loadErr
	}
	switch domain {
	case "goods":
		return cloneFacets(loadedConfig.Goods.Facets), nil
	case "services":
		return cloneFacets(loadedConfig.Services.Facets), nil
	default:
		return nil, fmt.Errorf("unknown domain: %s", domain)
	}
}

func cloneFacets(source []FacetDefinition) []FacetDefinition {
	if len(source) == 0 {
		return nil
	}
	clone := make([]FacetDefinition, len(source))
	copy(clone, source)
	return clone
}
