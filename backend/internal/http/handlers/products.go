package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/products"
)

// ProductsHandler обработчик для работы с товарами
type ProductsHandler struct {
	service *products.Service
	logger  *logger.Logger
}

// NewProductsHandler создаёт новый обработчик товаров
func NewProductsHandler(service *products.Service, log *logger.Logger) *ProductsHandler {
	return &ProductsHandler{
		service: service,
		logger:  log,
	}
}

// Search обрабатывает поиск товаров
// GET /api/v1/products/search?q=query
// GET /api/products?q=query&limit=10&offset=0 (старый формат)
func (h *ProductsHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.respondError(w, http.StatusBadRequest, "missing q parameter")
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
			h.respondError(w, http.StatusInternalServerError, "search failed")
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
		h.respondError(w, http.StatusInternalServerError, err.Error())
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
		h.respondError(w, http.StatusBadRequest, "product ID is required")
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
			h.respondError(w, http.StatusNotFound, "product not found")
			return
		}
		h.logger.Error("GetProduct failed", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		h.respondError(w, http.StatusInternalServerError, "failed to load product")
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
		h.respondError(w, http.StatusBadRequest, "product ID is required")
		return
	}

	prices, err := h.service.GetPrices(id)
	if err != nil {
		if err == products.ErrInvalidProductID {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := map[string]interface{}{
		"product_id": id,
		"prices":     prices,
	}

	h.respondJSON(w, http.StatusOK, result)
}

// respondJSON отправляет JSON ответ
func (h *ProductsHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
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

	res, err := h.service.Browse(ctx, products.BrowseParams{
		Query:    query,
		Category: category,
		ShopID:   shopID,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Page:     page,
		PerPage:  perPage,
		Sort:     sort,
	})
	if err != nil {
		h.logger.Error("browse failed", map[string]interface{}{
			"error": err.Error(),
		})
		h.respondError(w, http.StatusInternalServerError, "browse failed")
		return
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
func (h *ProductsHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

