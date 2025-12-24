package storage

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

// SetupTestDB создаёт подключение к тестовой БД
// Использует переменные окружения или значения по умолчанию
func SetupTestDB(t *testing.T) *Postgres {
	t.Helper()

	// Загружаем конфиг из переменных окружения
	cfg := &config.DBConfig{
		Host:            getEnv("TEST_DB_HOST", "localhost"),
		Port:            getEnvAsInt("TEST_DB_PORT", 5433),
		User:            getEnv("TEST_DB_USER", "postgres"),
		Password:        getEnv("TEST_DB_PASSWORD", "postgres"),
		Database:        getEnv("TEST_DB_NAME", "izborator_test"),
		MaxConnections:  5,
		MinConnections:  1,
		MaxIdleTime:     30 * time.Minute,
		ConnMaxLifetime: 1 * time.Hour,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := logger.New("error") // Минимальное логирование для тестов

	pg, err := NewPostgres(cfg, log)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Проверяем подключение
	if err := pg.DB().Ping(ctx); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return pg
}

// CleanupTestData очищает тестовые данные после теста
func CleanupTestData(t *testing.T, pg *Postgres, tables []string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Отключаем foreign key checks для быстрой очистки
	_, err := pg.DB().Exec(ctx, "SET session_replication_role = 'replica'")
	if err != nil {
		t.Logf("Warning: failed to disable FK checks: %v", err)
	}

	for _, table := range tables {
		_, err := pg.DB().Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE")
		if err != nil {
			t.Logf("Warning: failed to truncate %s: %v", table, err)
		}
	}

	// Включаем обратно
	_, err = pg.DB().Exec(ctx, "SET session_replication_role = 'origin'")
	if err != nil {
		t.Logf("Warning: failed to enable FK checks: %v", err)
	}
}

// Helper функции для работы с переменными окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

