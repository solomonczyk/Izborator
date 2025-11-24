package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/scraper"
)

// ScraperAdapter адаптер для работы с парсингом
type ScraperAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewScraperAdapter создаёт новый адаптер для парсинга
func NewScraperAdapter(pg *Postgres) scraper.Storage {
	return &ScraperAdapter{
		pg:  pg,
		ctx: pg.Context(),
	}
}

// SaveRawProduct сохраняет сырые данные товара
func (a *ScraperAdapter) SaveRawProduct(data *scraper.RawProduct) error {
	query := `
		INSERT INTO raw_products (
			id, shop_id, shop_name, external_id, name, description, price, currency,
			url, image_urls, category, brand, specs, in_stock, scraped_at, processed
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (shop_id, external_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			url = EXCLUDED.url,
			image_urls = EXCLUDED.image_urls,
			category = EXCLUDED.category,
			brand = EXCLUDED.brand,
			specs = EXCLUDED.specs,
			in_stock = EXCLUDED.in_stock,
			scraped_at = EXCLUDED.scraped_at,
			processed = false
	`

	rawID := uuid.New()

	// Сериализация JSON полей
	imageURLsJSON, err := json.Marshal(data.ImageURLs)
	if err != nil {
		return fmt.Errorf("failed to marshal image_urls: %w", err)
	}

	specsJSON, err := json.Marshal(data.Specs)
	if err != nil {
		return fmt.Errorf("failed to marshal specs: %w", err)
	}

	if data.ScrapedAt.IsZero() {
		data.ScrapedAt = time.Now()
	}

	_, err = a.pg.DB().Exec(a.ctx, query,
		rawID,
		data.ShopID,
		data.ShopName,
		data.ExternalID,
		data.Name,
		data.Description,
		data.Price,
		data.Currency,
		data.URL,
		imageURLsJSON,
		data.Category,
		data.Brand,
		specsJSON,
		data.InStock,
		data.ScrapedAt,
		false, // processed
	)

	if err != nil {
		return fmt.Errorf("failed to save raw product: %w", err)
	}

	return nil
}

// GetShopConfig получает конфигурацию магазина
func (a *ScraperAdapter) GetShopConfig(shopID string) (*scraper.ShopConfig, error) {
	query := `
		SELECT id, name, base_url, selectors, rate_limit, enabled
		FROM shops
		WHERE id = $1
	`

	var config scraper.ShopConfig
	var selectorsJSON []byte

	err := a.pg.DB().QueryRow(a.ctx, query, shopID).Scan(
		&config.ID,
		&config.Name,
		&config.BaseURL,
		&selectorsJSON,
		&config.RateLimit,
		&config.Enabled,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, scraper.ErrShopNotFound
		}
		return nil, fmt.Errorf("failed to get shop config: %w", err)
	}

	// Десериализация JSONB selectors
	if len(selectorsJSON) > 0 {
		if err := json.Unmarshal(selectorsJSON, &config.Selectors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal selectors: %w", err)
		}
	} else {
		config.Selectors = make(map[string]string)
	}

	return &config, nil
}

// ListShops получает список всех магазинов
func (a *ScraperAdapter) ListShops() ([]*scraper.ShopConfig, error) {
	query := `
		SELECT id, name, base_url, selectors, rate_limit, enabled
		FROM shops
		ORDER BY name
	`

	rows, err := a.pg.DB().Query(a.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list shops: %w", err)
	}
	defer rows.Close()

	var shops []*scraper.ShopConfig

	for rows.Next() {
		var shop scraper.ShopConfig
		var selectorsJSON []byte

		err := rows.Scan(
			&shop.ID,
			&shop.Name,
			&shop.BaseURL,
			&selectorsJSON,
			&shop.RateLimit,
			&shop.Enabled,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan shop: %w", err)
		}

		// Десериализация JSONB selectors
		if len(selectorsJSON) > 0 {
			if err := json.Unmarshal(selectorsJSON, &shop.Selectors); err != nil {
				return nil, fmt.Errorf("failed to unmarshal selectors: %w", err)
			}
		} else {
			shop.Selectors = make(map[string]string)
		}

		shops = append(shops, &shop)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shops: %w", err)
	}

	return shops, nil
}
