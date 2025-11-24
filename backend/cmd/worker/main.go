package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/scraper"
	"github.com/solomonczyk/izborator/internal/storage"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация логгера
	logger := logger.New(cfg.LogLevel)

	// Флаги для ручного запуска
	testURL := flag.String("url", "", "URL товара для теста")
	shopIDStr := flag.String("shop", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", "UUID магазина")
	flag.Parse()

	ctx := context.Background()

	// Подключение к БД
	logger.Info("Connecting to PostgreSQL", map[string]interface{}{
		"host":     cfg.DB.Host,
		"port":      cfg.DB.Port,
		"user":      cfg.DB.User,
		"database":  cfg.DB.Database,
		"dsn":       cfg.DB.DSN(),
	})
	pg, err := storage.NewPostgres(&cfg.DB, logger)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", map[string]interface{}{"error": err.Error(), "dsn": cfg.DB.DSN()})
	}
	defer pg.Close()

	// Создание адаптеров
	scraperStorage := storage.NewScraperAdapter(pg)

	// Создание сервисов (queue пока nil, так как не реализован)
	scraperService := scraper.New(scraperStorage, nil, logger)

	// Режим тестирования одного URL
	if *testURL != "" {
		logger.Info("🚀 Starting manual test scrape...", map[string]interface{}{
			"url":     *testURL,
			"shop_id": *shopIDStr,
		})

		// Получаем конфиг магазина
		shopConfig, err := scraperStorage.GetShopConfig(*shopIDStr)
		if err != nil {
			logger.Fatal("Shop config not found", map[string]interface{}{
				"error":  err,
				"shop_id": *shopIDStr,
			})
		}

		logger.Info("Shop config loaded", map[string]interface{}{
			"shop_name": shopConfig.Name,
			"base_url":  shopConfig.BaseURL,
		})

		// Парсим товар
		rawProduct, err := scraperService.ParseProduct(ctx, *testURL, shopConfig)
		if err != nil {
			logger.Fatal("❌ Scraping failed", map[string]interface{}{
				"error": err.Error(),
				"url":   *testURL,
			})
		}

		logger.Info("✅ SUCCESS! Product parsed", map[string]interface{}{
			"name":     rawProduct.Name,
			"price":    rawProduct.Price,
			"currency": rawProduct.Currency,
			"brand":    rawProduct.Brand,
			"category": rawProduct.Category,
		})

		// Сохраняем результат
		if err := scraperService.SaveRawProduct(ctx, rawProduct); err != nil {
			logger.Error("Failed to save raw product", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("💾 Saved to raw_products table", map[string]interface{}{})
		}

		return
	}

	// Обычный режим воркера (ожидание задач из очереди)
	logger.Info("Worker started (waiting for jobs...) - use -url to test scrape", map[string]interface{}{})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...", map[string]interface{}{})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// TODO: Graceful shutdown воркеров
	_ = ctx

	logger.Info("Worker exited", map[string]interface{}{})
}
