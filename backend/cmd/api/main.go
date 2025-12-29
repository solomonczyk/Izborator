package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/solomonczyk/izborator/internal/app"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/http/router"
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
	application, err := app.NewAPIApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer application.Close()

	// Инициализация роутера
	var redisClient *redis.Client
	if application.Redis() != nil {
		redisClient = application.Redis().Client()
	}
	r := router.New(application.Logger(), application.ProductsService, application.PriceHistoryService, application.ScrapingStatsService, application.CategoriesService, application.CitiesService, application.GetTranslator(), application.Postgres(), redisClient)

	// Настройка HTTP сервера
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в горутине
	go func() {
		application.Logger().Info("Starting API server", map[string]interface{}{"port": cfg.Server.Port})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			application.Logger().Fatal("Failed to start server", map[string]interface{}{"error": err})
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.Logger().Info("Shutting down server...", map[string]interface{}{})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		application.Logger().Error("Server forced to shutdown", map[string]interface{}{"error": err})
	}

	application.Logger().Info("Server exited", map[string]interface{}{})
}
