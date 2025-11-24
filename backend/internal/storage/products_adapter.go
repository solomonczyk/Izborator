package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/solomonczyk/izborator/internal/products"
)

// ProductsAdapter адаптер для работы с товарами
type ProductsAdapter struct {
	pg    *Postgres
	meili *Meilisearch
	ctx   context.Context
}

// NewProductsAdapter создаёт новый адаптер для товаров
func NewProductsAdapter(pg *Postgres, meili *Meilisearch) products.Storage {
	return &ProductsAdapter{
		pg:    pg,
		meili: meili,
		ctx:   pg.Context(),
	}
}

// GetProduct получает товар по ID
func (a *ProductsAdapter) GetProduct(id string) (*products.Product, error) {
	productUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		SELECT id, name, description, brand, category, image_url, specs, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product products.Product
	var specsJSON []byte
	var createdAt, updatedAt time.Time

	err = a.pg.DB().QueryRow(a.ctx, query, productUUID).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Brand,
		&product.Category,
		&product.ImageURL,
		&specsJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, products.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Десериализация JSONB specs
	if len(specsJSON) > 0 {
		if err := json.Unmarshal(specsJSON, &product.Specs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal specs: %w", err)
		}
	} else {
		product.Specs = make(map[string]string)
	}

	product.CreatedAt = createdAt
	product.UpdatedAt = updatedAt

	return &product, nil
}

// SearchProducts ищет товары по запросу
func (a *ProductsAdapter) SearchProducts(query string, limit, offset int) ([]*products.Product, int, error) {
	// Используем Meilisearch для полнотекстового поиска
	if a.meili != nil {
		return a.searchViaMeilisearch(query, limit, offset)
	}

	// Fallback на PostgreSQL, если Meilisearch недоступен
	return a.searchViaPostgres(query, limit, offset)
}

// searchViaMeilisearch поиск через Meilisearch
func (a *ProductsAdapter) searchViaMeilisearch(query string, limit, offset int) ([]*products.Product, int, error) {
	index := a.meili.Client().Index("products")

	searchRequest := &meilisearch.SearchRequest{
		Query:  query,
		Limit:  int64(limit),
		Offset: int64(offset),
	}

	searchResult, err := index.Search(query, searchRequest)
	if err != nil {
		// Если Meilisearch недоступен, fallback на PostgreSQL
		return a.searchViaPostgres(query, limit, offset)
	}

	// Преобразуем результаты Meilisearch в products.Product
	productsList := make([]*products.Product, 0, len(searchResult.Hits))

	for _, hit := range searchResult.Hits {
		// Meilisearch возвращает map[string]interface{}
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		product := &products.Product{
			Specs: make(map[string]string),
		}

		// Извлекаем поля из результата
		if id, ok := hitMap["id"].(string); ok {
			product.ID = id
		}
		if name, ok := hitMap["name"].(string); ok {
			product.Name = name
		}
		if desc, ok := hitMap["description"].(string); ok {
			product.Description = desc
		}
		if brand, ok := hitMap["brand"].(string); ok {
			product.Brand = brand
		}
		if category, ok := hitMap["category"].(string); ok {
			product.Category = category
		}
		if imageURL, ok := hitMap["image_url"].(string); ok {
			product.ImageURL = imageURL
		}

		// Обработка specs (может быть map или JSON string)
		if specs, ok := hitMap["specs"].(map[string]interface{}); ok {
			for k, v := range specs {
				if strVal, ok := v.(string); ok {
					product.Specs[k] = strVal
				}
			}
		}

		// Обработка timestamps
		if createdAt, ok := hitMap["created_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
				product.CreatedAt = t
			}
		}
		if updatedAt, ok := hitMap["updated_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
				product.UpdatedAt = t
			}
		}

		productsList = append(productsList, product)
	}

	total := int(searchResult.EstimatedTotalHits)

	return productsList, total, nil
}

// searchViaPostgres поиск через PostgreSQL (fallback)
func (a *ProductsAdapter) searchViaPostgres(query string, limit, offset int) ([]*products.Product, int, error) {
	searchQuery := fmt.Sprintf("%%%s%%", query)

	// Получаем общее количество
	countQuery := `
		SELECT COUNT(*) 
		FROM products 
		WHERE name ILIKE $1 OR description ILIKE $1 OR brand ILIKE $1
	`

	var total int
	err := a.pg.DB().QueryRow(a.ctx, countQuery, searchQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Получаем товары
	querySQL := `
		SELECT id, name, description, brand, category, image_url, specs, created_at, updated_at
		FROM products
		WHERE name ILIKE $1 OR description ILIKE $1 OR brand ILIKE $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := a.pg.DB().Query(a.ctx, querySQL, searchQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	var productsList []*products.Product

	for rows.Next() {
		var product products.Product
		var specsJSON []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Brand,
			&product.Category,
			&product.ImageURL,
			&specsJSON,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Десериализация JSONB specs
		if len(specsJSON) > 0 {
			if err := json.Unmarshal(specsJSON, &product.Specs); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal specs: %w", err)
			}
		} else {
			product.Specs = make(map[string]string)
		}

		product.CreatedAt = createdAt
		product.UpdatedAt = updatedAt

		productsList = append(productsList, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating products: %w", err)
	}

	return productsList, total, nil
}

// SaveProduct сохраняет товар
func (a *ProductsAdapter) SaveProduct(product *products.Product) error {
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

// GetProductPrices получает цены товара из разных магазинов
func (a *ProductsAdapter) GetProductPrices(productID string) ([]*products.ProductPrice, error) {
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		SELECT product_id, shop_id, shop_name, price, currency, url, in_stock, updated_at
		FROM product_prices
		WHERE product_id = $1
		ORDER BY price ASC, updated_at DESC
	`

	rows, err := a.pg.DB().Query(a.ctx, query, productUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product prices: %w", err)
	}
	defer rows.Close()

	var prices []*products.ProductPrice

	for rows.Next() {
		var price products.ProductPrice
		var updatedAt time.Time

		err := rows.Scan(
			&price.ProductID,
			&price.ShopID,
			&price.ShopName,
			&price.Price,
			&price.Currency,
			&price.URL,
			&price.InStock,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}

		price.UpdatedAt = updatedAt
		prices = append(prices, &price)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prices: %w", err)
	}

	return prices, nil
}

// SaveProductPrice сохраняет цену товара
func (a *ProductsAdapter) SaveProductPrice(price *products.ProductPrice) error {
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
