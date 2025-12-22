package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"izborator/internal/http/middleware"
	"izborator/internal/products"
)

type ProductsHandler struct {
	service   products.Service
	logger    Logger
	translator Translator
	citiesSvc CitiesService
}

type Logger interface {
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Debug(msg string, fields map[string]interface{})
}

type Translator interface {
	T(lang string, key string) string
}

type CitiesService interface {
	GetBySlug(slug string) (City, error)
}

type City struct {
	ID string
}

func NewProductsHandler(service products.Service, logger Logger, translator Translator, citiesSvc CitiesService) *ProductsHandler {
	return &ProductsHandler{
		service:   service,
		logger:    logger,
		translator: translator,
		citiesSvc:  citiesSvc,
	}
}

// GetProduct возвращает информацию о товаре по ID
func (h *ProductsHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	if productID == "" {
		h.respondError(w, r, http.StatusBadRequest, "api.errors.invalid_product_id")
		return
	}

	ctx := r.Context()
	product, err := h.service.GetByID(ctx, productID)
	if err != nil {
		h.logger.Error("get product failed", map[string]interface{}{
			"product_id": productID,
			"error":      err.Error(),
		})
		h.respondError(w, r, http.StatusNotFound, "api.errors.product_not_found")
		return
	}

	h.respondJSON(w, http.StatusOK, product)
}

// GetProductPrices возвращает историю цен товара
func (h *ProductsHandler) GetProductPrices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	if productID == "" {
		h.respondError(w, r, http.StatusBadRequest, "api.errors.invalid_product_id")
		return
	}

	ctx := r.Context()
	prices, err := h.service.GetPriceHistory(ctx, productID)
	if err != nil {
		h.logger.Error("get price history failed", map[string]interface{}{
			"product_id": productID,
			"error":      err.Error(),
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.price_history_failed")
		return
	}

	h.respondJSON(w, http.StatusOK, prices)
}

// GetProductStats возвращает статистику по ценам товара
func (h *ProductsHandler) GetProductStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	if productID == "" {
		h.respondError(w, r, http.StatusBadRequest, "api.errors.invalid_product_id")
		return
	}

	ctx := r.Context()
	chart, err := h.service.GetPriceChart(ctx, productID)
	if err != nil {
		h.logger.Error("get price chart failed", map[string]interface{}{
			"product_id": productID,
			"error":      err.Error(),
		})
		h.respondError(w, r, http.StatusInternalServerError, "api.errors.price_chart_failed")
		return
	}

	stats := h.calculateStats(chart)
	h.respondJSON(w, http.StatusOK, stats)
}

// PriceStats представляет статистику по ценам
type PriceStats struct {
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	AvgPrice    float64 `json:"avg_price"`
	PriceChange float64 `json:"price_change"`
	FirstPrice  float64 `json:"first_price"`
	LastPrice   float64 `json:"last_price"`
	FirstDate   string  `json:"first_date"`
	LastDate    string  `json:"last_date"`
}

// calculateStats вычисляет статистику по ценам из графика
func (h *ProductsHandler) calculateStats(chart *products.PriceChart) PriceStats {
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
		FirstDate:   firstDate.Format(time.RFC3339),
		LastDate:    lastDate.Format(time.RFC3339),
	}
}

// Browse возвращает список товаров с фильтрацией и пагинацией
func (h *ProductsHandler) Browse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Парсим query параметры
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	categoryID := r.URL.Query().Get("category_id")
	shopID := r.URL.Query().Get("shop_id")
	city := r.URL.Query().Get("city")
	minPriceStr := r.URL.Query().Get("min_price")
	maxPriceStr := r.URL.Query().Get("max_price")
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")
	sort := r.URL.Query().Get("sort")

	// Парсим цены
	var minPrice, maxPrice *float64
	if minPriceStr != "" {
		if val, err := strconv.ParseFloat(minPriceStr, 64); err == nil && val >= 0 {
			minPrice = &val
		}
	}
	if maxPriceStr != "" {
		if val, err := strconv.ParseFloat(maxPriceStr, 64); err == nil && val >= 0 {
			maxPrice = &val
		}
	}

	// Парсим пагинацию
	page := parseIntDefault(pageStr, 1)
	perPage := parseIntDefault(perPageStr, 20)
	if perPage > 100 {
		perPage = 100
	}

	// Парсим category_id (может быть несколько через запятую)
	var categoryIDs []string
	if categoryID != "" {
		ids := splitString(categoryID, ",")
		for _, id := range ids {
			if id != "" {
				categoryIDs = append(categoryIDs, id)
			}
		}
	}

	// Получаем city_id по slug
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
		message = key
	}

	h.respondJSON(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}

// respondJSON отправляет JSON ответ
func (h *ProductsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// splitString разбивает строку по разделителю
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := []string{}
	current := ""
	for _, char := range s {
		if string(char) == sep {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
