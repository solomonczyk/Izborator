package router

import (
	"encoding/json"
	"net/http"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/solomonczyk/izborator/internal/http/handlers"
	httpMiddleware "github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/pricehistory"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/scrapingstats"
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
	Stats    *handlers.StatsHandler
}

// New создаёт новый роутер
func New(log *logger.Logger, productsService *products.Service, priceHistoryService *pricehistory.Service, scrapingStatsService *scrapingstats.Service, translator *i18n.Translator, redisClient *redis.Client) *Router {
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpMiddleware.Recovery(log))
	r.Use(httpMiddleware.CORS)
	r.Use(httpMiddleware.DetectLanguage) // Определение языка
	r.Use(httpMiddleware.RequestLogger(log))
	r.Use(middleware.Compress(5))

	// Инициализация handlers
	handlers := &Handlers{
		Health:   handlers.NewHealthHandler(),
		Products: handlers.NewProductsHandler(productsService, priceHistoryService, log, translator),
		Stats:    handlers.NewStatsHandler(scrapingStatsService, log, translator),
	}

	// Настройка роутов
	setupRoutes(r, handlers, translator, redisClient, log)

	return &Router{
		chi:      r,
		logger:   log,
		handlers: handlers,
	}
}

// setupRoutes настраивает все роуты приложения
func setupRoutes(r *chi.Mux, h *Handlers, translator *i18n.Translator, redisClient *redis.Client, log *logger.Logger) {
	// Health check (не кэшируем)
	r.Get("/api/health", h.Health.Check)

	// API v1 роуты
	r.Route("/api/v1", func(api chi.Router) {
		// Статистика парсинга
		api.Route("/stats", func(sr chi.Router) {
			sr.Get("/overall", h.Stats.GetOverallStats)
			sr.Get("/recent", h.Stats.GetRecentStats)
			sr.Get("/shops/{shop_id}", h.Stats.GetShopStats)
		})

		// Товары
		api.Route("/products", func(pr chi.Router) {
			// Кэширование для популярных endpoints
			// Browse - 5 минут (часто меняется)
			pr.With(httpMiddleware.CacheMiddleware(redisClient, log, 5*time.Minute)).Get("/browse", h.Products.Browse)
			// Search - 5 минут
			pr.With(httpMiddleware.CacheMiddleware(redisClient, log, 5*time.Minute)).Get("/search", h.Products.Search)
			// GetByID - 10 минут (товары меняются реже)
			pr.With(httpMiddleware.CacheMiddleware(redisClient, log, 10*time.Minute)).Get("/{id}", h.Products.GetByID)
			// Prices - 2 минуты (цены обновляются часто)
			pr.With(httpMiddleware.CacheMiddleware(redisClient, log, 2*time.Minute)).Get("/{id}/prices", h.Products.GetPrices)
			// Price history - 15 минут (история меняется редко)
			pr.With(httpMiddleware.CacheMiddleware(redisClient, log, 15*time.Minute)).Get("/{id}/price-history", h.Products.GetPriceHistory)
		})
	})

	// Старые роуты для обратной совместимости
	r.Route("/api/products", func(r chi.Router) {
		r.Get("/", h.Products.Search)
		r.Get("/{id}", h.Products.GetByID)
		r.Get("/{id}/prices", h.Products.GetPrices)
	})

	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		lang := httpMiddleware.GetLangFromContext(r.Context())
		message := translator.T(lang, "api.errors.not_found")
		if message == "" {
			message = translator.T("en", "api.errors.not_found")
		}
		if message == "" {
			message = "not found"
		}
		json.NewEncoder(w).Encode(map[string]string{
			"error": message,
		})
	})
}

// ServeHTTP реализует интерфейс http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
