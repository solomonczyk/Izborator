package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/app"
	"github.com/solomonczyk/izborator/internal/classifier"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.LogLevel)

	ctx := context.Background()

	// Флаги
	testDomain := flag.String("domain", "", "Test single domain")
	testList := flag.Bool("test-list", false, "Test on predefined list of shops and non-shops")
	classifyAll := flag.Bool("classify-all", false, "Classify all domains with status 'new' from database")
	limit := flag.Int("limit", 0, "Limit number of domains to classify (0 = no limit)")
	flag.Parse()

	// Для тестирования создаем классификатор напрямую без БД
	// Storage не нужен для классификации - только для сохранения результатов
	classifierService := createTestClassifier(log)

	if *testDomain != "" {
		// Тест одного домена
		testSingleDomain(ctx, classifierService, *testDomain, log)
		return
	}

	if *testList {
		// Тест на предопределенном списке
		testPredefinedList(ctx, classifierService, log)
		return
	}

	if *classifyAll {
		// Классификация всех доменов из БД
		classifyAllDomains(ctx, *limit, log)
		return
	}

	// Если флагов нет, показываем usage
	fmt.Println("Usage:")
	fmt.Println("  -domain <domain>     Test single domain")
	fmt.Println("  -test-list           Test on predefined list")
	fmt.Println("  -classify-all        Classify all domains with status 'new' from database")
	fmt.Println("  -limit <number>      Limit number of domains to classify (use with -classify-all)")
}

// createTestClassifier создает классификатор для тестирования без БД
func createTestClassifier(log *logger.Logger) *classifier.Service {
	// Создаем nil storage - для тестирования он не нужен
	// Классификатор может работать без Storage (он только анализирует HTML)
	return classifier.New(nil, log)
}

func testSingleDomain(ctx context.Context, classifierService *classifier.Service, domain string, log *logger.Logger) {
	log.Info("Testing domain", map[string]interface{}{"domain": domain})

	result, err := classifierService.Classify(ctx, domain)
	if err != nil {
		log.Error("Classification failed", map[string]interface{}{"error": err.Error()})
		return
	}

	fmt.Println("\n=== Classification Result ===")
	fmt.Printf("Domain: %s\n", domain)
	fmt.Printf("Is Shop: %v\n", result.IsShop)
	fmt.Printf("Total Score: %.2f\n", result.Score.TotalScore)
	fmt.Printf("  - Keywords Score: %.2f\n", result.Score.KeywordsScore)
	fmt.Printf("  - Platform Score: %.2f\n", result.Score.PlatformScore)
	fmt.Printf("  - Structure Score: %.2f\n", result.Score.StructureScore)
	if result.DetectedPlatform != "" {
		fmt.Printf("Detected Platform: %s\n", result.DetectedPlatform)
	}
	fmt.Println("\nReasons:")
	for i, reason := range result.Reasons {
		fmt.Printf("  %d. %s\n", i+1, reason)
	}
	fmt.Println()
}

func testPredefinedList(ctx context.Context, classifierService *classifier.Service, log *logger.Logger) {
	// Список известных магазинов (должны быть определены как магазины)
	shops := []string{
		"gigatron.rs",
		"tehnomanija.rs",
		"winwin.rs",
		"ananas.rs",
		"emmi.rs",
	}

	// Список известных не-магазинов (не должны быть определены как магазины)
	nonShops := []string{
		"b92.net",
		"politika.rs",
		"rts.rs",
		"blic.rs",
		"novosti.rs",
	}

	fmt.Println("\n=== Testing on Predefined List ===\n")

	// Тестируем магазины
	fmt.Println("--- SHOPS (should be classified as shops) ---")
	shopCorrect := 0
	for _, domain := range shops {
		result, err := classifierService.Classify(ctx, domain)
		if err != nil {
			fmt.Printf("❌ %s: ERROR - %v\n", domain, err)
			continue
		}

		if result.IsShop {
			fmt.Printf("✅ %s: CORRECT (score: %.2f)\n", domain, result.Score.TotalScore)
			shopCorrect++
		} else {
			fmt.Printf("❌ %s: WRONG - classified as non-shop (score: %.2f)\n", domain, result.Score.TotalScore)
		}
	}

	// Тестируем не-магазины
	fmt.Println("\n--- NON-SHOPS (should NOT be classified as shops) ---")
	nonShopCorrect := 0
	for _, domain := range nonShops {
		result, err := classifierService.Classify(ctx, domain)
		if err != nil {
			fmt.Printf("❌ %s: ERROR - %v\n", domain, err)
			continue
		}

		if !result.IsShop {
			fmt.Printf("✅ %s: CORRECT (score: %.2f)\n", domain, result.Score.TotalScore)
			nonShopCorrect++
		} else {
			fmt.Printf("❌ %s: WRONG - classified as shop (score: %.2f)\n", domain, result.Score.TotalScore)
		}
	}

	// Итоговая статистика
	total := len(shops) + len(nonShops)
	correct := shopCorrect + nonShopCorrect
	accuracy := float64(correct) / float64(total) * 100

	fmt.Println("\n=== Results ===")
	fmt.Printf("Shops correctly identified: %d/%d\n", shopCorrect, len(shops))
	fmt.Printf("Non-shops correctly identified: %d/%d\n", nonShopCorrect, len(nonShops))
	fmt.Printf("Total accuracy: %.1f%% (%d/%d)\n", accuracy, correct, total)

	if accuracy >= 85.0 {
		fmt.Println("\n✅ SUCCESS: Accuracy >= 85%")
		os.Exit(0)
	} else {
		fmt.Println("\n❌ FAILED: Accuracy < 85%")
		os.Exit(1)
	}
}

// classifyAllDomains классифицирует все домены со статусом "new" из БД
func classifyAllDomains(ctx context.Context, limit int, log *logger.Logger) {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", map[string]interface{}{"error": err.Error()})
	}

	// Инициализация приложения
	application, err := app.NewWorkerApp(cfg)
	if err != nil {
		log.Fatal("Failed to init app", map[string]interface{}{"error": err.Error()})
	}
	defer application.Close()

	storage := application.GetClassifierStorage()
	classifierService := application.Classifier

	// Получаем все домены со статусом "new"
	if limit == 0 {
		limit = 1000 // Большое число, чтобы получить все
	}

	log.Info("Fetching domains to classify", map[string]interface{}{
		"status": "new",
		"limit":  limit,
	})

	domains, err := storage.ListPotentialShopsByStatus("new", limit)
	if err != nil {
		log.Fatal("Failed to fetch domains", map[string]interface{}{"error": err.Error()})
	}

	if len(domains) == 0 {
		log.Info("No domains to classify", nil)
		return
	}

	log.Info("Starting classification", map[string]interface{}{
		"total": len(domains),
	})

	classified := 0
	rejected := 0
	pendingReview := 0
	errors := 0

	for i, shop := range domains {
		log.Info("Classifying domain", map[string]interface{}{
			"domain": shop.Domain,
			"number": i + 1,
			"total":  len(domains),
		})

		result, err := classifierService.Classify(ctx, shop.Domain)
		if err != nil {
			log.Error("Classification failed", map[string]interface{}{
				"domain": shop.Domain,
				"error":  err.Error(),
			})
			errors++
			continue
		}

		// Обновляем статус и confidence score
		shop.ConfidenceScore = result.Score.TotalScore

		// Определяем статус на основе результата
		if result.IsShop {
			shop.Status = "classified"
			classified++
		} else if result.Score.TotalScore >= 0.5 {
			shop.Status = "pending_review"
			pendingReview++
		} else {
			shop.Status = "rejected"
			rejected++
		}

		// Обновляем метаданные с результатами классификации
		if shop.Metadata == nil {
			shop.Metadata = make(map[string]interface{})
		}
		shop.Metadata["classification"] = map[string]interface{}{
			"keywords_score":  result.Score.KeywordsScore,
			"platform_score":  result.Score.PlatformScore,
			"structure_score": result.Score.StructureScore,
			"total_score":      result.Score.TotalScore,
			"detected_platform": result.DetectedPlatform,
			"reasons":          result.Reasons,
			"classified_at":    time.Now().Format(time.RFC3339),
		}

		// Сохраняем обновленный результат
		if err := storage.UpdatePotentialShop(shop); err != nil {
			log.Error("Failed to update shop", map[string]interface{}{
				"domain": shop.Domain,
				"id":     shop.ID,
				"status": shop.Status,
				"error":  err.Error(),
			})
			fmt.Printf("❌ ERROR updating %s: %v\n", shop.Domain, err)
			errors++
			continue
		}

		statusIcon := "✅"
		if shop.Status == "rejected" {
			statusIcon = "❌"
		} else if shop.Status == "pending_review" {
			statusIcon = "⚠️"
		}

		log.Info("Domain classified", map[string]interface{}{
			"domain":  shop.Domain,
			"status":  shop.Status,
			"score":   result.Score.TotalScore,
			"platform": result.DetectedPlatform,
		})

		fmt.Printf("%s [%d/%d] %s -> %s (score: %.2f)\n",
			statusIcon, i+1, len(domains), shop.Domain, shop.Status, result.Score.TotalScore)

		// Небольшая задержка, чтобы не перегружать серверы
		time.Sleep(500 * time.Millisecond)
	}

	// Итоговая статистика
	fmt.Println("\n=== Classification Summary ===")
	fmt.Printf("Total processed: %d\n", len(domains))
	fmt.Printf("✅ Classified (shops): %d\n", classified)
	fmt.Printf("⚠️  Pending review: %d\n", pendingReview)
	fmt.Printf("❌ Rejected: %d\n", rejected)
	fmt.Printf("⚠️  Errors: %d\n", errors)
	fmt.Printf("\nSuccess rate: %.1f%%\n", float64(classified+pendingReview)/float64(len(domains))*100)
}

