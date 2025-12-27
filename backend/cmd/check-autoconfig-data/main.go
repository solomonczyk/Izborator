package main

import (
	"context"
	"fmt"

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
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Инициализация логгера
	log := logger.New(cfg.LogLevel)

	// Подключение к БД
	pg, err := storage.NewPostgres(&cfg.DB, log)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL", map[string]interface{}{"error": err.Error()})
	}
	defer pg.Close()

	ctx := context.Background()

	fmt.Println("🔍 Проверка данных для тестирования AutoConfig")
	fmt.Println("==============================================")
	fmt.Println()

	// 1. Статистика по потенциальным магазинам
	fmt.Println("1️⃣ Статистика по potential_shops:")
	query1 := `
		SELECT 
			status,
			COUNT(*) as count,
			COUNT(*) FILTER (WHERE metadata->>'site_type' = 'service_provider') as service_providers,
			COUNT(*) FILTER (WHERE metadata->>'site_type' = 'ecommerce') as ecommerce,
			MAX(confidence_score) as max_score,
			AVG(confidence_score) as avg_score
		FROM potential_shops
		GROUP BY status
		ORDER BY status;
	`

	rows, err := pg.DB().Query(ctx, query1)
	if err != nil {
		log.Fatal("Failed to query potential_shops", map[string]interface{}{"error": err.Error()})
	}
	defer rows.Close()

	fmt.Println("Status | Count | Service Providers | E-commerce | Max Score | Avg Score")
	fmt.Println("------|-------|-------------------|------------|-----------|----------")
	for rows.Next() {
		var status string
		var count, serviceProviders, ecommerce int
		var maxScore, avgScore *float64
		if err := rows.Scan(&status, &count, &serviceProviders, &ecommerce, &maxScore, &avgScore); err != nil {
			log.Error("Failed to scan row", map[string]interface{}{"error": err.Error()})
			continue
		}
		maxScoreStr := "N/A"
		avgScoreStr := "N/A"
		if maxScore != nil {
			maxScoreStr = fmt.Sprintf("%.2f", *maxScore)
		}
		if avgScore != nil {
			avgScoreStr = fmt.Sprintf("%.2f", *avgScore)
		}
		fmt.Printf("%s | %d | %d | %d | %s | %s\n", status, count, serviceProviders, ecommerce, maxScoreStr, avgScoreStr)
	}
	fmt.Println()

	// 2. Классифицированные кандидаты
	fmt.Println("2️⃣ Классифицированные кандидаты (готовы для AutoConfig):")
	query2 := `
		SELECT COUNT(*) 
		FROM potential_shops 
		WHERE status = 'classified';
	`

	var classifiedCount int
	err = pg.DB().QueryRow(ctx, query2).Scan(&classifiedCount)
	if err != nil {
		log.Fatal("Failed to query classified count", map[string]interface{}{"error": err.Error()})
	}

	if classifiedCount > 0 {
		fmt.Printf("✅ Найдено %d классифицированных кандидатов\n\n", classifiedCount)

		// Service providers
		query3 := `
			SELECT COUNT(*) 
			FROM potential_shops 
			WHERE status = 'classified' 
			AND metadata->>'site_type' = 'service_provider';
		`

		var serviceProviderCount int
		err = pg.DB().QueryRow(ctx, query3).Scan(&serviceProviderCount)
		if err != nil {
			log.Error("Failed to query service_provider count", map[string]interface{}{"error": err.Error()})
		} else {
			if serviceProviderCount > 0 {
				fmt.Printf("✅ Из них service_provider: %d (отлично для тестирования таблиц!)\n\n", serviceProviderCount)
			} else {
				fmt.Println("⚠️  Service providers не найдены. Нужно запустить Discovery для поиска услуг.")
			}
		}

		// Примеры кандидатов
		fmt.Println("   Примеры кандидатов:")
		query4 := `
			SELECT 
				domain,
				status,
				confidence_score,
				metadata->>'site_type' as site_type,
				discovered_at
			FROM potential_shops 
			WHERE status = 'classified'
			ORDER BY confidence_score DESC, discovered_at DESC
			LIMIT 5;
		`

		rows, err = pg.DB().Query(ctx, query4)
		if err == nil {
			defer rows.Close()
			fmt.Println("Domain | Status | Score | Site Type | Discovered")
			fmt.Println("-------|--------|-------|-----------|-----------")
			for rows.Next() {
				var domain, status, siteType, discoveredAt string
				var score float64
				if err := rows.Scan(&domain, &status, &score, &siteType, &discoveredAt); err == nil {
					if siteType == "" {
						siteType = "N/A"
					}
					fmt.Printf("%s | %s | %.2f | %s | %s\n", domain, status, score, siteType, discoveredAt)
				}
			}
		}
	} else {
		fmt.Println("❌ Классифицированных кандидатов нет (0)")
		fmt.Println("   Нужно запустить:")
		fmt.Println("   1. Discovery (поиск кандидатов)")
		fmt.Println("   2. Classifier (классификация)")
	}
	fmt.Println()

	// 3. Уже созданные магазины через AutoConfig
	fmt.Println("3️⃣ Магазины, созданные через AutoConfig:")
	query5 := `
		SELECT COUNT(*) 
		FROM shops 
		WHERE is_auto_configured = true;
	`

	var autoconfigCount int
	err = pg.DB().QueryRow(ctx, query5).Scan(&autoconfigCount)
	if err != nil {
		log.Error("Failed to query autoconfig count", map[string]interface{}{"error": err.Error()})
	} else {
		if autoconfigCount > 0 {
			fmt.Printf("✅ Найдено %d автоматически созданных магазинов\n\n", autoconfigCount)

			// Последние созданные
			fmt.Println("   Последние созданные:")
			query6 := `
				SELECT 
					name,
					base_url,
					is_active,
					ai_config_model,
					selectors->>'name' as name_selector,
					selectors->>'price' as price_selector,
					created_at
				FROM shops 
				WHERE is_auto_configured = true 
				ORDER BY created_at DESC 
				LIMIT 5;
			`

			rows, err = pg.DB().Query(ctx, query6)
			if err == nil {
				defer rows.Close()
				fmt.Println("Name | URL | Active | Model | Name Selector | Price Selector | Created")
				fmt.Println("-----|-----|--------|-------|--------------|----------------|--------")
				for rows.Next() {
					var name, baseURL, model, nameSel, priceSel, createdAt string
					var isActive bool
					if err := rows.Scan(&name, &baseURL, &isActive, &model, &nameSel, &priceSel, &createdAt); err == nil {
						activeStr := "No"
						if isActive {
							activeStr = "Yes"
						}
						if nameSel == "" {
							nameSel = "N/A"
						}
						if priceSel == "" {
							priceSel = "N/A"
						}
						fmt.Printf("%s | %s | %s | %s | %s | %s | %s\n", name, baseURL, activeStr, model, nameSel, priceSel, createdAt)
					}
				}
			}
		} else {
			fmt.Println("⚠️  Автоматически созданных магазинов нет")
		}
	}
	fmt.Println()

	// 4. Рекомендации
	fmt.Println("4️⃣ Рекомендации для тестирования:")
	fmt.Println()

	if classifiedCount == 0 {
		fmt.Println("📋 Для получения данных выполните:")
		fmt.Println()
		fmt.Println("   1. Запустить Discovery для поиска сайтов услуг:")
		fmt.Println("      ./backend/discovery -max-results 200")
		fmt.Println()
		fmt.Println("   2. Запустить Classifier для классификации:")
		fmt.Println("      ./backend/classifier -classify-all -limit 50")
		fmt.Println()
		fmt.Println("   3. Запустить AutoConfig для тестирования:")
		fmt.Println("      ./backend/autoconfig -limit 5")
	} else {
		var serviceProviderCount int
		query3 := `
			SELECT COUNT(*) 
			FROM potential_shops 
			WHERE status = 'classified' 
			AND metadata->>'site_type' = 'service_provider';
		`
		err = pg.DB().QueryRow(ctx, query3).Scan(&serviceProviderCount)
		if err == nil && serviceProviderCount > 0 {
			fmt.Println("✅ Отлично! Есть данные для тестирования табличных данных")
			fmt.Println()
			fmt.Println("   Запустите AutoConfig:")
			fmt.Println("   ./backend/autoconfig -limit 3")
		} else {
			fmt.Println("⚠️  Есть классифицированные кандидаты, но нет service_provider")
			fmt.Println()
			fmt.Println("   Для получения service_provider запустите Discovery с запросами для услуг:")
			fmt.Println("   ./backend/discovery -max-results 200")
		}
	}

	fmt.Println()
	fmt.Println("✅ Проверка завершена!")
}

