package autoconfig

// Storage интерфейс для работы с БД (кандидаты и магазины)
type Storage interface {
	GetClassifiedCandidates(limit int) ([]Candidate, error)
	MarkAsConfigured(id string, config ShopConfig) error
	MarkAsFailed(id string, reason string) error
}

// Candidate кандидат на магазин для авто-конфигурации
type Candidate struct {
	ID       string
	Domain   string
	SiteType string // "ecommerce" | "service_provider" | "unknown"
}

// ShopConfig конфигурация магазина с селекторами
type ShopConfig struct {
	Selectors map[string]string `json:"selectors"`
}
