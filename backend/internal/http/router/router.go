package router

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/solomonczyk/izborator/internal/http/handlers"
	httpMiddleware "github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/products"
)

// Router обёртка над HTTP роутером
type Router struct {
	chi     *chi.Mux
	logger  *logger.Logger
	handlers *Handlers
}

// Handlers содержит все обработчики
type Handlers struct {
	Health   *handlers.HealthHandler
	Products *handlers.ProductsHandler
}

// New создаёт новый роутер
func New(log *logger.Logger, productsService *products.Service) *Router {
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpMiddleware.Recovery(log))
	r.Use(httpMiddleware.CORS)
	r.Use(httpMiddleware.RequestLogger(log))
	r.Use(middleware.Compress(5))

	// Инициализация handlers
	handlers := &Handlers{
		Health:   handlers.NewHealthHandler(),
		Products: handlers.NewProductsHandler(productsService, log),
	}

	// Настройка роутов
	setupRoutes(r, handlers)

	return &Router{
		chi:      r,
		logger:   log,
		handlers: handlers,
	}
}

// setupRoutes настраивает все роуты приложения
func setupRoutes(r *chi.Mux, h *Handlers) {
	// Health check
	r.Get("/api/health", h.Health.Check)

	// API роуты для товаров
	r.Route("/api/products", func(r chi.Router) {
		r.Get("/", h.Products.Search)
		r.Get("/{id}", h.Products.GetByID)
		r.Get("/{id}/prices", h.Products.GetPrices)
	})

	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "not found",
		})
	})
}

// ServeHTTP реализует интерфейс http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
