package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
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
	TenantID string   `json:"tenant_id" validate:"required"`
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
	Domain   string                      `json:"domain"`
	TenantID string                      `json:"tenant_id"`
	Facets   []domainpack.FacetDefinition `json:"facets"`
}

type TenantHealthResponse struct {
	TenantID        string `json:"tenant_id"`
	Domain          string `json:"domain"`
	FacetsCount     int    `json:"facets_count"`
	FacetsLimit     int    `json:"facets_limit"`
	FacetsOverLimit bool   `json:"facets_over_limit"`
	BrandsCount     int    `json:"brands_count"`
	BrandsLimit     int    `json:"brands_limit"`
	BrandsOverLimit bool   `json:"brands_over_limit"`
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

type rateEntry struct {
	tokens float64
	last   time.Time
}

type tenantRateLimiter struct {
	mu         sync.Mutex
	ratePerSec float64
	burst      float64
	entries    map[string]*rateEntry
	disabled   bool
}

func newTenantRateLimiter(perMinute int, burst int) *tenantRateLimiter {
	if perMinute <= 0 || burst <= 0 {
		return &tenantRateLimiter{disabled: true}
	}
	return &tenantRateLimiter{
		ratePerSec: float64(perMinute) / 60,
		burst:      float64(burst),
		entries:    make(map[string]*rateEntry),
	}
}

func (l *tenantRateLimiter) Allow(key string) bool {
	if l == nil || l.disabled {
		return true
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.entries[key]
	if entry == nil {
		entry = &rateEntry{tokens: l.burst, last: now}
		l.entries[key] = entry
	} else {
		elapsed := now.Sub(entry.last).Seconds()
		entry.tokens = minFloat(l.burst, entry.tokens+elapsed*l.ratePerSec)
		entry.last = now
	}

	if entry.tokens < 1 {
		return false
	}
	entry.tokens -= 1
	return true
}

func minFloat(a float64, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

var tenantLimiter = newTenantRateLimiter(
	envInt("TENANT_RATE_LIMIT_PER_MIN", 60),
	envInt("TENANT_RATE_LIMIT_BURST", 30),
)

type tenantLimitConfig struct {
	MaxFacets int `json:"max_facets"`
	MaxBrands int `json:"max_brands"`
}

func resolveTenantLimits(tenantID string, defaultFacets int, defaultBrands int, log *logger.Logger) (int, int) {
	raw := strings.TrimSpace(os.Getenv("TENANT_LIMITS_JSON"))
	if raw == "" || tenantID == "" {
		return defaultFacets, defaultBrands
	}
	limits := map[string]tenantLimitConfig{}
	if err := json.Unmarshal([]byte(raw), &limits); err != nil {
		if log != nil {
			log.Warn("tenant limits json parse failed", map[string]interface{}{
				"event":     "tenant_limits_parse_failed",
				"tenant_id": tenantID,
				"error":     err.Error(),
			})
		}
		return defaultFacets, defaultBrands
	}
	tenantLimits, ok := limits[tenantID]
	if !ok {
		return defaultFacets, defaultBrands
	}
	if tenantLimits.MaxFacets > 0 {
		defaultFacets = tenantLimits.MaxFacets
	}
	if tenantLimits.MaxBrands > 0 {
		defaultBrands = tenantLimits.MaxBrands
	}
	return defaultFacets, defaultBrands
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
// GET /api/v1/products/facets?type=<domain>&tenant_id=<tenant>
func (h *ProductsHandler) Facets(w http.ResponseWriter, r *http.Request) {
	domain := validation.SanitizeString(r.URL.Query().Get("type"))
	if domain == "" {
		appErr := appErrors.NewValidationError("type is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}
	tenantID := validation.SanitizeString(r.URL.Query().Get("tenant_id"))
	if tenantID == "" {
		appErr := appErrors.NewValidationError("tenant_id is required", nil)
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
	maxFacetsDefault := envInt("TENANT_MAX_FACETS_COUNT", 20)
	maxBrandsDefault := envInt("TENANT_MAX_BRANDS_COUNT", 200)
	maxFacets, maxBrands := resolveTenantLimits(tenantID, maxFacetsDefault, maxBrandsDefault, h.logger)
	facetsCount := len(facets)
	if maxFacets > 0 && facetsCount > maxFacets {
		h.logger.Warn("tenant facet count exceeded hard limit; truncating", map[string]interface{}{
			"tenant_id": tenantID,
			"domain":    domain,
			"count":     facetsCount,
			"limit":     maxFacets,
		})
		facets = facets[:maxFacets]
	}

	if domain == "goods" {
		brands, err := h.service.ListBrands(r.Context(), string(products.ProductTypeGood))
		if err != nil {
			appErr := appErrors.NewInternalError("failed to load brand facets", err)
			h.RespondAppError(w, r, appErr)
			return
		}
		brandsCount := len(brands)
		if maxBrands > 0 && brandsCount > maxBrands {
			h.logger.Warn("tenant brand count exceeded hard limit; truncating", map[string]interface{}{
				"tenant_id": tenantID,
				"domain":    domain,
				"count":     brandsCount,
				"limit":     maxBrands,
			})
			brands = brands[:maxBrands]
		}
		for i := range facets {
			if facets[i].SemanticType == "brand" {
				facets[i].Values = brands
				break
			}
		}
	}

	resp := FacetSchemaResponse{
		Domain:   domain,
		TenantID: tenantID,
		Facets:   facets,
	}
	h.RespondJSON(w, http.StatusOK, resp)
}

// TenantHealth returns a lightweight snapshot of tenant limits vs counts.
// GET /api/internal/tenant/health?tenant_id=<tenant>&type=<domain>
func (h *ProductsHandler) TenantHealth(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	tenantID := validation.SanitizeString(q.Get("tenant_id"))
	if tenantID == "" {
		appErr := appErrors.NewValidationError("tenant_id is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}
	domain := validation.SanitizeString(q.Get("type"))
	if domain == "" {
		domain = "goods"
	}
	if !domainpack.HasDomain(domain) {
		allowed := strings.Join(domainpack.Domains(), ", ")
		appErr := appErrors.NewValidationError("type must be one of: "+allowed, nil)
		h.RespondAppError(w, r, appErr)
		return
	}
	if !tenantLimiter.Allow(tenantID + ":facets") {
		h.logger.Warn("tenant rate limited", map[string]interface{}{
			"event":     "rate_limited",
			"tenant_id": tenantID,
			"endpoint":  "facets",
		})
		appErr := appErrors.NewAppErrorWithDetails(appErrors.CodeRateLimited, "rate limit exceeded", http.StatusTooManyRequests, nil, map[string]interface{}{
			"endpoint": "facets",
		})
		h.RespondAppError(w, r, appErr)
		return
	}

	facets, err := domainpack.Facets(domain)
	if err != nil {
		appErr := appErrors.NewInternalError("failed to load facet schema", err)
		h.RespondAppError(w, r, appErr)
		return
	}

	maxFacetsDefault := envInt("TENANT_MAX_FACETS_COUNT", 20)
	maxBrandsDefault := envInt("TENANT_MAX_BRANDS_COUNT", 200)
	maxFacets, maxBrands := resolveTenantLimits(tenantID, maxFacetsDefault, maxBrandsDefault, h.logger)

	brandsCount := 0
	if domain == "goods" {
		brands, err := h.service.ListBrands(r.Context(), string(products.ProductTypeGood))
		if err != nil {
			appErr := appErrors.NewInternalError("failed to load brands", err)
			h.RespondAppError(w, r, appErr)
			return
		}
		brandsCount = len(brands)
	}

	resp := TenantHealthResponse{
		TenantID:        tenantID,
		Domain:          domain,
		FacetsCount:     len(facets),
		FacetsLimit:     maxFacets,
		FacetsOverLimit: len(facets) > maxFacets,
		BrandsCount:     brandsCount,
		BrandsLimit:     maxBrands,
		BrandsOverLimit: brandsCount > maxBrands,
	}
	h.RespondJSON(w, http.StatusOK, resp)
}

func (h *ProductsHandler) Browse(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	query := validation.SanitizeString(q.Get("query"))
	tenantID := validation.SanitizeString(q.Get("tenant_id"))
	category := validation.SanitizeString(q.Get("category"))
	brand := validation.SanitizeString(q.Get("brand"))
	city := validation.SanitizeString(q.Get("city"))
	shopID := validation.SanitizeString(q.Get("shop_id"))
	productType := validation.SanitizeString(q.Get("type")) // "good" | "service" | ""
	sort := validation.SanitizeString(q.Get("sort"))

	if tenantID == "" {
		appErr := appErrors.NewValidationError("tenant_id is required", nil)
		h.RespondAppError(w, r, appErr)
		return
	}

	// Валидация типа продукта
	if productType != "" && productType != "good" && productType != "service" {
		appErr := appErrors.NewValidationError("type must be 'good' or 'service'", nil)
		h.RespondAppError(w, r, appErr)
		return
	}
	if !tenantLimiter.Allow(tenantID + ":browse") {
		h.logger.Warn("tenant rate limited", map[string]interface{}{
			"event":     "rate_limited",
			"tenant_id": tenantID,
			"endpoint":  "browse",
		})
		appErr := appErrors.NewAppErrorWithDetails(appErrors.CodeRateLimited, "rate limit exceeded", http.StatusTooManyRequests, nil, map[string]interface{}{
			"endpoint": "browse",
		})
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
		TenantID: tenantID,
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
