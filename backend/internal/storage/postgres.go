package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

// Repository — это главная структура, которая держит соединение с БД.
// В будущем мы добавим сюда методы (SaveProduct, GetPrices и т.д.),
// чтобы она удовлетворяла интерфейсам из domain-слоя.
type Repository struct {
	Pool *pgxpool.Pool
	Log  *logger.Logger
}

// New создает пул соединений с PostgreSQL
func New(ctx context.Context, cfg *config.DBConfig, log *logger.Logger) (*Repository, error) {
	connString := cfg.DSN()

	// Конфигурация пула
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Настройки тайм-аутов и макс. соединений из конфига
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MinConnections)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxIdleTime

	// Создаем пул
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем, что база жива (Ping)
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Устанавливаем кодировку UTF-8 для всех соединений в пуле
	_, err = pool.Exec(ctx, "SET client_encoding = 'UTF8'")
	if err != nil {
		return nil, fmt.Errorf("failed to set client encoding: %w", err)
	}

	log.Info("Successfully connected to PostgreSQL", map[string]interface{}{
		"max_connections": cfg.MaxConnections,
		"min_connections": cfg.MinConnections,
		"encoding":        "UTF8",
	})

	return &Repository{
		Pool: pool,
		Log:  log,
	}, nil
}

// Close закрывает соединение
func (r *Repository) Close() {
	if r.Pool != nil {
		r.Pool.Close()
	}
}

// Postgres — обёртка для обратной совместимости
// Использует Repository внутри
type Postgres struct {
	repo *Repository
	ctx  context.Context
}

// NewPostgres создаёт новый клиент PostgreSQL (обёртка для обратной совместимости)
func NewPostgres(cfg *config.DBConfig, log *logger.Logger) (*Postgres, error) {
	// Используем контекст с таймаутом для инициализации
	ctx := context.Background()
	connCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Debug("Creating PostgreSQL connection", map[string]interface{}{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Database,
	})

	repo, err := New(connCtx, cfg, log)
	if err != nil {
		log.Error("Failed to create PostgreSQL connection", map[string]interface{}{
			"error": err.Error(),
			"host":  cfg.Host,
			"port":  cfg.Port,
		})
		return nil, fmt.Errorf("failed to create postgres connection: %w", err)
	}

	return &Postgres{
		repo: repo,
		ctx:  ctx,
	}, nil
}

// Close закрывает соединение с базой данных
func (p *Postgres) Close() error {
	if p.repo != nil {
		p.repo.Close()
	}
	return nil
}

// DB возвращает *pgxpool.Pool для прямого доступа
func (p *Postgres) DB() *pgxpool.Pool {
	return p.repo.Pool
}

// Context возвращает context для использования в запросах
func (p *Postgres) Context() context.Context {
	return p.ctx
}

// Repository возвращает Repository для прямого доступа
func (p *Postgres) Repository() *Repository {
	return p.repo
}
