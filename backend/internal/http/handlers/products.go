package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/solomonczyk/izborator/internal/categories"
	"github.com/solomonczyk/izborator/internal/cities"
	"github.com/solomonczyk/izborator/internal/domainpack"
	appErrors "github.com/solomonczyk/izborator/internal/errors"
	"github.com/solomonczyk/izborator/internal/http/validation"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/pricehistory"
	"github.com/solomonczyk/izborator/internal/products"
)

type SearchRequest struct {
	Query string `json:"query" validate:"required,min=2,max=200"`
}

type BrowseRequest struct {
	Query    string   `json:"query" validate:"omitempty,max=200"`
	Category string   `json:"category" validate:"omitempty"`
	Brand    string   `json:"brand" validate:"omitempty"`
	City     string   `json:"city" validate:"omitempty"`
	ShopID   string   `json:"shop_id" validate:"omitempty,uuid4"`
	Type     string   `json:"type" validate:"omitempty,oneof=good service"` // Фильтр по типу: "good" | "service" | ""
	MinPrice *float64 `json:"min_price" validate:"omitempty,gte=0"`
	MaxPrice *float64 `json:"max_price" validate:"omitempty,gte=0"`
	Page     int      `json:"page" validate:"gte=1"`
	PerPage  int      `json:"per_page" validate:"gte=1,lte=100"`
	Sort     string   `json:"sort" validate:"omitempty,oneof=price_asc price_desc name_asc name_desc newest"`
}

type FacetSchemaResponse struct {
	Domain string                      `json:"domain"`
	Facets []domainpack.FacetDefinition `json:"facets"`
}

// ProductsHandler обработчик для работы с товарами
type ProductsHandler struct {
	*BaseHandler
	service         *products.Service
	priceHistorySvc *pricehistory.Service
	categoriesSvc   *categories.Service
	citiesSvc       *cities.Service
}

// NewProductsHandler создаёт новый обработчик товаров
func NewProductsHandler(service *products.Service, priceHistorySvc *pricehistory.Service, categoriesSvc *categories.Service, citiesSvc *cities.Service, log *logger.Logger, translator *i18n.Translator) *ProductsHandler {
	return &ProductsHandler{
		BaseHandler:     NewBaseHandler(log, translator),
		service:         service,
		priceHistorySvc: priceHistorySvc,
		categoriesSvc:   categoriesSvc,
		citiesSvc:       citiesSvc,
	}
}

// Search обрабатывает поиск товаров
// GET /api/v1/products/search?q=query
// GET /api/products?q=query&limit=10&offset=0 (старый формат)
func (h *ProductsHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := validation.SanitizeString(r.URL.Query().Get("q"))
	req := SearchRequest{Query: query}

	if err := validation.ValidateStruct(req); err != nil {
		message := validation.FormatValidationErrors(err)
		appErr := appErrors.NewValidationError(message, err)
		h.RespondAppError(w, r, appErr)
		return
	}
	ctx := r.Context()

	// Для нового endpoint /api/v1/products/search используем простой поиск
	if r.URL.Path == "/api/v1/products/search" {
		results, err := h.service.Search(ctx, query)
		if err != nil {
			appErr := appErrors.NewInternalError("Search failed", err)
			h.RespondAppError(w, r, appErr)
			return
		}

		h.RespondJSON(w, http.StatusOK, results)
		return
	}

	// Старый формат с limit/offset
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	result, err := h.service.SearchWithPagination(ctx, query, limit, offset)
	if err != nil {
		appErr := appErrors.NewInternalError("Search failed", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	h.RespondJSON(w, http.StatusOK, result)
}

// ProductResponse структура ответа для GetByID (товар + цены)
type ProductResponse struct {
	ID       string                   `json:"id"`
	Name     string                   `json:"name"`
	Brand    string                   `json:"brand,omitempty"`
	Category string                   `json:"category,omitempty"`
	ImageURL string                   `json:"image_url,omitempty"`
	Specs    map[string]string        `json:"specs,omitempty"`
	Prices   []*products.ProductPrice `json:"prices"`
}

// GetByID обрабатывает получение товара по ID с ценами
// GET /api/v1/products/:id
func (h *ProductsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		appErr := appErrors.NewValidationError("Product ID is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	// Валидация UUID
	if err := validation.ValidateUUID(id); err != nil {
		appErr := appErrors.NewValidationError("Invalid product ID format", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	ctx := r.Context()
	_ = ctx // для будущего использования

	// 1. Получаем товар
	product, err := h.service.GetByID(id)
	if err != nil {
		var appErr *appErrors.AppError
		if err == products.ErrProductNotFound || err == products.ErrInvalidProductID {
			appErr = appErrors.NewNotFound("Product not found")
		} else {
			appErr = appErrors.NewInternalError("Failed to load product", err)
		}
		h.RespondAppError(w, r, appErr)
		return
	}

	// 2. Получаем цены
	prices, err := h.service.GetPrices(id)
	if err != nil {
		h.logger.Error("GetProductPrices failed", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		// Не критично - возвращаем товар без цен, но логируем ошибку
		prices = []*products.ProductPrice{}
	}

	// 3. Формируем ответ
	resp := ProductResponse{
		ID:       product.ID,
		Name:     product.Name,
		Brand:    product.Brand,
		Category: product.Category,
		ImageURL: product.ImageURL,
		Specs:    product.Specs,
		Prices:   prices,
	}

	h.RespondJSON(w, http.StatusOK, resp)
}

// GetPrices обрабатывает получение цен товара из разных магазинов
// GET /api/products/:id/prices
func (h *ProductsHandler) GetPrices(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		appErr := appErrors.NewValidationError("Product ID is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	// Валидация UUID
	if err := validation.ValidateUUID(id); err != nil {
		appErr := appErrors.NewValidationError("Invalid product ID format", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	prices, err := h.service.GetPrices(id)
	if err != nil {
		var appErr *appErrors.AppError
		if err == products.ErrInvalidProductID {
			appErr = appErrors.NewValidationError("Invalid product ID", err)
		} else {
			appErr = appErrors.NewInternalError("Failed to get prices", err)
		}
		h.RespondAppError(w, r, appErr)
		return
	}

	result := map[string]interface{}{
		"product_id": id,
		"prices":     prices,
	}

	h.RespondJSON(w, http.StatusOK, result)
}

// GetPriceHistory обрабатывает получение истории цен товара
// GET /api/v1/products/{id}/price-history?period=month&shops=shop1,shop2
func (h *ProductsHandler) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		appErr := appErrors.NewValidationError("Product ID is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	// Валидация UUID
	if err := validation.ValidateUUID(id); err != nil {
		appErr := appErrors.NewValidationError("Invalid product ID format", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	// Параметры запроса
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month" // По умолчанию месяц
	}

	// Парсим список магазинов (опционально)
	shopsParam := r.URL.Query().Get("shops")
	var shopIDs []string
	if shopsParam != "" {
		// Простой парсинг через запятую
		shops := strings.Split(shopsParam, ",")
		for _, shop := range shops {
			shop = strings.TrimSpace(shop)
			if shop != "" {
				shopIDs = append(shopIDs, shop)
			}
		}
	}

	// Получаем данные для графика
	chart, err := h.priceHistorySvc.GetPriceChart(id, period, shopIDs)
	if err != nil {
		appErr := appErrors.NewInternalError("Failed to get price history", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	// Рассчитываем статистику
	stats := calculatePriceStats(chart)

	result := map[string]interface{}{
		"product_id": id,
		"period":     period,
		"from":       chart.From.Format(time.RFC3339),
		"to":         chart.To.Format(time.RFC3339),
		"shops":      chart.Shops,
		"shop_names": chart.ShopNames,
		"stats":      stats,
	}

	h.RespondJSON(w, http.StatusOK, result)
}

// PriceStats статистика цен
type PriceStats struct {
	MinPrice    float64   `json:"min_price"`
	MaxPrice    float64   `json:"max_price"`
	AvgPrice    float64   `json:"avg_price"`
	PriceChange float64   `json:"price_change"` // Изменение за период (%)
	FirstPrice  float64   `json:"first_price"`
	LastPrice   float64   `json:"last_price"`
	FirstDate   time.Time `json:"first_date"`
	LastDate    time.Time `json:"last_date"`
}

// calculatePriceStats рассчитывает статистику цен из графика
func calculatePriceStats(chart *pricehistory.PriceChart) PriceStats {
	if chart == nil || len(chart.Shops) == 0 {
		return PriceStats{}
	}

	var allPrices []float64
	var firstPrice, lastPrice float64
	var firstDate, lastDate time.Time

	// Собираем все цены из всех магазинов
	for _, points := range chart.Shops {
		for _, point := range points {
			allPrices = append(allPrices, point.Price)

			if firstDate.IsZero() || point.Timestamp.Before(firstDate) {
				firstDate = point.Timestamp
				firstPrice = point.Price
			}
			if lastDate.IsZero() || point.Timestamp.After(lastDate) {
				lastDate = point.Timestamp
				lastPrice = point.Price
			}
		}
	}

	if len(allPrices) == 0 {
		return PriceStats{}
	}

	// Находим мин/макс
	minPrice := allPrices[0]
	maxPrice := allPrices[0]
	sum := 0.0

	for _, price := range allPrices {
		if price < minPrice {
			minPrice = price
		}
		if price > maxPrice {
			maxPrice = price
		}
		sum += price
	}

	avgPrice := sum / float64(len(allPrices))

	// Рассчитываем изменение цены (%)
	var priceChange float64
	if firstPrice > 0 {
		priceChange = ((lastPrice - firstPrice) / firstPrice) * 100
	}

	return PriceStats{
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		AvgPrice:    avgPrice,
		PriceChange: priceChange,
		FirstPrice:  firstPrice,
		LastPrice:   lastPrice,
		FirstDate:   firstDate,
		LastDate:    lastDate,
	}
}

// Browse обрабатывает каталог товаров с фильтрами
// GET /api/v1/products/browse?query=motorola&category=phones&min_price=10000&max_price=30000&shop_id=...&page=1&per_page=20&sort=price_asc
// Facets returns facet schema for a domain.
// GET /api/v1/products/facets?type=<domain>
func (h *ProductsHandler) Facets(w http.ResponseWriter, r *http.Request) {
	domain := validation.SanitizeString(r.URL.Query().Get("type"))
	if domain == "" {
		appErr := appErrors.NewValidationError("type is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}
	if !domainpack.HasDomain(domain) {
		allowed := strings.Join(domainpack.Domains(), ", ")
		appErr := appErrors.NewValidationError("type must be one of: "+allowed, nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	facets, err := domainpack.Facets(domain)
	if err != nil {
		appErr := appErrors.NewInternalError("failed to load facet schema", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	if domain == "goods" {
		brands, err := h.service.ListBrands(r.Context(), string(products.ProductTypeGood))
		if err != nil {
			appErr := appErrors.NewInternalError("failed to load brand facets", err)
			h.RespondAppError(w, r, appErr)
			return
		}
		for i := range facets {
			if facets[i].SemanticType == "brand" {
				facets[i].Values = brands
				break
			}
		}
	}

	resp := FacetSchemaResponse{
		Domain: domain,
		Facets: facets,
	}
	h.RespondJSON(w, http.StatusOK, resp)
}

func (h *ProductsHandler) Browse(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	query := validation.SanitizeString(q.Get("query"))
	category := validation.SanitizeString(q.Get("category"))
	brand := validation.SanitizeString(q.Get("brand"))
	city := validation.SanitizeString(q.Get("city"))
	shopID := validation.SanitizeString(q.Get("shop_id"))
	productType := validation.SanitizeString(q.Get("type")) // "good" | "service" | ""
	sort := validation.SanitizeString(q.Get("sort"))

	// Валидация типа продукта
	if productType != "" && productType != "good" && productType != "service" {
		appErr := appErrors.NewValidationError("type must be 'good' or 'service'", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	minPriceStr := q.Get("min_price")
	maxPriceStr := q.Get("max_price")
	minDurationStr := q.Get("min_duration")
	maxDurationStr := q.Get("max_duration")

	page, err := validation.ParseIntParam(q, "page", 1)
	if err != nil {
		appErr := appErrors.NewValidationError(err.Error(), err)
		h.RespondAppError(w, r, appErr)
		return
	}
	perPage, err := validation.ParseIntParam(q, "per_page", 20)
	if err != nil {
		appErr := appErrors.NewValidationError(err.Error(), err)
		h.RespondAppError(w, r, appErr)
		return
	}

	// ????????? ????????? (???????) ????? ???????? ??????????
	if err := validation.ValidatePagination(page, perPage); err != nil {
		appErr := appErrors.NewValidationError(err.Error(), err)
		h.RespondAppError(w, r, appErr)
		return
	}

	var (
		minPrice *float64
		maxPrice *float64
		minDuration *int
		maxDuration *int
	)

	// ?????? ????
	if minPriceStr != "" {
		if v, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			if err := validation.ValidatePrice(v); err != nil {
				appErr := appErrors.NewValidationError(err.Error(), err)
				h.RespondAppError(w, r, appErr)
				return
			}
			minPrice = &v
		} else {
			appErr := appErrors.NewValidationError("min_price must be a number", err)
			h.RespondAppError(w, r, appErr)
			return
		}
	}
	if maxPriceStr != "" {
		if v, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			if err := validation.ValidatePrice(v); err != nil {
				appErr := appErrors.NewValidationError(err.Error(), err)
				h.RespondAppError(w, r, appErr)
				return
			}
			maxPrice = &v
		} else {
			appErr := appErrors.NewValidationError("max_price must be a number", err)
			h.RespondAppError(w, r, appErr)
			return
		}
	}

	if minPrice != nil && maxPrice != nil && *minPrice > *maxPrice {
		appErr := appErrors.NewValidationError("min_price cannot be greater than max_price", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	if productType == "service" {
		if minDurationStr != "" {
			if v, err := strconv.Atoi(minDurationStr); err == nil {
				minDuration = &v
			} else {
				appErr := appErrors.NewValidationError("min_duration must be a number", err)
				h.RespondAppError(w, r, appErr)
				return
			}
		}
		if maxDurationStr != "" {
			if v, err := strconv.Atoi(maxDurationStr); err == nil {
				maxDuration = &v
			} else {
				appErr := appErrors.NewValidationError("max_duration must be a number", err)
				h.RespondAppError(w, r, appErr)
				return
			}
		}
		if minDuration != nil && maxDuration != nil && *minDuration > *maxDuration {
			h.logger.Warn("min_duration greater than max_duration; swapping", map[string]interface{}{
				"min_duration": *minDuration,
				"max_duration": *maxDuration,
			})
			*minDuration, *maxDuration = *maxDuration, *minDuration
		}
	}

	req := BrowseRequest{
		Query:    query,
		Category: category,
		Brand:    brand,
		City:     city,
		ShopID:   shopID,
		Type:     productType,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Page:     page,
		PerPage:  perPage,
		Sort:     sort,
	}

	if err := validation.ValidateStruct(req); err != nil {
		message := validation.FormatValidationErrors(err)
		appErr := appErrors.NewValidationError(message, err)
		h.RespondAppError(w, r, appErr)
		return
	}
	ctx := r.Context()

	// Преобразуем category slug в category_id, если указан
	var categoryID *string
	var categoryIDs []string
	if category != "" {
		cat, err := h.categoriesSvc.GetBySlug(category)
		if err == nil {
			categoryID = &cat.ID
			// Получаем все дочерние категории для включения в фильтр
			childCats, err := h.categoriesSvc.GetByParentID(cat.ID)
			if err == nil {
				// Добавляем родительскую категорию и все дочерние
				categoryIDs = append(categoryIDs, cat.ID)
				for _, childCat := range childCats {
					categoryIDs = append(categoryIDs, childCat.ID)
				}
			} else {
				// Если не удалось получить дочерние, используем только родительскую
				categoryIDs = []string{cat.ID}
			}
		} else {
			// Если категория не найдена по slug, оставляем category как строку (для обратной совместимости)
			h.logger.Warn("Category not found by slug", map[string]interface{}{
				"slug":  category,
				"error": err.Error(),
			})
		}
	}

	// Преобразуем city slug в city_id, если указан
	var cityID *string
	if city != "" {
		cityObj, err := h.citiesSvc.GetBySlug(city)
		if err == nil {
			cityID = &cityObj.ID
		} else {
			h.logger.Warn("City not found by slug", map[string]interface{}{
				"slug":  city,
				"error": err.Error(),
			})
		}
	}

	res, err := h.service.Browse(ctx, products.BrowseParams{
		Query:       query,
		Category:    category,
		Brand:       brand,
		Type:        productType,
		CategoryID:  categoryID,
		CategoryIDs: categoryIDs,
		City:        city,
		CityID:      cityID,
		ShopID:      shopID,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		MinDuration: minDuration,
		MaxDuration: maxDuration,
		Page:        page,
		PerPage:     perPage,
		Sort:        sort,
	})
	if err != nil {
		appErr := appErrors.NewInternalError("Browse failed", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	// Логируем shop_names перед отправкой
	if res != nil && len(res.Items) > 0 {
		for i, item := range res.Items {
			if i < 3 {
				h.logger.Info("Browse handler: item before JSON", map[string]interface{}{
					"product_id":     item.ID,
					"product_name":   item.Name,
					"shop_names":     item.ShopNames,
					"shop_names_len": len(item.ShopNames),
					"shops_count":    item.ShopsCount,
				})
			}
		}
	}

	h.RespondJSON(w, http.StatusOK, res)
}
