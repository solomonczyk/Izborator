package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
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

	// Если флагов нет, показываем usage
	fmt.Println("Usage:")
	fmt.Println("  -domain <domain>     Test single domain")
	fmt.Println("  -test-list           Test on predefined list")
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

