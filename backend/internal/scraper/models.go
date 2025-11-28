package scraper

import "time"

// RawProduct сырые данные товара с сайта магазина
type RawProduct struct {
	ShopID      string                 `json:"shop_id"`
	ShopName    string                 `json:"shop_name"`
	ExternalID  string                 `json:"external_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Price       float64                `json:"price"`
	Currency    string                 `json:"currency"`
	URL         string                 `json:"url"`
	ImageURLs   []string               `json:"image_urls"`
	Category    string                 `json:"category"`
	Brand       string                 `json:"brand"`
	Specs       map[string]string      `json:"specs"`
	RawPayload  map[string]interface{} `json:"raw_payload"` // полный сырой объект
	InStock     bool                   `json:"in_stock"`
	ParsedAt    time.Time              `json:"parsed_at"` // переименовано из ScrapedAt
	// ScrapedAt оставлено для обратной совместимости, но используем ParsedAt
	ScrapedAt time.Time `json:"scraped_at"` // deprecated, используй ParsedAt
}

// ShopConfig конфигурация магазина для парсинга
type ShopConfig struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	BaseURL     string   `json:"base_url"`
	Selectors   map[string]string `json:"selectors"`
	RateLimit   int      `json:"rate_limit"` // запросов в секунду
	Enabled     bool     `json:"enabled"`
}

