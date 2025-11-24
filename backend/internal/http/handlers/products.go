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
// GET /api/products?q=query&limit=10&offset=0
func (h *ProductsHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
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

	result, err := h.service.Search(query, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// GetByID обрабатывает получение товара по ID
// GET /api/products/:id
func (h *ProductsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		h.respondError(w, http.StatusBadRequest, "product ID is required")
		return
	}

	product, err := h.service.GetByID(id)
	if err != nil {
		if err == products.ErrProductNotFound || err == products.ErrInvalidProductID {
			h.respondError(w, http.StatusNotFound, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, product)
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

// respondError отправляет JSON ошибку
func (h *ProductsHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{
		"error": message,
	})
}

