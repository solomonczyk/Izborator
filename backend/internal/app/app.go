package app

import (
	"fmt"

	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/matching"
	"github.com/solomonczyk/izborator/internal/pricehistory"
	"github.com/solomonczyk/izborator/internal/processor"
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

	// Services (публичные - используются в cmd/*)
	ScraperService       *scraper.Service
	ProductsService      *products.Service
	ProcessorService     *processor.Service
	MatchingService      *matching.Service
	PriceHistoryService  *pricehistory.Service
	ScrapingStatsService *scrapingstats.Service

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
	a.productsStorage = storage.NewProductsAdapter(a.pg, a.meili)
	a.processorStorage = storage.NewProcessorAdapter(a.pg, a.meili)
	a.matchingStorage = storage.NewMatchingAdapter(a.pg)
	a.priceHistoryStorage = storage.NewPriceHistoryAdapter(a.pg)
	a.scrapingStatsStorage = storage.NewScrapingStatsAdapter(a.pg)
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
	app.productsStorage = storage.NewProductsAdapter(app.pg, app.meili)
	app.matchingStorage = storage.NewMatchingAdapter(app.pg)
	app.priceHistoryStorage = storage.NewPriceHistoryAdapter(app.pg)
	app.scrapingStatsStorage = storage.NewScrapingStatsAdapter(app.pg)

	// Инициализация сервисов (только для API)
	app.ProductsService = products.New(app.productsStorage, app.logger)
	app.MatchingService = matching.New(app.matchingStorage, app.logger)
	app.PriceHistoryService = pricehistory.New(app.priceHistoryStorage, app.logger)
	app.ScrapingStatsService = scrapingstats.New(app.scrapingStatsStorage, app.logger)

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
