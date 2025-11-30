package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/app"
	"github.com/solomonczyk/izborator/internal/config"
)

func main() {
	// Загрузка .env файла (игнорируем ошибку, если файл не найден)
	_ = godotenv.Load()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация приложения
	application, err := app.NewWorkerApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer application.Close()

	// Флаги для ручного запуска
	testURL := flag.String("url", "", "URL товара для теста")
	shopIDStr := flag.String("shop", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", "UUID магазина")
	processRaw := flag.Bool("process", false, "Обработать необработанные сырые данные")
	batchSize := flag.Int("batch-size", 10, "Размер батча для обработки")
	flag.Parse()

	ctx := context.Background()

	// Режим тестирования одного URL
	if *testURL != "" {
		application.Logger().Info("🚀 Starting manual test scrape...", map[string]interface{}{
			"url":     *testURL,
			"shop_id": *shopIDStr,
		})

		// Получаем конфиг магазина
		shopConfig, err := application.GetShopConfig(*shopIDStr)
		if err != nil {
			application.Logger().Fatal("Shop config not found", map[string]interface{}{
				"error":   err,
				"shop_id": *shopIDStr,
			})
		}

		application.Logger().Info("Shop config loaded", map[string]interface{}{
			"shop_name": shopConfig.Name,
			"base_url":  shopConfig.BaseURL,
		})

		// Парсим товар
		rawProduct, err := application.ScraperService.ScrapeAndSave(ctx, *testURL, shopConfig)
		if err != nil {
			application.Logger().Fatal("❌ Scrape & save failed", map[string]interface{}{
				"error": err.Error(),
				"url":   *testURL,
			})
		}

		application.Logger().Info("✅ SUCCESS! Product parsed & saved", map[string]interface{}{
			"name":     rawProduct.Name,
			"price":    rawProduct.Price,
			"currency": rawProduct.Currency,
			"brand":    rawProduct.Brand,
			"category": rawProduct.Category,
		})

		return
	}

	// Режим обработки сырых данных
	if *processRaw {
		application.Logger().Info("🔄 Starting raw products processing...", map[string]interface{}{
			"batch_size": *batchSize,
		})

		processed, err := application.ProcessorService.ProcessRawProducts(ctx, *batchSize)
		if err != nil {
			application.Logger().Fatal("Failed to process raw products", map[string]interface{}{
				"error": err.Error(),
			})
		}

		application.Logger().Info("✅ Processing completed", map[string]interface{}{
			"processed": processed,
		})

		return
	}

	// Обычный режим воркера (ожидание задач из очереди)
	application.Logger().Info("Worker started (waiting for jobs...) - use -url to test scrape or -process to process raw data", map[string]interface{}{})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.Logger().Info("Shutting down worker...", map[string]interface{}{})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// TODO: Graceful shutdown воркеров
	_ = ctx

	application.Logger().Info("Worker exited", map[string]interface{}{})
}
