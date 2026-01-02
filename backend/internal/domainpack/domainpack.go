package domainpack

import (
	"encoding/json"
	"fmt"
	"sort"
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

type PackConfig map[string]DomainPack

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
	pack, ok := loadedConfig[domain]
	if !ok {
		return nil, fmt.Errorf("unknown domain: %s", domain)
	}
	return cloneFacets(pack.Facets), nil
}

func HasDomain(domain string) bool {
	loadOnce.Do(loadConfig)
	if loadErr != nil {
		return false
	}
	_, ok := loadedConfig[domain]
	return ok
}

func Domains() []string {
	loadOnce.Do(loadConfig)
	if loadErr != nil {
		return nil
	}
	domains := make([]string, 0, len(loadedConfig))
	for domain := range loadedConfig {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	return domains
}

func cloneFacets(source []FacetDefinition) []FacetDefinition {
	if len(source) == 0 {
		return nil
	}
	clone := make([]FacetDefinition, len(source))
	copy(clone, source)
	return clone
}
