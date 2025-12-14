package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	l := logger.New(cfg.LogLevel)

	down := flag.Bool("down", false, "Rollback migrations")
	flag.Parse()

	migrationsPath := "file://migrations"

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Database,
	)

	l.Info("Starting migration...", map[string]interface{}{
		"source": migrationsPath,
		"db":     cfg.DB.Database,
		"host":   cfg.DB.Host,
		"port":   cfg.DB.Port,
	})

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		l.Fatal("Failed to create migrate instance", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if *down {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			l.Fatal("Migration down failed", map[string]interface{}{"error": err.Error()})
		}
		l.Info("Migration down finished successfully", nil)
	} else {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			l.Fatal("Migration up failed", map[string]interface{}{"error": err.Error()})
		}
		l.Info("Migration up finished successfully", nil)
	}
}
