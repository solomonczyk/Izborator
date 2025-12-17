package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/solomonczyk/izborator/internal/ai"
	"github.com/solomonczyk/izborator/internal/attributes"
	"github.com/solomonczyk/izborator/internal/categories"
	"github.com/solomonczyk/izborator/internal/cities"
	"github.com/solomonczyk/izborator/internal/classifier"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/matching"
	"github.com/solomonczyk/izborator/internal/pricehistory"
	"github.com/solomonczyk/izborator/internal/processor"
	"github.com/solomonczyk/izborator/internal/producttypes"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/scraper"
	"github.com/solomonczyk/izborator/internal/scrapingstats"
	"github.com/solomonczyk/izborator/internal/storage"
)

// App централизованная структура приложения
type App struct {
	// Приватные поля
	config *config.Config
	logger *logger.Logger

	// Storage (приватные)
	pg    *storage.Postgres
	meili *storage.Meilisearch
	redis *storage.Redis

	// Adapters (приватные)
	scraperStorage       scraper.Storage
	productsStorage      products.Storage
	processorStorage     processor.ProcessedStorage
	matchingStorage      matching.Storage
	priceHistoryStorage  pricehistory.Storage
	scrapingStatsStorage scrapingstats.Storage
	categoriesStorage    categories.Storage
	productTypesStorage  producttypes.Storage
	attributesStorage    attributes.Storage
	citiesStorage        cities.Storage
	classifierStorage    classifier.Storage

	// Services (публичные - используются в cmd/*)
	ScraperService       *scraper.Service
	ProductsService      *products.Service
	ProcessorService     *processor.Service
	MatchingService      *matching.Service
	PriceHistoryService  *pricehistory.Service
	ScrapingStatsService *scrapingstats.Service
	CategoriesService    *categories.Service
	ProductTypesService  *producttypes.Service
	AttributesService    *attributes.Service
	CitiesService        *cities.Service
	Classifier           *classifier.Service

	// AI
	AIClient *ai.Client

	// i18n
	Translator *i18n.Translator
}

// Logger возвращает логгер приложения
func (a *App) Logger() *logger.Logger {
	return a.logger
}

// Redis возвращает Redis клиент (может быть nil)
func (a *App) Redis() *storage.Redis {
	return a.redis
}

// GetTranslator возвращает переводчик
func (a *App) GetTranslator() *i18n.Translator {
	return a.Translator
}

// GetShopConfig получает конфигурацию магазина (для worker)
func (a *App) GetShopConfig(shopID string) (*scraper.ShopConfig, error) {
	return a.scraperStorage.GetShopConfig(shopID)
}

// GetClassifierStorage возвращает storage для classifier (для discovery worker)
func (a *App) GetClassifierStorage() classifier.Storage {
	return a.classifierStorage
}

// GetAIClient возвращает AI клиент (может быть nil, если API ключ не задан)
func (a *App) GetAIClient() *ai.Client {
	return a.AIClient
}

// ReindexAll переиндексирует все товары в Meilisearch
// Использует те же методы, что и cmd/indexer
func (a *App) ReindexAll() error {
	if a.meili == nil {
		return fmt.Errorf("Meilisearch is not available")
	}

	// Используем processor adapter для реиндексации
	// Он уже имеет доступ к Meilisearch и Postgres
	processorAdapter := storage.NewProcessorAdapter(a.pg, a.meili)
	
	// Получаем все товары из PostgreSQL и индексируем их
	// Используем тот же подход, что и в cmd/indexer
	ctx := context.Background()
	query := `
		SELECT id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
	`

	rows, err := a.pg.DB().Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var productList []*products.Product
	for rows.Next() {
		var p products.Product
		var specsJSON []byte
		var description, brand, category, imageURL *string
		var categoryID *string
		
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&description,
			&brand,
			&category,
			&categoryID,
			&imageURL,
			&specsJSON,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return fmt.Errorf("failed to scan product: %w", err)
		}

		if description != nil {
			p.Description = *description
		}
		if brand != nil {
			p.Brand = *brand
		}
		if category != nil {
			p.Category = *category
		}
		if imageURL != nil {
			p.ImageURL = *imageURL
		}
		p.CategoryID = categoryID
		
		if len(specsJSON) > 0 {
			if err := json.Unmarshal(specsJSON, &p.Specs); err != nil {
				a.logger.Warn("Failed to unmarshal specs", map[string]interface{}{"error": err.Error()})
			}
		} else {
			p.Specs = make(map[string]string)
		}

		productList = append(productList, &p)
	}

	// Удаляем все документы из индекса
	index := a.meili.Client().Index("products")
	if _, err := index.DeleteAllDocuments(); err != nil {
		// Если Meilisearch недоступен или ключ неверный, просто логируем и продолжаем
		a.logger.Warn("Failed to delete documents from Meilisearch (will continue anyway)", map[string]interface{}{
			"error": err.Error(),
		})
		// Не прерываем выполнение - просто пропускаем реиндексацию
		return nil
	}

	a.logger.Info("Deleted all documents from index")

	// Индексируем все товары
	for _, p := range productList {
		if err := processorAdapter.IndexProduct(p); err != nil {
			a.logger.Error("Failed to index product", map[string]interface{}{
				"product_id": p.ID,
				"error":      err.Error(),
			})
		}
	}

	a.logger.Info("Reindex completed", map[string]interface{}{
		"total_products": len(productList),
	})

	return nil
}

// NewApp создаёт новое приложение и инициализирует все зависимости
func NewApp(cfg *config.Config) (*App, error) {
	app := &App{
		config: cfg,
		logger: logger.New(cfg.LogLevel),
	}

	// Инициализация storage
	if err := app.initStorage(); err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	// Инициализация адаптеров
	app.initAdapters()

	// Инициализация сервисов
	app.initServices()

	return app, nil
}

// initStorage инициализирует подключения к хранилищам
func (a *App) initStorage() error {
	// PostgreSQL
	pg, err := storage.NewPostgres(&a.config.DB, a.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	a.pg = pg

	// Meilisearch (опционально, может быть nil)
	meili, err := storage.NewMeilisearch(&a.config.Meili, a.logger)
	if err != nil {
		a.logger.Warn("Meilisearch connection failed, continuing without it", map[string]interface{}{
			"error": err.Error(),
		})
		// Не критично, продолжаем без Meilisearch
	} else {
		a.meili = meili
	}

	// Redis (опционально, может быть nil)
	redis, err := storage.NewRedis(&a.config.Redis, a.logger)
	if err != nil {
		a.logger.Warn("Redis connection failed, continuing without it", map[string]interface{}{
			"error": err.Error(),
		})
		// Не критично, продолжаем без Redis
	} else {
		a.redis = redis
	}

	return nil
}

// initAdapters инициализирует адаптеры для доменных модулей
func (a *App) initAdapters() {
	a.scraperStorage = storage.NewScraperAdapter(a.pg)
	a.productsStorage = storage.NewProductsAdapter(a.pg, a.meili, a.logger)
	a.processorStorage = storage.NewProcessorAdapter(a.pg, a.meili)
	a.matchingStorage = storage.NewMatchingAdapter(a.pg)
	a.priceHistoryStorage = storage.NewPriceHistoryAdapter(a.pg)
	a.scrapingStatsStorage = storage.NewScrapingStatsAdapter(a.pg)
	a.categoriesStorage = storage.NewCategoriesAdapter(a.pg)
	a.productTypesStorage = storage.NewProductTypesAdapter(a.pg)
	a.attributesStorage = storage.NewAttributesAdapter(a.pg)
	a.citiesStorage = storage.NewCitiesAdapter(a.pg)
	a.classifierStorage = storage.NewClassifierAdapter(a.pg)
}

// initServices инициализирует доменные сервисы
func (a *App) initServices() {
	// Scraper service (queue пока nil)
	a.ScraperService = scraper.New(a.scraperStorage, nil, a.ScrapingStatsService, a.logger)

	// Products service
	a.ProductsService = products.New(a.productsStorage, a.logger)

	// Matching service
	a.MatchingService = matching.New(a.matchingStorage, a.logger)

	// Processor service
	a.ProcessorService = processor.New(
		a.scraperStorage,   // как processor.RawStorage
		a.processorStorage, // как processor.ProcessedStorage
		a.MatchingService,  // как processor.Matching
		a.logger,
	)

	// Price history service
	a.PriceHistoryService = pricehistory.New(a.priceHistoryStorage, a.logger)

	// Scraping stats service
	a.ScrapingStatsService = scrapingstats.New(a.scrapingStatsStorage, a.logger)

	// Categories service
	a.CategoriesService = categories.New(a.categoriesStorage, a.logger)

	// Product types service
	a.ProductTypesService = producttypes.New(a.productTypesStorage, a.logger)

	// Attributes service
	a.AttributesService = attributes.New(a.attributesStorage, a.logger)

	// Cities service
	a.CitiesService = cities.New(a.citiesStorage, a.logger)

	// Classifier service
	a.Classifier = classifier.New(a.classifierStorage, a.logger)

	// AI Client (опционально, если API ключ задан)
	if a.config.OpenAI.APIKey != "" {
		a.AIClient = ai.New(a.config.OpenAI.APIKey, a.config.OpenAI.Model)
		a.logger.Info("AI client initialized", map[string]interface{}{
			"model": a.config.OpenAI.Model,
		})
	} else {
		a.logger.Warn("OpenAI API key not set, AI features will be unavailable", nil)
	}
}

// initI18n инициализирует переводчик
func (a *App) initI18n() error {
	// Путь к локалям относительно корня проекта
	localesDir := "internal/i18n/locales"

	translator, err := i18n.NewTranslator(localesDir)
	if err != nil {
		a.logger.Warn("Failed to load i18n locales, continuing without translations", map[string]interface{}{
			"error": err.Error(),
		})
		// Создаём пустой translator, чтобы не ломать приложение
		translator, _ = i18n.NewTranslator("")
	}

	a.Translator = translator
	return nil
}

// Close закрывает все подключения
func (a *App) Close() error {
	var errs []error

	if a.pg != nil {
		if err := a.pg.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close PostgreSQL: %w", err))
		}
	}

	if a.meili != nil {
		if err := a.meili.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Meilisearch: %w", err))
		}
	}

	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// NewAPIApp создаёт приложение для API сервера (только нужные сервисы)
func NewAPIApp(cfg *config.Config) (*App, error) {
	app := &App{
		config: cfg,
		logger: logger.New(cfg.LogLevel),
	}

	// Инициализация storage
	if err := app.initStorage(); err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	// Инициализация адаптеров (только нужные для API)
	app.scraperStorage = storage.NewScraperAdapter(app.pg)
		app.productsStorage = storage.NewProductsAdapter(app.pg, app.meili, app.logger)
	app.matchingStorage = storage.NewMatchingAdapter(app.pg)
	app.priceHistoryStorage = storage.NewPriceHistoryAdapter(app.pg)
	app.scrapingStatsStorage = storage.NewScrapingStatsAdapter(app.pg)
	app.categoriesStorage = storage.NewCategoriesAdapter(app.pg)
	app.productTypesStorage = storage.NewProductTypesAdapter(app.pg)
	app.attributesStorage = storage.NewAttributesAdapter(app.pg)
	app.citiesStorage = storage.NewCitiesAdapter(app.pg)

	// Инициализация сервисов (только для API)
	app.ProductsService = products.New(app.productsStorage, app.logger)
	app.MatchingService = matching.New(app.matchingStorage, app.logger)
	app.PriceHistoryService = pricehistory.New(app.priceHistoryStorage, app.logger)
	app.ScrapingStatsService = scrapingstats.New(app.scrapingStatsStorage, app.logger)
	app.CategoriesService = categories.New(app.categoriesStorage, app.logger)
	app.ProductTypesService = producttypes.New(app.productTypesStorage, app.logger)
	app.AttributesService = attributes.New(app.attributesStorage, app.logger)
	app.CitiesService = cities.New(app.citiesStorage, app.logger)

	// i18n
	if err := app.initI18n(); err != nil {
		// Не критично, продолжаем без переводов
		app.logger.Warn("Failed to init i18n", map[string]interface{}{"error": err.Error()})
	}

	return app, nil
}

// NewWorkerApp создаёт приложение для воркера (все сервисы)
func NewWorkerApp(cfg *config.Config) (*App, error) {
	app := &App{
		config: cfg,
		logger: logger.New(cfg.LogLevel),
	}

	// Инициализация storage
	if err := app.initStorage(); err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	// Инициализация всех адаптеров
	app.initAdapters()

	// Инициализация всех сервисов
	app.initServices()

	return app, nil
}
