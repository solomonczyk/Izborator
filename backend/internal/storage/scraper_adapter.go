package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/scraper"
)

// ScraperAdapter адаптер для работы с парсингом
type ScraperAdapter struct {
	*BaseAdapter
}

// NewScraperAdapter создаёт новый адаптер для парсинга
func NewScraperAdapter(pg *Postgres) scraper.Storage {
	return &ScraperAdapter{
		BaseAdapter: NewBaseAdapter(pg, nil),
	}
}

// SaveRawProduct сохраняет сырые данные товара в raw_products
func (a *ScraperAdapter) SaveRawProduct(data *scraper.RawProduct) error {
	// Определяем время парсинга
	parsedAt := data.ParsedAt
	if parsedAt.IsZero() {
		parsedAt = time.Now()
	}

	// Сериализация JSON полей
	specsJSON, err := json.Marshal(data.Specs)
	if err != nil {
		return fmt.Errorf("failed to marshal specs: %w", err)
	}

	var rawPayloadJSON []byte
	if data.RawPayload != nil {
		rawPayloadJSON, err = json.Marshal(data.RawPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal raw_payload: %w", err)
		}
	}

	// Сериализация image_urls
	var imageURLsJSON []byte
	if len(data.ImageURLs) > 0 {
		imageURLsJSON, err = json.Marshal(data.ImageURLs)
		if err != nil {
			return fmt.Errorf("failed to marshal image_urls: %w", err)
		}
	}

	query := `
		INSERT INTO raw_products (
			shop_id,
			shop_name,
			external_id,
			url,
			name,
			description,
			brand,
			category,
			price,
			currency,
			image_urls,
			specs_json,
			raw_payload,
			in_stock,
			parsed_at,
			processed
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, FALSE
		)
		ON CONFLICT (shop_id, external_id)
		DO UPDATE SET
			shop_name   = EXCLUDED.shop_name,
			url         = EXCLUDED.url,
			name        = EXCLUDED.name,
			description = EXCLUDED.description,
			brand       = EXCLUDED.brand,
			category    = EXCLUDED.category,
			price       = EXCLUDED.price,
			currency    = EXCLUDED.currency,
			image_urls  = EXCLUDED.image_urls,
			specs_json  = EXCLUDED.specs_json,
			raw_payload = EXCLUDED.raw_payload,
			in_stock    = EXCLUDED.in_stock,
			parsed_at   = EXCLUDED.parsed_at,
			processed   = FALSE,
			processed_at = NULL
	`

	_, err = a.pg.DB().Exec(a.GetContext(), query,
		data.ShopID,
		data.ShopName,
		data.ExternalID,
		data.URL,
		data.Name,
		data.Description,
		data.Brand,
		data.Category,
		data.Price,
		data.Currency,
		imageURLsJSON,
		specsJSON,
		rawPayloadJSON,
		data.InStock,
		parsedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save raw product: %w", err)
	}

	return nil
}

// GetShopConfig получает конфигурацию магазина
func (a *ScraperAdapter) GetShopConfig(shopID string) (*scraper.ShopConfig, error) {
	query := `
		SELECT 
			id,
			name,
			base_url,
			selectors,
			rate_limit,
			is_active,
			COALESCE(retry_limit, 3) AS retry_limit,
			COALESCE(retry_backoff_ms, 3000) AS retry_backoff_ms
		FROM shops
		WHERE id = $1
	`

	var config scraper.ShopConfig
	var selectorsJSON []byte
	var isActive bool

	err := a.pg.DB().QueryRow(a.GetContext(), query, shopID).Scan(
		&config.ID,
		&config.Name,
		&config.BaseURL,
		&selectorsJSON,
		&config.RateLimit,
		&isActive,
		&config.RetryLimit,
		&config.RetryBackoffMs,
	)

	config.Enabled = isActive

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
		SELECT 
			id,
			name,
			base_url,
			selectors,
			rate_limit,
			is_active,
			COALESCE(retry_limit, 3) AS retry_limit,
			COALESCE(retry_backoff_ms, 3000) AS retry_backoff_ms
		FROM shops
		ORDER BY name
	`

	rows, err := a.pg.DB().Query(a.GetContext(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to list shops: %w", err)
	}
	defer rows.Close()

	var shops []*scraper.ShopConfig

	for rows.Next() {
		var shop scraper.ShopConfig
		var selectorsJSON []byte
		var isActive bool

		err := rows.Scan(
			&shop.ID,
			&shop.Name,
			&shop.BaseURL,
			&selectorsJSON,
			&shop.RateLimit,
			&isActive,
			&shop.RetryLimit,
			&shop.RetryBackoffMs,
		)

		shop.Enabled = isActive

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

// GetShopDefaultCityID returns default city id for a shop.
func (a *ScraperAdapter) GetShopDefaultCityID(shopID string) (*string, error) {
	query := `
		SELECT
			default_city_id
		FROM shops
		WHERE id = $1
	`

	var cityID *string
	err := a.pg.DB().QueryRow(a.GetContext(), query, shopID).Scan(&cityID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("shop not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get shop default city id: %w", err)
	}

	return cityID, nil
}

// GetUnprocessedRawProducts возвращает батч необработанных сырых товаров
func (a *ScraperAdapter) GetUnprocessedRawProducts(limit int) ([]*scraper.RawProduct, error) {
	query := `
		SELECT
			shop_id,
			shop_name,
			external_id,
			url,
			name,
			description,
			brand,
			category,
			price,
			currency,
			image_urls,
			specs_json,
			raw_payload,
			in_stock,
			parsed_at
		FROM raw_products
		WHERE processed = FALSE
		ORDER BY parsed_at ASC
		LIMIT $1
	`

	rows, err := a.pg.DB().Query(a.GetContext(), query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed raw products: %w", err)
	}
	defer rows.Close()

	var result []*scraper.RawProduct

	for rows.Next() {
		var (
			r             scraper.RawProduct
			imageURLsJSON []byte
			specsJSON     []byte
			rawPayload    []byte
			parsedAtTime  time.Time
			description   *string
			brand         *string
			category      *string
			url           *string
		)

		if err := rows.Scan(
			&r.ShopID,
			&r.ShopName,
			&r.ExternalID,
			&url,
			&r.Name,
			&description,
			&brand,
			&category,
			&r.Price,
			&r.Currency,
			&imageURLsJSON,
			&specsJSON,
			&rawPayload,
			&r.InStock,
			&parsedAtTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan raw product: %w", err)
		}

		// Обработка nullable полей
		if description != nil {
			r.Description = *description
		}
		if brand != nil {
			r.Brand = *brand
		}
		if category != nil {
			r.Category = *category
		}
		if url != nil {
			r.URL = *url
		}

		// Десериализация JSON полей
		if len(imageURLsJSON) > 0 {
			if err := json.Unmarshal(imageURLsJSON, &r.ImageURLs); err != nil {
				// Не критично, продолжаем с пустым массивом
				r.ImageURLs = []string{}
			}
		} else {
			r.ImageURLs = []string{}
		}

		if len(specsJSON) > 0 {
			if err := json.Unmarshal(specsJSON, &r.Specs); err != nil {
				// Не критично, продолжаем с пустым specs
				r.Specs = make(map[string]string)
			}
		} else {
			r.Specs = make(map[string]string)
		}

		if len(rawPayload) > 0 {
			if err := json.Unmarshal(rawPayload, &r.RawPayload); err != nil {
				// Не критично, продолжаем с nil
				r.RawPayload = nil
			}
		}

		r.ParsedAt = parsedAtTime
		r.ScrapedAt = parsedAtTime // для обратной совместимости

		result = append(result, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating raw products: %w", err)
	}

	return result, nil
}

// MarkRawProductAsProcessed помечает сырой товар как обработанный
func (a *ScraperAdapter) MarkRawProductAsProcessed(shopID, externalID string) error {
	query := `
		UPDATE raw_products
		SET processed = TRUE,
		    processed_at = NOW()
		WHERE shop_id = $1
		  AND external_id = $2
	`

	result, err := a.pg.DB().Exec(a.GetContext(), query, shopID, externalID)
	if err != nil {
		return fmt.Errorf("failed to mark raw product as processed: %w", err)
	}

	// Проверяем, что запись была обновлена
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected: shop_id=%s, external_id=%s", shopID, externalID)
	}

	return nil
}

