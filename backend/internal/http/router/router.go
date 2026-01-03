package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/solomonczyk/izborator/internal/categories"
	"github.com/solomonczyk/izborator/internal/cities"
	appErrors "github.com/solomonczyk/izborator/internal/errors"
	"github.com/solomonczyk/izborator/internal/http/handlers"
	httpMiddleware "github.com/solomonczyk/izborator/internal/http/middleware"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/pricehistory"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/scrapingstats"
	"github.com/solomonczyk/izborator/internal/storage"
)

// Router обёртка над HTTP роутером
type Router struct {
	chi      *chi.Mux
	logger   *logger.Logger
	handlers *Handlers
}

// Handlers содержит все обработчики
type Handlers struct {
	Health     *handlers.HealthHandler
	Home       *handlers.HomeHandler
	Products   *handlers.ProductsHandler
	Stats      *handlers.StatsHandler
	Categories *handlers.CategoriesHandler
	Cities     *handlers.CitiesHandler
}

// New создаёт новый роутер
func New(log *logger.Logger, productsService *products.Service, priceHistoryService *pricehistory.Service, scrapingStatsService *scrapingstats.Service, categoriesService *categories.Service, citiesService *cities.Service, translator *i18n.Translator, db *storage.Postgres, redisClient *redis.Client) *Router {
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
	var pgPool *pgxpool.Pool
	if db != nil {
		pgPool = db.DB()
	}

	var redisPool *redis.Client
	if redisClient != nil {
		redisPool = redisClient
	}

	handlers := &Handlers{
		Health:     handlers.NewHealthHandler(pgPool, redisPool, log),
		Home:       handlers.NewHomeHandler(log, translator),
		Products:   handlers.NewProductsHandler(productsService, priceHistoryService, categoriesService, citiesService, log, translator),
		Stats:      handlers.NewStatsHandler(scrapingStatsService, log, translator),
		Categories: handlers.NewCategoriesHandler(categoriesService, log, translator),
		Cities:     handlers.NewCitiesHandler(citiesService, log, translator),
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
	// Health check endpoints
	r.Get("/api/health", h.Health.Check)
	r.Get("/api/health/live", h.Health.Alive)
	r.Get("/api/health/ready", h.Health.Ready)
	r.Get("/api/health/full", h.Health.Full)

	// Internal tenant health snapshot
	r.Route("/api/internal", func(ir chi.Router) {
		ir.Get("/tenant/health", h.Products.TenantHealth)
	})

	// API v1 роуты
	r.Route("/api/v1", func(api chi.Router) {
		api.With(httpMiddleware.CacheMiddleware(redisClient, log, time.Minute)).Get("/home", h.Home.GetHome)
		api.With(httpMiddleware.CacheMiddleware(redisClient, log, time.Minute)).Get("/home/meta", h.Home.GetHomeMeta)
		// Статистика парсинга
		api.Route("/stats", func(sr chi.Router) {
			sr.Get("/overall", h.Stats.GetOverallStats)
			sr.Get("/recent", h.Stats.GetRecentStats)
			sr.Get("/shops/{shop_id}", h.Stats.GetShopStats)
		})

		// Категории
		api.Route("/categories", func(cr chi.Router) {
			// Tree - 30 минут (категории меняются редко)
			cr.With(httpMiddleware.CacheMiddleware(redisClient, log, 30*time.Minute)).Get("/tree", h.Categories.GetTree)
		})

		// Города
		api.Route("/cities", func(cr chi.Router) {
			// GetAllActive - 30 минут (города меняются редко)
			cr.With(httpMiddleware.CacheMiddleware(redisClient, log, 30*time.Minute)).Get("/", h.Cities.GetAllActive)
		})

		// Товары
		api.Route("/products", func(pr chi.Router) {
			// Кэширование для популярных endpoints
			// Browse - 5 минут (часто меняется)
			pr.With(httpMiddleware.CacheMiddleware(redisClient, log, 5*time.Minute)).Get("/facets", h.Products.Facets)
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
		messageKey := "api.errors.not_found"
		message := translator.T(lang, messageKey)
		if message == "" || message == messageKey {
			message = translator.T("en", messageKey)
		}
		if message == "" || message == messageKey {
			message = "not found"
		}
		_ = json.NewEncoder(w).Encode(appErrors.NewErrorResponse(appErrors.CodeNotFound, message, nil))
	})
}

// ServeHTTP реализует интерфейс http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}
