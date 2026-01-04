package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/app"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/queue"
	"github.com/solomonczyk/izborator/internal/scraper"
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
	discover := flag.Bool("discover", false, "Run catalog discovery once")

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

	if *discover {
		runCatalogDiscovery(ctx, application, log)
		return
	}

	// --- 2. РЕЖИМ ДЕМОНА (Автоматизация) ---

	if *daemonMode {
		log.Info("🚀 Starting Worker Daemon...", nil)

		// Каналы для остановки
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		// WaitGroup для отслеживания активных задач
		var wg sync.WaitGroup

		queueClient := application.QueueClient()
		if queueClient != nil && cfg.Queue.Topic != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runQueueConsumer(ctx, application, queueClient, cfg.Queue.Topic, log)
			}()
		} else {
			log.Info("Queue consumer disabled", map[string]interface{}{
				"queue_type": cfg.Queue.Type,
				"topic":      cfg.Queue.Topic,
			})
		}

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
			wg.Add(4)
			go func() { defer wg.Done(); runCatalogDiscovery(ctx, application, log) }()      // Обнаружение новых товаров в каталогах
			go func() { defer wg.Done(); runMonitoring(ctx, application, log) }()            // Скрапинг списка
			go func() { defer wg.Done(); runProcessor(ctx, application, *batchSize, log) }() // Процессинг
			go func() { defer wg.Done(); runReindex(ctx, application, log) }()               // Индексация

			for {
				select {
				case <-processTicker.C:
					wg.Add(1)
					go func() {
						defer wg.Done()
						runProcessor(ctx, application, *batchSize, log)
					}()

				case <-scrapeTicker.C:
					log.Info("⏰ Scheduled scraping started", nil)
					wg.Add(3)
					go func() { defer wg.Done(); runCatalogDiscovery(ctx, application, log) }() // Обнаружение новых товаров
					go func() { defer wg.Done(); runMonitoring(ctx, application, log) }()       // Обновление цен
					go func() { defer wg.Done(); runReindex(ctx, application, log) }()          // Индексация

				case <-ctx.Done():
					return
				}
			}
		}()

		// Блокируем main, пока не придет сигнал стоп
		<-stop
		log.Info("Shutting down daemon...", nil)

		// Отменяем контекст для остановки всех горутин
		cancel()

		// Останавливаем тикеры
		processTicker.Stop()
		scrapeTicker.Stop()

		// Даем время на завершение активных задач (graceful shutdown)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Ждем завершения всех активных задач с таймаутом
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Info("✅ Graceful shutdown completed - all tasks finished", nil)
		case <-shutdownCtx.Done():
			log.Warn("⚠️ Shutdown timeout reached, forcing exit", nil)
		}

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
	// Используем контекст с таймаутом для реиндексации
	reindexCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	
	if err := app.ReindexAllWithContext(reindexCtx); err != nil {
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

func runCatalogDiscovery(ctx context.Context, app *app.App, log *logger.Logger) {
	log.Info("🔍 Starting catalog discovery...", nil)

	// Получаем список всех активных магазинов
	shops, err := app.ScraperService.ListShops(ctx)
	if err != nil {
		log.Error("Failed to list shops", map[string]interface{}{"error": err.Error()})
		return
	}

	for _, shop := range shops {
		if !shop.Enabled {
			continue
		}

		// Получаем URL каталога из конфига
		catalogURL := shop.Selectors["catalog_url"]

		// Если catalog_url не указан, пробуем использовать base_url как точку входа
		if catalogURL == "" {
			log.Info("No catalog_url configured, trying base_url", map[string]interface{}{"shop": shop.Name})
			catalogURL = shop.BaseURL
		}

		if catalogURL == "" {
			log.Info("No catalog URL available, skipping", map[string]interface{}{"shop": shop.Name})
			continue
		}

		log.Info("Discovering products from catalog", map[string]interface{}{
			"shop":        shop.Name,
			"catalog_url": catalogURL,
		})

		// Парсим каталог (максимум 3 страницы за раз, чтобы не перегружать)
		result, err := app.ScraperService.ParseCatalog(ctx, catalogURL, shop, 3)
		if err != nil {
			log.Error("Catalog parsing failed", map[string]interface{}{
				"shop":  shop.Name,
				"error": err.Error(),
			})
			continue
		}

		if result.TotalFound == 0 {
			log.Info("No products found in catalog", map[string]interface{}{"shop": shop.Name})
			continue
		}

		log.Info("Found products in catalog", map[string]interface{}{
			"shop":        shop.Name,
			"total_found": result.TotalFound,
		})

		// Сохраняем найденные URL в базу для последующего парсинга
		savedCount := 0
		for _, productURL := range result.ProductURLs {
			// Создаем минимальный RawProduct для сохранения URL
			rawProduct := &scraper.RawProduct{
				ShopID:    shop.ID,
				ShopName:  shop.Name,
				URL:       productURL,
				ParsedAt:  time.Now(),
				ScrapedAt: time.Now(),
			}

			// Извлекаем external_id из URL
			parts := strings.Split(productURL, "/")
			if len(parts) > 0 {
				rawProduct.ExternalID = parts[len(parts)-1]
			}

			// Сохраняем через ScraperService (он использует ScrapeAndSave, который сохраняет в raw_products)
			// Но нам нужно просто сохранить URL, поэтому используем прямой вызов storage
			// Вместо этого, запустим быстрый парсинг каждого товара
			_, err = app.ScraperService.ScrapeAndSave(ctx, productURL, shop)
			if err != nil {
				log.Error("Failed to scrape product from catalog", map[string]interface{}{
					"url":   productURL,
					"error": err.Error(),
				})
				continue
			}
			savedCount++

			// Небольшая пауза между товарами
			time.Sleep(2 * time.Second)
		}

		log.Info("✅ Catalog discovery completed", map[string]interface{}{
			"shop":  shop.Name,
			"found": result.TotalFound,
			"saved": savedCount,
		})

		// Пауза между магазинами
		time.Sleep(5 * time.Second)
	}
}


func runQueueConsumer(ctx context.Context, app *app.App, queueClient queue.Client, topic string, log *logger.Logger) {
	log.Info("Queue consumer started", map[string]interface{}{
		"topic": topic,
	})

	err := queueClient.Consume(ctx, topic, func(payload []byte) error {
		var raw scraper.RawProduct
		if err := json.Unmarshal(payload, &raw); err != nil {
			log.Error("Queue payload decode failed", map[string]interface{}{
				"topic": topic,
				"error": err.Error(),
			})
			return err
		}
		if app.ProcessorService == nil {
			return errors.New("processor service is not initialized")
		}
		return app.ProcessorService.ProcessRawProduct(ctx, &raw)
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Warn("Queue consumer stopped", map[string]interface{}{
			"topic": topic,
			"error": err.Error(),
		})
	}
}
