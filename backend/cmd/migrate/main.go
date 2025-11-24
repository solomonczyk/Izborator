package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solomonczyk/izborator/internal/config"
)

func main() {
	var (
		up      = flag.Bool("up", false, "Apply all pending migrations")
		down    = flag.Int("down", 0, "Rollback N migrations")
		status  = flag.Bool("status", false, "Show migration status")
		version = flag.Bool("version", false, "Show current migration version")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	
	pool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Инициализация таблицы миграций
	if err := initMigrationsTable(ctx, pool); err != nil {
		log.Fatalf("Failed to init migrations table: %v", err)
	}

	// Определяем путь к миграциям относительно корня проекта
	// Если запускаем из cmd/migrate, нужно подняться на уровень выше
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Пробуем путь относительно cmd/migrate
		migrationsDir = "../migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			log.Fatalf("Migrations directory not found. Tried: migrations and ../migrations")
		}
	}

	migrations, err := loadMigrations(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	currentVersion, err := getCurrentVersion(ctx, pool)
	if err != nil {
		log.Fatalf("Failed to get current version: %v", err)
	}

	switch {
	case *version:
		fmt.Printf("Current migration version: %d\n", currentVersion)
	case *status:
		showStatus(migrations, currentVersion)
	case *down > 0:
		if err := rollbackMigrations(ctx, pool, migrations, currentVersion, *down); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		fmt.Printf("Rolled back %d migration(s)\n", *down)
	case *up:
		if err := applyMigrations(ctx, pool, migrations, currentVersion); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
}

// Migration представляет одну миграцию
type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

// initMigrationsTable создаёт таблицу для отслеживания миграций
func initMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := pool.Exec(ctx, query)
	return err
}

// loadMigrations загружает все миграции из директории
func loadMigrations(dir string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".up.sql") {
			return nil
		}

		// Парсим имя файла: 0001_name.up.sql
		baseName := filepath.Base(path)
		parts := strings.Split(baseName, "_")
		if len(parts) < 2 {
			return nil
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			return nil
		}

		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

		// Загружаем UP миграцию
		upSQL, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Загружаем DOWN миграцию
		downPath := strings.Replace(path, ".up.sql", ".down.sql", 1)
		downSQL, err := os.ReadFile(downPath)
		if err != nil {
			// DOWN миграция не обязательна
			downSQL = []byte("-- No rollback available")
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(upSQL),
			DownSQL: string(downSQL),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Сортируем по версии
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getCurrentVersion получает текущую версию миграции
func getCurrentVersion(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var version int
	err := pool.QueryRow(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}
	return version, nil
}

// applyMigrations применяет все неприменённые миграции
func applyMigrations(ctx context.Context, pool *pgxpool.Pool, migrations []Migration, currentVersion int) error {
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		fmt.Printf("Applying migration %d: %s...\n", migration.Version, migration.Name)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Применяем миграцию
		if _, err := tx.Exec(ctx, migration.UpSQL); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		// Записываем версию
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		fmt.Printf("Migration %d applied successfully\n", migration.Version)
	}

	return nil
}

// rollbackMigrations откатывает N последних миграций
func rollbackMigrations(ctx context.Context, pool *pgxpool.Pool, migrations []Migration, currentVersion int, count int) error {
	// Находим миграции для отката (в обратном порядке)
	var toRollback []Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if migrations[i].Version <= currentVersion && len(toRollback) < count {
			toRollback = append(toRollback, migrations[i])
		}
	}

	if len(toRollback) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	for _, migration := range toRollback {
		fmt.Printf("Rolling back migration %d: %s...\n", migration.Version, migration.Name)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Откатываем миграцию
		if _, err := tx.Exec(ctx, migration.DownSQL); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
		}

		// Удаляем запись о версии
		if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", migration.Version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to remove migration record %d: %w", migration.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit rollback %d: %w", migration.Version, err)
		}

		fmt.Printf("Migration %d rolled back successfully\n", migration.Version)
	}

	return nil
}

// showStatus показывает статус всех миграций
func showStatus(migrations []Migration, currentVersion int) {
	fmt.Println("Migration Status:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-10s %-30s %-10s\n", "Version", "Name", "Status")
	fmt.Println(strings.Repeat("-", 60))

	for _, migration := range migrations {
		status := "Pending"
		if migration.Version <= currentVersion {
			status = "Applied"
		}
		fmt.Printf("%-10d %-30s %-10s\n", migration.Version, migration.Name, status)
	}
}

