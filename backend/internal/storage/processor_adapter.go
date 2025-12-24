package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/scraper"
)

// ProcessorAdapter адаптер для работы с обработкой сырых данных
type ProcessorAdapter struct {
	pg    *Postgres
	meili *Meilisearch
	ctx   context.Context
}

// NewProcessorAdapter создаёт новый адаптер для обработки
func NewProcessorAdapter(pg *Postgres, meili *Meilisearch) *ProcessorAdapter {
	return &ProcessorAdapter{
		pg:    pg,
		meili: meili,
		ctx:   pg.Context(),
	}
}

// ScraperStorage интерфейс для чтения сырых данных (реализуется ScraperAdapter)
type ScraperStorage interface {
	GetUnprocessedRawProducts(limit int) ([]*scraper.RawProduct, error)
	MarkRawProductAsProcessed(shopID, externalID string) error
}

// Storage интерфейс для записи обработанных данных
type ProcessorStorage interface {
	SaveProduct(product *products.Product) error
	SavePrice(price *products.ProductPrice) error
}

// SaveProduct сохраняет товар в products
func (a *ProcessorAdapter) SaveProduct(product *products.Product) error {
	var productUUID uuid.UUID
	var err error

	// Если ID пустой, создаём новый UUID
	if product.ID == "" {
		productUUID = uuid.New()
		product.ID = productUUID.String()
	} else {
		productUUID, err = uuid.Parse(product.ID)
		if err != nil {
			return fmt.Errorf("invalid product ID: %w", err)
		}
	}

	// Сериализация specs в JSON
	specsJSON, err := json.Marshal(product.Specs)
	if err != nil {
		return fmt.Errorf("failed to marshal specs: %w", err)
	}

	query := `
		INSERT INTO products (id, name, description, brand, category, image_url, specs, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			brand = EXCLUDED.brand,
			category = EXCLUDED.category,
			image_url = EXCLUDED.image_url,
			specs = EXCLUDED.specs,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}
	product.UpdatedAt = now

	_, err = a.pg.DB().Exec(a.ctx, query,
		productUUID,
		product.Name,
		product.Description,
		product.Brand,
		product.Category,
		product.ImageURL,
		specsJSON,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	return nil
}

// SavePrice сохраняет цену товара в product_prices
func (a *ProcessorAdapter) SavePrice(price *products.ProductPrice) error {
	productUUID, err := uuid.Parse(price.ProductID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		INSERT INTO product_prices (product_id, shop_id, shop_name, price, currency, url, in_stock, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (product_id, shop_id) DO UPDATE SET
			shop_name = EXCLUDED.shop_name,
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			url = EXCLUDED.url,
			in_stock = EXCLUDED.in_stock,
			updated_at = EXCLUDED.updated_at
	`

	price.UpdatedAt = time.Now()

	_, err = a.pg.DB().Exec(a.ctx, query,
		productUUID,
		price.ShopID,
		price.ShopName,
		price.Price,
		price.Currency,
		price.URL,
		price.InStock,
		price.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save product price: %w", err)
	}

	return nil
}

// IndexProduct индексирует товар в Meilisearch
func (a *ProcessorAdapter) IndexProduct(product *products.Product) error {
	if a.meili == nil {
		// Meilisearch недоступен - не критично, просто пропускаем индексацию
		return nil
	}

	// Получаем названия магазинов для этого товара
	query := `
		SELECT DISTINCT s.name 
		FROM product_prices pp
		JOIN shops s ON pp.shop_id = s.id
		WHERE pp.product_id = $1
	`

	rows, err := a.pg.DB().Query(a.ctx, query, product.ID)
	if err != nil {
		// Не прерываем индексацию, просто будет без имен
	} else {
		defer rows.Close()
	}

	var shopNames []string
	if rows != nil {
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil && name != "" {
				shopNames = append(shopNames, name)
			}
		}
	}

	type MeiliDoc struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Brand       string            `json:"brand"`
		Category    string            `json:"category"`
		CategoryID  *string           `json:"category_id,omitempty"`
		Description string            `json:"description,omitempty"`
		ImageURL    string            `json:"image_url,omitempty"`
		Specs       map[string]string `json:"specs,omitempty"`
		ShopNames   []string          `json:"shop_names,omitempty"`
		ShopsCount  int               `json:"shops_count,omitempty"`
	}

	doc := MeiliDoc{
		ID:          product.ID,
		Name:        product.Name,
		Brand:       product.Brand,
		Category:    product.Category,
		CategoryID:  product.CategoryID,
		Description: product.Description,
		ImageURL:    product.ImageURL,
		Specs:       product.Specs,
		ShopNames:   shopNames,
		ShopsCount:  len(shopNames),
	}

	_, err = a.meili.Client().Index("products").AddDocuments([]MeiliDoc{doc})
	if err != nil {
		return fmt.Errorf("failed to index product in Meilisearch: %w", err)
	}

	return nil
}
