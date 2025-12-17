package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/app"
	"github.com/solomonczyk/izborator/internal/autoconfig"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

func main() {
	// Загрузка .env файла
	_ = godotenv.Load()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Инициализация приложения
	application, err := app.NewWorkerApp(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to init app: %v", err))
	}
	defer application.Close()

	log := application.Logger()
	autoconfigService := application.GetAutoconfigService()

	if autoconfigService == nil {
		log.Fatal("Autoconfig service is not available. Check OPENAI_API_KEY in .env", nil)
	}

	ctx := context.Background()

	// Флаги
	limit := flag.Int("limit", 1, "Number of candidates to process (default: 1)")
	daemon := flag.Bool("daemon", false, "Run in daemon mode (process candidates continuously)")
	interval := flag.Duration("interval", 5*time.Minute, "Interval between processing batches in daemon mode")
	flag.Parse()

	if *daemon {
		// Режим демона - обрабатываем кандидатов непрерывно
		log.Info("🤖 Starting Autoconfig daemon", map[string]interface{}{
			"interval": interval.String(),
		})

		for {
			processed := processCandidates(ctx, autoconfigService, *limit, log)
			
			if processed == 0 {
				log.Info("No candidates to process, waiting...", map[string]interface{}{
					"interval": interval.String(),
				})
			}

			// Ждем перед следующей итерацией
			time.Sleep(*interval)
		}
	} else {
		// Одноразовая обработка
		processCandidates(ctx, autoconfigService, *limit, log)
	}
}

// processCandidates обрабатывает кандидатов
func processCandidates(ctx context.Context, service *autoconfig.Service, limit int, log *logger.Logger) int {
	processed := 0
	successful := 0
	failed := 0

	log.Info("🔍 Processing candidates", map[string]interface{}{
		"limit": limit,
	})

	for i := 0; i < limit; i++ {
		err := service.ProcessNextCandidate(ctx)
		if err != nil {
			// Проверяем, это ошибка "нет работы" или реальная ошибка
			if err.Error() == "no candidates available" || err.Error() == "no work" {
				// Нет кандидатов для обработки
				break
			}
			
			failed++
			log.Error("Failed to process candidate", map[string]interface{}{
				"error":  err.Error(),
				"number": i + 1,
			})
			// Продолжаем обработку следующих кандидатов
			continue
		}

		// Успешно обработан
		successful++
		processed++

		log.Info("✅ Candidate processed successfully", map[string]interface{}{
			"number": i + 1,
		})

		// Небольшая задержка между обработкой кандидатов
		if i < limit-1 {
			time.Sleep(2 * time.Second)
		}
	}

	log.Info("📊 Processing summary", map[string]interface{}{
		"processed": processed,
		"successful": successful,
		"failed": failed,
	})

	return processed
}

