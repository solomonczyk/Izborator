package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/storage"
)

func main() {
	// Загрузка .env файла
	_ = godotenv.Load()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализация логгера
	log := logger.New(cfg.LogLevel)

	// Подключение к PostgreSQL
	pg, err := storage.NewPostgres(&cfg.DB, log)
	if err != nil {
		fmt.Printf("❌ Failed to connect to PostgreSQL: %v\n", err)
		fmt.Printf("   Make sure PostgreSQL is running on %s:%d\n", cfg.DB.Host, cfg.DB.Port)
		os.Exit(1)
	}
	defer pg.Close()

	// Создание scraper storage
	scraperStorage := storage.NewScraperAdapter(pg)

	// Получаем список всех магазинов
	shops, err := scraperStorage.ListShops()
	if err != nil {
		fmt.Printf("❌ Failed to list shops: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=" + strings.Repeat("=", 100))
	fmt.Println("📊 ПРОВЕРКА КОНФИГУРАЦИИ МАГАЗИНОВ ДЛЯ DISCOVERY")
	fmt.Println("=" + strings.Repeat("=", 100))
	fmt.Println()

	activeCount := 0
	configuredCount := 0
	notConfiguredCount := 0

	for _, shop := range shops {
		status := "❌ Неактивен"
		if shop.Enabled {
			activeCount++
			status = "✅ Активен"

			catalogURL := shop.Selectors["catalog_url"]
			productLinkSelector := shop.Selectors["catalog_product_link"]
			nextPageSelector := shop.Selectors["catalog_next_page"]

			hasCatalogURL := catalogURL != ""
			hasProductLink := productLinkSelector != ""
			hasNextPage := nextPageSelector != ""

			if hasCatalogURL {
				configuredCount++
				status += " | ✅ Настроен для discovery"
			} else {
				notConfiguredCount++
				status += " | ⚠️ Нет catalog_url"
			}

			fmt.Printf("🏪 %s\n", shop.Name)
			fmt.Printf("   ID: %s\n", shop.ID)
			fmt.Printf("   Base URL: %s\n", shop.BaseURL)
			fmt.Printf("   Статус: %s\n", status)
			
			if hasCatalogURL {
				fmt.Printf("   📍 Catalog URL: %s\n", catalogURL)
			} else {
				fmt.Printf("   📍 Catalog URL: ❌ Не указан\n")
			}

			if hasProductLink {
				fmt.Printf("   🔗 Product Link Selector: %s\n", productLinkSelector)
			} else {
				fmt.Printf("   🔗 Product Link Selector: ⚠️ Не указан (будет использован дефолтный)\n")
			}

			if hasNextPage {
				fmt.Printf("   📄 Next Page Selector: %s\n", nextPageSelector)
			} else {
				fmt.Printf("   📄 Next Page Selector: ⚠️ Не указан (будет использован дефолтный)\n")
			}

			// Показываем все селекторы в JSON формате
			if len(shop.Selectors) > 0 {
				selectorsJSON, _ := json.MarshalIndent(shop.Selectors, "   ", "  ")
				fmt.Printf("   📋 Все селекторы:\n%s\n", string(selectorsJSON))
			}

			fmt.Println()
		} else {
			fmt.Printf("🏪 %s - %s\n", shop.Name, status)
			fmt.Println()
		}
	}

	fmt.Println("=" + strings.Repeat("=", 100))
	fmt.Println("📈 СТАТИСТИКА:")
	fmt.Printf("   Всего магазинов: %d\n", len(shops))
	fmt.Printf("   Активных: %d\n", activeCount)
	fmt.Printf("   Настроенных для discovery: %d\n", configuredCount)
	fmt.Printf("   Требуют настройки: %d\n", notConfiguredCount)
	fmt.Println("=" + strings.Repeat("=", 100))

	if notConfiguredCount > 0 {
		fmt.Println()
		fmt.Println("⚠️ ВНИМАНИЕ: Некоторые активные магазины не имеют catalog_url!")
		fmt.Println("   Для настройки используйте SQL скрипт: backend/scripts/check_shop_catalog_config.sql")
		fmt.Println("   Или обновите селекторы через API/базу данных.")
	}
}

