package scrapingstats

import "time"

// ScrapingStat статистика одного парсинга
type ScrapingStat struct {
	ID           string    `json:"id"`
	ShopID       string    `json:"shop_id"`
	ShopName     string    `json:"shop_name"`
	ScrapedAt    time.Time `json:"scraped_at"`
	Status       string    `json:"status"` // success, error, partial
	ProductsFound int      `json:"products_found"`
	ProductsSaved int      `json:"products_saved"`
	ErrorsCount   int      `json:"errors_count"`
	ErrorMessage  string   `json:"error_message,omitempty"`
	DurationMs    int      `json:"duration_ms"`
	CreatedAt     time.Time `json:"created_at"`
}

// ShopStats статистика по магазину
type ShopStats struct {
	ShopID          string    `json:"shop_id"`
	ShopName        string    `json:"shop_name"`
	LastScrapedAt   *time.Time `json:"last_scraped_at,omitempty"`
	TotalScrapes    int       `json:"total_scrapes"`
	SuccessCount    int       `json:"success_count"`
	ErrorCount       int       `json:"error_count"`
	TotalProducts    int       `json:"total_products"`
	AvgDurationMs   int       `json:"avg_duration_ms"`
	ScrapingEnabled bool      `json:"scraping_enabled"`
}

// OverallStats общая статистика парсинга
type OverallStats struct {
	TotalShops        int       `json:"total_shops"`
	ActiveShops       int       `json:"active_shops"`
	TotalScrapes      int       `json:"total_scrapes"`
	SuccessRate       float64   `json:"success_rate"` // процент успешных парсингов
	TotalProducts     int       `json:"total_products"`
	LastScrapeAt      *time.Time `json:"last_scrape_at,omitempty"`
	RecentStats       []*ScrapingStat `json:"recent_stats,omitempty"`
}

