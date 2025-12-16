package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/products"
)

// ProductsAdapter адаптер для работы с товарами
type ProductsAdapter struct {
	pg     *Postgres
	meili  *Meilisearch
	ctx    context.Context
	logger *logger.Logger
}

// NewProductsAdapter создаёт новый адаптер для товаров
func NewProductsAdapter(pg *Postgres, meili *Meilisearch, log *logger.Logger) products.Storage {
	return &ProductsAdapter{
		pg:     pg,
		meili:  meili,
		ctx:    pg.Context(),
		logger: log,
	}
}

// GetProduct получает товар по ID
func (a *ProductsAdapter) GetProduct(id string) (*products.Product, error) {
	productUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		SELECT id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product products.Product
	var specsJSON []byte
	var createdAt, updatedAt time.Time
	var categoryID *uuid.UUID

	err = a.pg.DB().QueryRow(a.ctx, query, productUUID).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Brand,
		&product.Category,
		&categoryID,
		&product.ImageURL,
		&specsJSON,
		&createdAt,
		&updatedAt,
	)

	if categoryID != nil {
		categoryIDStr := categoryID.String()
		product.CategoryID = &categoryIDStr
	}

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
		SELECT id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at
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
		var categoryID *string

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Brand,
			&product.Category,
			&categoryID,
			&product.ImageURL,
			&specsJSON,
			&createdAt,
			&updatedAt,
		)

		product.CategoryID = categoryID

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

	var categoryID *uuid.UUID
	if product.CategoryID != nil {
		catID, err := uuid.Parse(*product.CategoryID)
		if err == nil {
			categoryID = &catID
		}
	}

	query := `
		INSERT INTO products (id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			brand = EXCLUDED.brand,
			category = EXCLUDED.category,
			category_id = EXCLUDED.category_id,
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
		categoryID,
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
		
		// Логируем для отладки shop_name
		if a.logger != nil && price.ShopName == "" {
			a.logger.Warn("GetProductPrices: empty ShopName", map[string]interface{}{
				"product_id": price.ProductID,
				"shop_id":    price.ShopID,
			})
		}
		
		prices = append(prices, &price)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prices: %w", err)
	}

	return prices, nil
}

// GetProductPricesByCity получает цены товара для конкретного города
func (a *ProductsAdapter) GetProductPricesByCity(productID string, cityID string) ([]*products.ProductPrice, error) {
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	cityUUID, err := uuid.Parse(cityID)
	if err != nil {
		return nil, fmt.Errorf("invalid city ID: %w", err)
	}

	query := `
		SELECT product_id, shop_id, shop_name, price, currency, url, in_stock, updated_at
		FROM product_prices
		WHERE product_id = $1 AND (city_id = $2 OR city_id IS NULL)
		ORDER BY price ASC, updated_at DESC
	`

	rows, err := a.pg.DB().Query(a.ctx, query, productUUID, cityUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product prices by city: %w", err)
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
	// Используем Meilisearch для поиска с фильтрами (основной метод)
	if a.meili != nil {
		return a.browseViaMeilisearch(ctx, params)
	}

	// Fallback на PostgreSQL, если Meilisearch недоступен
	if a.logger != nil {
		a.logger.Info("Browse: Using PostgreSQL fallback (Meilisearch unavailable)", map[string]interface{}{
			"query": params.Query,
			"meili_available": false,
		})
	}
	return a.browseViaPostgres(ctx, params)
}

// browseViaMeilisearch каталог через Meilisearch с фильтрами
func (a *ProductsAdapter) browseViaMeilisearch(ctx context.Context, params products.BrowseParams) (*products.BrowseResult, error) {
	// Логирование для отладки
	if a.logger != nil {
		a.logger.Info("Browse: Using Meilisearch", map[string]interface{}{
			"query": params.Query,
		})
	}
	
	index := a.meili.Client().Index("products")

	searchReq := &meilisearch.SearchRequest{
		Query:  params.Query,
		Limit:  int64(params.PerPage),
		Offset: int64((params.Page - 1) * params.PerPage),
	}

	// Фильтры Meilisearch
	var filters []string
	if len(params.CategoryIDs) > 0 {
		// Фильтр по списку category_id (родитель + дочерние категории)
		categoryFilter := make([]string, len(params.CategoryIDs))
		for i, catID := range params.CategoryIDs {
			categoryFilter[i] = fmt.Sprintf("category_id = \"%s\"", catID)
		}
		filters = append(filters, "("+strings.Join(categoryFilter, " OR ")+")")
	} else if params.CategoryID != nil {
		// Фильтр по category_id (приоритет над category slug)
		filters = append(filters, fmt.Sprintf("category_id = \"%s\"", *params.CategoryID))
	} else if params.Category != "" {
		// Фильтр по категории (строка, для обратной совместимости)
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
		// Если Meilisearch недоступен или API ключ неверный, fallback на PostgreSQL
		return a.browseViaPostgres(ctx, params)
	}

	// Проверяем, что есть результаты (может быть пустой результат из-за неверного API ключа)
	// Если результатов нет - делаем fallback на PostgreSQL
	if searchResult.Hits == nil || len(searchResult.Hits) == 0 {
		// Fallback на PostgreSQL для получения товаров из БД
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
		// (с фильтром по городу, если указан)
		var prices []*products.ProductPrice
		var err error
		if params.CityID != nil {
			prices, err = a.GetProductPricesByCity(productID, *params.CityID)
		} else {
			prices, err = a.GetProductPrices(productID)
		}
		if err != nil {
			// Пропускаем товар, если не удалось получить цены
			continue
		}

		// Если фильтр по городу указан и нет цен в этом городе - пропускаем товар
		if params.CityID != nil && len(prices) == 0 {
			continue
		}

		// Собираем уникальные названия магазинов
		shopNamesMap := make(map[string]bool)
		for _, price := range prices {
			if price.ShopName != "" {
				shopNamesMap[price.ShopName] = true
			}
		}
		// Инициализируем как пустой слайс, а не nil, чтобы он всегда сериализовался в JSON
		shopNames := make([]string, 0, len(shopNamesMap))
		for shopName := range shopNamesMap {
			shopNames = append(shopNames, shopName)
		}
		// Если слайс все еще nil (не должно быть, но на всякий случай)
		if shopNames == nil {
			shopNames = []string{}
		}

		browseProduct := products.BrowseProduct{
			ID:         productID,
			ShopsCount: len(prices),
			ShopNames:  shopNames,
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

	// Убеждаемся, что ShopNames не nil для всех элементов
	for i := range items {
		if items[i].ShopNames == nil {
			items[i].ShopNames = []string{}
		}
	}

	// Финальная проверка shop_names перед возвратом
	if a.logger != nil && len(items) > 0 {
		for i, item := range items {
			if i < 3 { // Проверяем первые 3
				// Сериализуем в JSON для проверки
				jsonBytes, _ := json.Marshal(item)
				a.logger.Info("browseViaPostgres: final item check", map[string]interface{}{
					"product_id":   item.ID,
					"product_name": item.Name,
					"shop_names":   item.ShopNames,
					"shop_names_len": len(item.ShopNames),
					"shops_count":  item.ShopsCount,
					"json_preview": string(jsonBytes),
				})
			}
		}
	}

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
	// Явное логирование для отладки
	if a.logger != nil {
		a.logger.Info("browseViaPostgres: called", map[string]interface{}{
			"query": params.Query,
			"page":  params.Page,
		})
	}
	
	// Простая реализация через PostgreSQL
	// Можно расширить позже с фильтрами
	var productsList []*products.Product
	var err error

	if params.Query == "" {
		// Если запрос пустой, получаем все товары напрямую
		querySQL := `
			SELECT id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at
			FROM products
			ORDER BY name
			LIMIT $1
		`
		rows, err := a.pg.DB().Query(a.ctx, querySQL, 1000)
		if err != nil {
			return nil, fmt.Errorf("failed to get products: %w", err)
		}
		defer rows.Close()

		productsList = make([]*products.Product, 0)
		for rows.Next() {
			var product products.Product
			var specsJSON []byte
			var createdAt, updatedAt time.Time
			var categoryID *string

			err := rows.Scan(
				&product.ID,
				&product.Name,
				&product.Description,
				&product.Brand,
				&product.Category,
				&categoryID,
				&product.ImageURL,
				&specsJSON,
				&createdAt,
				&updatedAt,
			)

			product.CategoryID = categoryID

			if err != nil {
				return nil, fmt.Errorf("failed to scan product: %w", err)
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

			productsList = append(productsList, &product)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating products: %w", err)
		}
	} else {
		// Если есть запрос, используем searchViaPostgres
		query := "%" + params.Query + "%"
		productsList, _, err = a.searchViaPostgres(query, 1000, 0)
		if err != nil {
			return nil, err
		}
	}

	// Преобразуем в BrowseProduct
	items := make([]products.BrowseProduct, 0, len(productsList))
	for _, p := range productsList {
		// Фильтр по списку category_id (родитель + дочерние категории)
		if len(params.CategoryIDs) > 0 {
			found := false
			if p.CategoryID != nil {
				for _, catID := range params.CategoryIDs {
					if *p.CategoryID == catID {
						found = true
						break
					}
				}
			}
			if !found {
				continue
			}
		} else if params.CategoryID != nil {
			// Фильтр по category_id (приоритет)
			if p.CategoryID == nil || *p.CategoryID != *params.CategoryID {
				continue
			}
		} else if params.Category != "" {
			// Фильтр по категории (строка, для обратной совместимости)
			if p.Category == "" || p.Category != params.Category {
				continue
			}
		}

		// Получаем цены (с фильтром по городу, если указан)
		var prices []*products.ProductPrice
		var err error
		if params.CityID != nil {
			prices, err = a.GetProductPricesByCity(p.ID, *params.CityID)
		} else {
			prices, err = a.GetProductPrices(p.ID)
		}
		if err != nil {
			// Если ошибка получения цен - пропускаем товар (но это не должно происходить для товаров с ценами)
			continue
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: логируем prices сразу после получения
		if a.logger != nil && len(prices) > 0 {
			priceDebug := make([]map[string]interface{}, 0, len(prices))
			for _, price := range prices {
				priceDebug = append(priceDebug, map[string]interface{}{
					"shop_id":   price.ShopID,
					"shop_name": price.ShopName,
					"price":     price.Price,
				})
			}
			a.logger.Info("browseViaPostgres: prices from GetProductPrices", map[string]interface{}{
				"product_id":   p.ID,
				"product_name": p.Name,
				"prices_count": len(prices),
				"price_details": priceDebug,
			})
		}


		// Если фильтр по городу указан и нет цен в этом городе - пропускаем товар
		if params.CityID != nil && len(prices) == 0 {
			continue
		}

		// Собираем уникальные названия магазинов
		shopNamesMap := make(map[string]bool)
		priceDetails := make([]map[string]interface{}, 0, len(prices))
		for _, price := range prices {
			// Логируем каждую цену для отладки
			if a.logger != nil {
				priceDetails = append(priceDetails, map[string]interface{}{
					"shop_id":   price.ShopID,
					"shop_name": price.ShopName,
					"price":     price.Price,
				})
			}
			if price.ShopName != "" {
				shopNamesMap[price.ShopName] = true
			}
		}
		// Инициализируем как пустой слайс, а не nil, чтобы он всегда сериализовался в JSON
		shopNames := make([]string, 0, len(shopNamesMap))
		for shopName := range shopNamesMap {
			shopNames = append(shopNames, shopName)
		}
		// Если слайс все еще nil (не должно быть, но на всякий случай)
		if shopNames == nil {
			shopNames = []string{}
		}

		// Отладочное логирование
		if a.logger != nil {
			a.logger.Info("browseViaPostgres: shop_names processing", map[string]interface{}{
				"product_id":   p.ID,
				"product_name": p.Name,
				"prices_count": len(prices),
				"shop_names":   shopNames,
				"shop_names_len": len(shopNames),
				"price_details": priceDetails,
			})
		}

		// Убеждаемся, что shopNames не nil
		if shopNames == nil {
			shopNames = []string{}
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: логируем shopNames перед созданием BrowseProduct
		if a.logger != nil {
			a.logger.Info("browseViaPostgres: BEFORE creating BrowseProduct", map[string]interface{}{
				"product_id":   p.ID,
				"product_name": p.Name,
				"shopNames":    shopNames,
				"shopNames_len": len(shopNames),
				"shopNames_nil": shopNames == nil,
				"prices_count": len(prices),
			})
		}

		browseProduct := products.BrowseProduct{
			ID:         p.ID,
			Name:       p.Name,
			Brand:      p.Brand,
			Category:   p.Category,
			CategoryID: p.CategoryID,
			ImageURL:   p.ImageURL,
			ShopsCount: len(prices),
			ShopNames:  shopNames, // Должен всегда быть массивом, даже пустым
			Specs:      p.Specs,
		}

		// Дополнительная проверка после создания
		if browseProduct.ShopNames == nil {
			browseProduct.ShopNames = []string{}
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: логируем shopNames после создания BrowseProduct
		if a.logger != nil {
			jsonBytes, _ := json.Marshal(browseProduct)
			a.logger.Info("browseViaPostgres: AFTER creating BrowseProduct", map[string]interface{}{
				"product_id":   browseProduct.ID,
				"product_name": browseProduct.Name,
				"shopNames":    browseProduct.ShopNames,
				"shopNames_len": len(browseProduct.ShopNames),
				"shopNames_nil": browseProduct.ShopNames == nil,
				"json_preview": string(jsonBytes),
			})
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

		// Финальная проверка перед добавлением в результат
		if a.logger != nil {
			if len(browseProduct.ShopNames) > 0 {
				a.logger.Info("browseViaPostgres: adding product with shop_names", map[string]interface{}{
					"product_id":   browseProduct.ID,
					"product_name": browseProduct.Name,
					"shop_names":   browseProduct.ShopNames,
					"shops_count":  browseProduct.ShopsCount,
					"shop_names_len": len(browseProduct.ShopNames),
				})
			} else {
				a.logger.Warn("browseViaPostgres: adding product WITHOUT shop_names", map[string]interface{}{
					"product_id":   browseProduct.ID,
					"product_name": browseProduct.Name,
					"shops_count":  browseProduct.ShopsCount,
					"prices_count": len(prices),
					"shopNames_nil": browseProduct.ShopNames == nil,
				})
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

// GetURLsForRescrape возвращает список URL и ID магазинов для товаров,
// цена которых не обновлялась дольше указанного времени
func (a *ProductsAdapter) GetURLsForRescrape(ctx context.Context, olderThan time.Duration, limit int) ([]products.RescrapeItem, error) {
	// Выбираем цены, обновленные давно, но у активных магазинов
	// Используем DISTINCT ON для получения уникальных пар (url, shop_id) с самой старой датой
	query := `
		SELECT url, shop_id
		FROM (
			SELECT DISTINCT ON (pp.url, pp.shop_id) 
				pp.url, 
				pp.shop_id, 
				pp.updated_at
			FROM product_prices pp
			JOIN shops s ON pp.shop_id = s.id
			WHERE s.is_active = true
			  AND pp.url IS NOT NULL
			  AND pp.url != ''
			  AND pp.updated_at < $1
			ORDER BY pp.url, pp.shop_id, pp.updated_at ASC
		) AS distinct_prices
		ORDER BY updated_at ASC
		LIMIT $2
	`

	threshold := time.Now().Add(-olderThan)

	rows, err := a.pg.DB().Query(ctx, query, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs for rescrape: %w", err)
	}
	defer rows.Close()

	var results []products.RescrapeItem

	for rows.Next() {
		var item products.RescrapeItem
		if err := rows.Scan(&item.URL, &item.ShopID); err != nil {
			a.logger.Warn("Failed to scan rescrape item", map[string]interface{}{"error": err.Error()})
			continue
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rescrape items: %w", err)
	}

	return results, nil
}
