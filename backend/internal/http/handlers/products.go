package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/solomonczyk/izborator/internal/categories"
	"github.com/solomonczyk/izborator/internal/cities"
	httpMiddleware "github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/pricehistory"
	"github.com/solomonczyk/izborator/internal/products"
)

// ProductsHandler обработчик для работы с товарами
type ProductsHandler struct {
	service         *products.Service
	priceHistorySvc *pricehistory.Service
	categoriesSvc   *categories.Service
	citiesSvc       *cities.Service
	logger          *logger.Logger
	translator      *i18n.Translator
}

// NewProductsHandler создаёт новый обработчик товаров
func NewProductsHandler(service *products.Service, priceHistorySvc *pricehistory.Service, categoriesSvc *categories.Service, citiesSvc *cities.Service, log *logger.Logger, translator *i18n.Translator) *ProductsHandler {
	return &ProductsHandler{
		service:         service,
		priceHistorySvc: priceHistorySvc,
		categoriesSvc:   categoriesSvc,
		citiesSvc:       citiesSvc,
		logger:          log,
		translator:      translator,
	}
}

// Search обрабатывает поиск товаров
// GET /api/v1/products/search?q=query
// GET /api/products?q=query&limit=10&offset=0 (старый формат)
func (h *ProductsHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.respondError(w, r, http.StatusBadRequest, "api.errors.missing_query")
		return
	}

	ctx := r.Context()

	// Для нового endpoint /api/v1/products/search используем простой поиск
	if r.URL.Path == "/api/v1/products/search" {
		results, err := h.service.Search(ctx, query)
		if err != nil {
			h.logger.Error("search failed", map[string]interface{}{
				"q":     query,
				"error": err.Error(),
			})
			h.respondError(w, r, http.StatusInternalServerError, "api.errors.search_failed")
			return
		}

		h.respondJSON(w, http.StatusOK, results)
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
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.search_failed")
		return
	}

	h.respondJSON(w, http.StatusOK, result)
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
		h.respondError(w, r, http.StatusBadRequest, "api.errors.product_id_required")
		return
	}

	ctx := r.Context()
	_ = ctx // для будущего использования

	// 1. Получаем товар
	product, err := h.service.GetByID(id)
	if err != nil {
		if err == products.ErrProductNotFound || err == products.ErrInvalidProductID {
			h.logger.Error("GetProduct failed", map[string]interface{}{
				"id":    id,
				"error": err.Error(),
			})
			h.respondError(w, r, http.StatusNotFound, "api.errors.product_not_found")
			return
		}
		h.logger.Error("GetProduct failed", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.product_load_failed")
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

	h.respondJSON(w, http.StatusOK, resp)
}

// GetPrices обрабатывает получение цен товара из разных магазинов
// GET /api/products/:id/prices
func (h *ProductsHandler) GetPrices(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		h.respondError(w, r, http.StatusBadRequest, "api.errors.product_id_required")
		return
	}

	prices, err := h.service.GetPrices(id)
	if err != nil {
		if err == products.ErrInvalidProductID {
			h.respondError(w, r, http.StatusBadRequest, "api.errors.product_id_required")
			return
		}
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.internal")
		return
	}

	result := map[string]interface{}{
		"product_id": id,
		"prices":     prices,
	}

	h.respondJSON(w, http.StatusOK, result)
}

// GetPriceHistory обрабатывает получение истории цен товара
// GET /api/v1/products/{id}/price-history?period=month&shops=shop1,shop2
func (h *ProductsHandler) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		h.respondError(w, r, http.StatusBadRequest, "api.errors.product_id_required")
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
		h.logger.Error("GetPriceHistory failed", map[string]interface{}{
			"id":     id,
			"error":  err.Error(),
			"period": period,
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.price_history_failed")
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

	h.respondJSON(w, http.StatusOK, result)
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
	var allDates []time.Time
	var firstPrice, lastPrice float64
	var firstDate, lastDate time.Time

	// Собираем все цены из всех магазинов
	for _, points := range chart.Shops {
		for _, point := range points {
			allPrices = append(allPrices, point.Price)
			allDates = append(allDates, point.Timestamp)

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

// respondJSON отправляет JSON ответ
func (h *ProductsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", map[string]interface{}{
			"error": err,
		})
	}
}

// Browse обрабатывает каталог товаров с фильтрами
// GET /api/v1/products/browse?query=motorola&category=phones&min_price=10000&max_price=30000&shop_id=...&page=1&per_page=20&sort=price_asc
func (h *ProductsHandler) Browse(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	query := q.Get("query")
	category := q.Get("category")
	city := q.Get("city")
	shopID := q.Get("shop_id")
	sort := q.Get("sort")

	minPriceStr := q.Get("min_price")
	maxPriceStr := q.Get("max_price")

	page := parseIntDefault(q.Get("page"), 1)
	perPage := parseIntDefault(q.Get("per_page"), 20)
	if perPage > 100 {
		perPage = 100
	}

	var (
		minPrice *float64
		maxPrice *float64
	)

	if minPriceStr != "" {
		if v, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			minPrice = &v
		}
	}
	if maxPriceStr != "" {
		if v, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			maxPrice = &v
		}
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
		CategoryID:  categoryID,
		CategoryIDs: categoryIDs,
		City:        city,
		CityID:      cityID,
		ShopID:      shopID,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		Page:        page,
		PerPage:     perPage,
		Sort:        sort,
	})
	if err != nil {
		h.logger.Error("browse failed", map[string]interface{}{
			"error": err.Error(),
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.browse_failed")
		return
	}

	// Логируем shop_names перед отправкой
	if res != nil && len(res.Items) > 0 {
		for i, item := range res.Items {
			if i < 3 {
				h.logger.Info("Browse handler: item before JSON", map[string]interface{}{
					"product_id":   item.ID,
					"product_name": item.Name,
					"shop_names":   item.ShopNames,
					"shop_names_len": len(item.ShopNames),
					"shops_count":  item.ShopsCount,
				})
			}
		}
	}

	h.respondJSON(w, http.StatusOK, res)
}

// parseIntDefault парсит строку в int с значением по умолчанию
func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return def
	}
	return v
}

// respondError отправляет JSON ошибку
func (h *ProductsHandler) respondError(w http.ResponseWriter, r *http.Request, status int, key string) {
	lang := httpMiddleware.GetLangFromContext(r.Context())
	message := h.translator.T(lang, key)
	if message == key || message == "" {
		// fallback на английский
		message = h.translator.T("en", key)
		if message == "" {
			message = key
		}
	}
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}
