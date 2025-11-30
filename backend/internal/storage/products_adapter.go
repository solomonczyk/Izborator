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

// Browse возвращает каталог товаров с фильтрами
func (a *ProductsAdapter) Browse(ctx context.Context, params products.BrowseParams) (*products.BrowseResult, error) {
	// Используем Meilisearch для поиска с фильтрами
	if a.meili != nil {
		return a.browseViaMeilisearch(ctx, params)
	}

	// Fallback на PostgreSQL, если Meilisearch недоступен
	return a.browseViaPostgres(ctx, params)
}

// browseViaMeilisearch каталог через Meilisearch с фильтрами
func (a *ProductsAdapter) browseViaMeilisearch(ctx context.Context, params products.BrowseParams) (*products.BrowseResult, error) {
	index := a.meili.Client().Index("products")

	searchReq := &meilisearch.SearchRequest{
		Query:  params.Query,
		Limit:  int64(params.PerPage),
		Offset: int64((params.Page - 1) * params.PerPage),
	}

	// Фильтры Meilisearch
	var filters []string
	if params.Category != "" {
		// Фильтр по категории (если указана)
		// Показываем только товары с указанной категорией (не NULL)
		filters = append(filters, fmt.Sprintf("category = \"%s\"", params.Category))
	}
	// Пока shop_id и price фильтры пропускаем, т.к. эти поля могут быть не в индексе
	// Их можно добавить позже, когда обновим индекс

	if len(filters) > 0 {
		searchReq.Filter = filters
	}

	// Сортировка
	switch params.Sort {
	case "price_asc":
		// Пока сортировка по цене не работает без min_price в индексе
		// Можно добавить позже
	case "price_desc":
		// Пока сортировка по цене не работает без min_price в индексе
	case "newest":
		searchReq.Sort = []string{"created_at:desc"}
	case "name_asc":
		searchReq.Sort = []string{"name:asc"}
	default:
		// relevance по умолчанию
	}

	searchResult, err := index.Search(params.Query, searchReq)
	if err != nil {
		// Если Meilisearch недоступен, fallback на PostgreSQL
		return a.browseViaPostgres(ctx, params)
	}

	// Преобразуем результаты и обогащаем данными о ценах из PostgreSQL
	items := make([]products.BrowseProduct, 0, len(searchResult.Hits))

	for _, hit := range searchResult.Hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		productID, ok := hitMap["id"].(string)
		if !ok || productID == "" {
			continue
		}

		// Получаем цены из PostgreSQL для вычисления min_price, max_price, shops_count
		prices, err := a.GetProductPrices(productID)
		if err != nil {
			// Пропускаем товар, если не удалось получить цены
			continue
		}

		browseProduct := products.BrowseProduct{
			ID:         productID,
			ShopsCount: len(prices),
			Specs:      make(map[string]string),
		}

		// Извлекаем базовые поля из Meilisearch
		if name, ok := hitMap["name"].(string); ok {
			browseProduct.Name = name
		}
		if brand, ok := hitMap["brand"].(string); ok {
			browseProduct.Brand = brand
		}
		if category, ok := hitMap["category"].(string); ok {
			browseProduct.Category = category
		}
		if imageURL, ok := hitMap["image_url"].(string); ok {
			browseProduct.ImageURL = imageURL
		}

		// Обработка specs
		if specs, ok := hitMap["specs"].(map[string]interface{}); ok {
			for k, v := range specs {
				if strVal, ok := v.(string); ok {
					browseProduct.Specs[k] = strVal
				}
			}
		}

		// Вычисляем min_price, max_price, currency из цен
		if len(prices) > 0 {
			minPrice := prices[0].Price
			maxPrice := prices[0].Price
			currency := prices[0].Currency

			for _, price := range prices {
				if price.Price < minPrice {
					minPrice = price.Price
				}
				if price.Price > maxPrice {
					maxPrice = price.Price
				}
			}

			browseProduct.MinPrice = minPrice
			browseProduct.MaxPrice = maxPrice
			browseProduct.Currency = currency

			// Применяем фильтры по цене и shop_id (если указаны)
			// Товар должен попадать в диапазон: min_price >= MinPrice и max_price <= MaxPrice
			if params.MinPrice != nil && browseProduct.MaxPrice < *params.MinPrice {
				continue // Максимальная цена товара меньше минимальной в фильтре - исключаем
			}
			if params.MaxPrice != nil && browseProduct.MinPrice > *params.MaxPrice {
				continue // Минимальная цена товара больше максимальной в фильтре - исключаем
			}
			if params.ShopID != "" {
				found := false
				for _, price := range prices {
					if price.ShopID == params.ShopID {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}

		items = append(items, browseProduct)
	}

	// Применяем сортировку ПОСЛЕ получения всех данных (включая цены)
	switch params.Sort {
	case "price_asc":
		// Сортируем по минимальной цене (возрастание)
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].MinPrice > items[j].MinPrice {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
	case "price_desc":
		// Сортируем по минимальной цене (убывание)
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].MinPrice < items[j].MinPrice {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
	case "name_asc":
		// Сортируем по названию (A-Z)
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].Name > items[j].Name {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
	case "name_desc":
		// Сортируем по названию (Z-A)
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].Name < items[j].Name {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
	}

	// Пересчитываем total после фильтрации
	total := int64(len(items))
	totalPages := int((total + int64(params.PerPage) - 1) / int64(params.PerPage))

	return &products.BrowseResult{
		Items:      items,
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// browseViaPostgres каталог через PostgreSQL (fallback)
func (a *ProductsAdapter) browseViaPostgres(ctx context.Context, params products.BrowseParams) (*products.BrowseResult, error) {
	// Простая реализация через PostgreSQL
	// Можно расширить позже с фильтрами
	query := params.Query
	if query == "" {
		query = "%"
	} else {
		query = "%" + query + "%"
	}

	// Получаем товары (без пагинации, т.к. будем фильтровать и сортировать)
	productsList, _, err := a.searchViaPostgres(query, 1000, 0) // Получаем больше товаров для фильтрации
	if err != nil {
		return nil, err
	}

	// Преобразуем в BrowseProduct
	items := make([]products.BrowseProduct, 0, len(productsList))
	for _, p := range productsList {
		// Фильтр по категории
		// Если категория указана в фильтре, показываем только товары с этой категорией (не NULL)
		if params.Category != "" {
			if p.Category == "" || p.Category != params.Category {
				continue
			}
		}

		// Получаем цены
		prices, err := a.GetProductPrices(p.ID)
		if err != nil {
			continue
		}

		browseProduct := products.BrowseProduct{
			ID:         p.ID,
			Name:       p.Name,
			Brand:      p.Brand,
			Category:   p.Category,
			ImageURL:   p.ImageURL,
			ShopsCount: len(prices),
			Specs:      p.Specs,
		}

		if len(prices) > 0 {
			minPrice := prices[0].Price
			maxPrice := prices[0].Price
			currency := prices[0].Currency

			for _, price := range prices {
				if price.Price < minPrice {
					minPrice = price.Price
				}
				if price.Price > maxPrice {
					maxPrice = price.Price
				}
			}

			browseProduct.MinPrice = minPrice
			browseProduct.MaxPrice = maxPrice
			browseProduct.Currency = currency

			// Применяем фильтры по цене и shop_id
			if params.MinPrice != nil && browseProduct.MaxPrice < *params.MinPrice {
				continue // Максимальная цена товара меньше минимальной в фильтре - исключаем
			}
			if params.MaxPrice != nil && browseProduct.MinPrice > *params.MaxPrice {
				continue // Минимальная цена товара больше максимальной в фильтре - исключаем
			}
			if params.ShopID != "" {
				found := false
				for _, price := range prices {
					if price.ShopID == params.ShopID {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}

		items = append(items, browseProduct)
	}

	// Применяем сортировку
	switch params.Sort {
	case "price_asc":
		// Сортируем по минимальной цене (возрастание)
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].MinPrice > items[j].MinPrice {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
	case "price_desc":
		// Сортируем по минимальной цене (убывание)
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].MinPrice < items[j].MinPrice {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
	case "name_asc":
		// Сортируем по названию (A-Z)
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].Name > items[j].Name {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
	case "name_desc":
		// Сортируем по названию (Z-A)
		for i := 0; i < len(items)-1; i++ {
			for j := i + 1; j < len(items); j++ {
				if items[i].Name < items[j].Name {
					items[i], items[j] = items[j], items[i]
				}
			}
		}
	}

	// Сохраняем total ДО пагинации
	totalCount := len(items)

	// Применяем пагинацию ПОСЛЕ фильтрации и сортировки
	start := (params.Page - 1) * params.PerPage
	end := start + params.PerPage
	if start >= len(items) {
		items = []products.BrowseProduct{}
	} else if end > len(items) {
		items = items[start:]
	} else {
		items = items[start:end]
	}

	totalPages := (totalCount + params.PerPage - 1) / params.PerPage

	return &products.BrowseResult{
		Items:      items,
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      int64(totalCount),
		TotalPages: totalPages,
	}, nil
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
