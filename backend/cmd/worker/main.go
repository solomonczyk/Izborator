package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/app"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

func main() {
	// Загрузка .env файла (игнорируем ошибку, если файл не найден)
	_ = godotenv.Load()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Инициализация приложения
	application, err := app.NewWorkerApp(cfg)
	if err != nil {
		panic(err)
	}
	defer application.Close()

	log := application.Logger()

	// Флаги
	daemonMode := flag.Bool("daemon", false, "Run in daemon mode (scheduler)")
	testURL := flag.String("url", "", "Single URL scrape test")
	shopIDStr := flag.String("shop", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", "Shop ID for single scrape")
	processRaw := flag.Bool("process", false, "Run processor once")
	batchSize := flag.Int("batch-size", 100, "Batch size for processing")
	reindex := flag.Bool("reindex", false, "Run full reindex once")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- 1. РЕЖИМЫ ONE-OFF (Ручной запуск) ---

	if *testURL != "" {
		if *shopIDStr == "" {
			log.Fatal("Shop ID is required for url scrape", nil)
		}
		runSingleScrape(ctx, application, *testURL, *shopIDStr, log)
		return
	}

	if *processRaw {
		runProcessor(ctx, application, *batchSize, log)
		return
	}

	if *reindex {
		runReindex(ctx, application, log)
		return
	}

	// --- 2. РЕЖИМ ДЕМОНА (Автоматизация) ---

	if *daemonMode {
		log.Info("🚀 Starting Worker Daemon...", nil)

		// Каналы для остановки
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		// Тикеры (Таймеры)
		// Процессинг запускаем часто (каждые 30 сек), чтобы быстро подхватывать новые данные
		processTicker := time.NewTicker(30 * time.Second)
		defer processTicker.Stop()

		// Скрапинг запускаем редко (каждые 6 часов), чтобы обновлять цены
		// Для теста поставим 10 минут, чтобы ты увидел результат быстрее
		scrapeTicker := time.NewTicker(10 * time.Minute)
		defer scrapeTicker.Stop()

		// Запуск горутины планировщика
		go func() {
			// Сразу при старте сделаем один прогон всего
			log.Info("Running initial startup tasks...", nil)
			runMonitoring(ctx, application, log) // Скрапинг списка
			runProcessor(ctx, application, *batchSize, log)  // Процессинг
			runReindex(ctx, application, log)    // Индексация

			for {
				select {
				case <-processTicker.C:
					runProcessor(ctx, application, *batchSize, log)

				case <-scrapeTicker.C:
					log.Info("⏰ Scheduled scraping started", nil)
					runMonitoring(ctx, application, log)
					// После скрапинга логично обновить индекс
					runReindex(ctx, application, log)

				case <-ctx.Done():
					return
				}
			}
		}()

		// Блокируем main, пока не придет сигнал стоп
		<-stop
		log.Info("Shutting down daemon...", nil)
		processTicker.Stop()
		scrapeTicker.Stop()
		return
	}

	// Если флагов нет
	log.Info("Worker started (waiting for jobs...) - use -url to test scrape, -process to process raw data, or -daemon to start scheduler", map[string]interface{}{})
	flag.Usage()
}

// --- ХЕЛПЕРЫ ---

func runSingleScrape(ctx context.Context, app *app.App, url, shopID string, log *logger.Logger) {
	log.Info("Manual scrape started", map[string]interface{}{
		"url":     url,
		"shop_id": shopID,
	})

	// Получаем конфиг магазина
	shopConfig, err := app.GetShopConfig(shopID)
	if err != nil {
		log.Fatal("Shop config not found", map[string]interface{}{
			"error":   err.Error(),
			"shop_id": shopID,
		})
	}

	log.Info("Shop config loaded", map[string]interface{}{
		"shop_name": shopConfig.Name,
		"base_url":  shopConfig.BaseURL,
	})

	// Парсим товар
	rawProduct, err := app.ScraperService.ScrapeAndSave(ctx, url, shopConfig)
	if err != nil {
		log.Fatal("❌ Scrape & save failed", map[string]interface{}{
			"error": err.Error(),
			"url":   url,
		})
	}

	log.Info("✅ SUCCESS! Product parsed & saved", map[string]interface{}{
		"name":     rawProduct.Name,
		"price":    rawProduct.Price,
		"currency": rawProduct.Currency,
		"brand":    rawProduct.Brand,
		"category": rawProduct.Category,
	})
}

func runProcessor(ctx context.Context, app *app.App, batchSize int, log *logger.Logger) {
	log.Info("🔄 Processor tick", nil)
	count, err := app.ProcessorService.ProcessRawProducts(ctx, batchSize)
	if err != nil {
		log.Error("Processing failed", map[string]interface{}{"error": err.Error()})
	} else if count > 0 {
		log.Info("Processed items", map[string]interface{}{"count": count})
	}
}

func runReindex(ctx context.Context, app *app.App, log *logger.Logger) {
	log.Info("🔍 Reindex tick", nil)
	if err := app.ReindexAll(); err != nil {
		log.Error("Reindex failed", map[string]interface{}{"error": err.Error()})
	} else {
		log.Info("✅ Reindex completed successfully", nil)
	}
}

func runMonitoring(ctx context.Context, app *app.App, log *logger.Logger) {
	log.Info("🕵️ Checking for outdated prices...", nil)

	// Получаем ссылки, которые не обновлялись более 6 часов
	// Берем пачками по 10 штук, чтобы не заспамить магазины
	outdatedItems, err := app.ProductsService.GetURLsForRescrape(ctx, 6*time.Hour, 10)
	if err != nil {
		log.Error("Failed to get urls for rescrape", map[string]interface{}{"error": err.Error()})
		return
	}

	if len(outdatedItems) == 0 {
		log.Info("All prices are fresh ✨", nil)
		return
	}

	log.Info("Found outdated items", map[string]interface{}{"count": len(outdatedItems)})

	// Обходим их
	for _, item := range outdatedItems {
		log.Info("Rescraping item", map[string]interface{}{"url": item.URL, "shop_id": item.ShopID})

		// Получаем конфиг магазина
		shopConfig, err := app.GetShopConfig(item.ShopID)
		if err != nil {
			log.Error("Shop config not found", map[string]interface{}{
				"shop_id": item.ShopID,
				"error":   err.Error(),
			})
			continue
		}

		// Парсим
		// ScrapeAndSave обновляет existing product через Processor.ProcessRawProducts (UPSERT)
		_, err = app.ScraperService.ScrapeAndSave(ctx, item.URL, shopConfig)
		if err != nil {
			log.Error("Rescrape failed", map[string]interface{}{
				"url":   item.URL,
				"error": err.Error(),
			})
		} else {
			log.Info("Rescraped successfully", map[string]interface{}{"url": item.URL})
		}

		// Вежливая пауза
		time.Sleep(5 * time.Second)
	}

	log.Info("✅ Monitoring scrape completed", map[string]interface{}{"count": len(outdatedItems)})
}
