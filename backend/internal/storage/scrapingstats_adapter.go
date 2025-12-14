package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/scrapingstats"
)

// ScrapingStatsAdapter адаптер для работы со статистикой парсинга
type ScrapingStatsAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewScrapingStatsAdapter создаёт новый адаптер статистики
func NewScrapingStatsAdapter(pg *Postgres) scrapingstats.Storage {
	return &ScrapingStatsAdapter{
		pg:  pg,
		ctx: pg.Context(), // Используем контекст из Postgres вместо Background()
	}
}

// SaveStat сохраняет статистику парсинга
func (a *ScrapingStatsAdapter) SaveStat(stat *scrapingstats.ScrapingStat) error {
	statID := uuid.New()
	if stat.ID != "" {
		var err error
		statID, err = uuid.Parse(stat.ID)
		if err != nil {
			return fmt.Errorf("invalid stat ID: %w", err)
		}
	}

	query := `
		INSERT INTO scraping_stats (
			id, shop_id, shop_name, scraped_at, status,
			products_found, products_saved, errors_count,
			error_message, duration_ms, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := a.pg.DB().Exec(a.ctx, query,
		statID, stat.ShopID, stat.ShopName, stat.ScrapedAt, stat.Status,
		stat.ProductsFound, stat.ProductsSaved, stat.ErrorsCount,
		stat.ErrorMessage, stat.DurationMs, stat.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save scraping stat: %w", err)
	}

	return nil
}

// GetShopStats получает статистику по магазину
func (a *ScrapingStatsAdapter) GetShopStats(shopID string, days int) (*scrapingstats.ShopStats, error) {
	fromDate := time.Now().AddDate(0, 0, -days)

	// Получаем статистику из scraping_stats
	query := `
		SELECT 
			COUNT(*) as total_scrapes,
			COUNT(*) FILTER (WHERE status = 'success') as success_count,
			COUNT(*) FILTER (WHERE status = 'error') as error_count,
			SUM(products_saved) as total_products,
			AVG(duration_ms)::INTEGER as avg_duration_ms
		FROM scraping_stats
		WHERE shop_id = $1 AND scraped_at >= $2
	`

	var stats scrapingstats.ShopStats
	stats.ShopID = shopID

	err := a.pg.DB().QueryRow(a.ctx, query, shopID, fromDate).Scan(
		&stats.TotalScrapes,
		&stats.SuccessCount,
		&stats.ErrorCount,
		&stats.TotalProducts,
		&stats.AvgDurationMs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop stats: %w", err)
	}

	// Получаем информацию о магазине
	shopQuery := `
		SELECT name, last_scraped_at, is_active
		FROM shops
		WHERE id = $1
	`

	var lastScrapedAt *time.Time
	err = a.pg.DB().QueryRow(a.ctx, shopQuery, shopID).Scan(
		&stats.ShopName,
		&lastScrapedAt,
		&stats.ScrapingEnabled,
	)
	if err != nil {
		// Магазин может не существовать, продолжаем
		stats.ShopName = shopID
	} else {
		stats.LastScrapedAt = lastScrapedAt
	}

	return &stats, nil
}

// GetOverallStats получает общую статистику
func (a *ScrapingStatsAdapter) GetOverallStats(days int) (*scrapingstats.OverallStats, error) {
	fromDate := time.Now().AddDate(0, 0, -days)

	// Общая статистика
	query := `
		SELECT 
			COUNT(DISTINCT shop_id) as total_shops,
			COUNT(*) as total_scrapes,
			COUNT(*) FILTER (WHERE status = 'success') as success_count,
			SUM(products_saved) as total_products,
			MAX(scraped_at) as last_scrape_at
		FROM scraping_stats
		WHERE scraped_at >= $1
	`

	var stats scrapingstats.OverallStats
	var successCount int
	var lastScrapeAt *time.Time

	err := a.pg.DB().QueryRow(a.ctx, query, fromDate).Scan(
		&stats.TotalShops,
		&stats.TotalScrapes,
		&successCount,
		&stats.TotalProducts,
		&lastScrapeAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	stats.LastScrapeAt = lastScrapeAt

	// Рассчитываем success rate
	if stats.TotalScrapes > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalScrapes) * 100
	}

	// Получаем количество активных магазинов
	activeQuery := `
		SELECT COUNT(*) 
		FROM shops 
		WHERE is_active = true
	`
	err = a.pg.DB().QueryRow(a.ctx, activeQuery).Scan(&stats.ActiveShops)
	if err != nil {
		stats.ActiveShops = 0
	}

	// Получаем последние записи
	recent, _ := a.GetRecentStats(10)
	stats.RecentStats = recent

	return &stats, nil
}

// GetRecentStats получает последние N записей статистики
func (a *ScrapingStatsAdapter) GetRecentStats(limit int) ([]*scrapingstats.ScrapingStat, error) {
	query := `
		SELECT id, shop_id, shop_name, scraped_at, status,
		       products_found, products_saved, errors_count,
		       error_message, duration_ms, created_at
		FROM scraping_stats
		ORDER BY scraped_at DESC
		LIMIT $1
	`

	rows, err := a.pg.DB().Query(a.ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent stats: %w", err)
	}
	defer rows.Close()

	var stats []*scrapingstats.ScrapingStat
	for rows.Next() {
		var stat scrapingstats.ScrapingStat
		var errorMsg *string

		err := rows.Scan(
			&stat.ID,
			&stat.ShopID,
			&stat.ShopName,
			&stat.ScrapedAt,
			&stat.Status,
			&stat.ProductsFound,
			&stat.ProductsSaved,
			&stat.ErrorsCount,
			&errorMsg,
			&stat.DurationMs,
			&stat.CreatedAt,
		)
		if err != nil {
			continue
		}

		if errorMsg != nil {
			stat.ErrorMessage = *errorMsg
		}

		stats = append(stats, &stat)
	}

	return stats, nil
}

// UpdateShopLastScraped обновляет last_scraped_at для магазина
func (a *ScrapingStatsAdapter) UpdateShopLastScraped(shopID string) error {
	query := `
		UPDATE shops
		SET last_scraped_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := a.pg.DB().Exec(a.ctx, query, shopID)
	if err != nil {
		return fmt.Errorf("failed to update shop last_scraped_at: %w", err)
	}

	return nil
}

