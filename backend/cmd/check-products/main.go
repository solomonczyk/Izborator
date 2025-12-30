package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/app"
	"github.com/solomonczyk/izborator/internal/config"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	application, err := app.NewAPIApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer application.Close()

	ctx := context.Background()
	logger := application.Logger()

	// Проверка товаров в PostgreSQL
	db := application.Postgres().DB()
	var count int64
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		logger.Error("Failed to count products", map[string]interface{}{
			"error": err,
		})
		fmt.Printf("❌ PostgreSQL error: %v\n", err)
	} else {
		fmt.Printf("📊 Total products in PostgreSQL: %d\n", count)
	}

	// Проверка магазинов
	var shopCount int64
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM shops").Scan(&shopCount)
	if err == nil {
		fmt.Printf("📊 Total shops in PostgreSQL: %d\n", shopCount)
	}

	// Попробуем поиск через сервис
	fmt.Println("\n🔍 Searching for 'айфон'...")
	results, err := application.ProductsService.Search(ctx, "айфон")
	if err != nil {
		logger.Error("Search failed", map[string]interface{}{
			"error": err,
		})
		fmt.Printf("❌ Search error: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d products\n", len(results))
		for i, p := range results {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s\n", i+1, p.Name)
		}
	}

	// Покажем первые товары в БД
	fmt.Println("\n📋 Sample products from database:")
	rows, err := db.Query(ctx, "SELECT id, name FROM products LIMIT 5")
	if err == nil {
		defer rows.Close()
		count := 0
		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				continue
			}
			fmt.Printf("  %d. %s\n", count+1, name)
			count++
		}
		if count == 0 {
			fmt.Println("  (no products found)")
		}
	}
}
