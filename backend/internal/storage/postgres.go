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
func New(ctx context.Context, connString string, log *logger.Logger) (*Repository, error) {
	// Конфигурация пула
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Настройки тайм-аутов и макс. соединений (базовые для highload)
	config.MaxConns = 50
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	// Создаем пул
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем, что база жива (Ping)
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Successfully connected to PostgreSQL", map[string]interface{}{})

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
	ctx := context.Background()
	
	dsn := cfg.DSN()
	log.Debug("Creating PostgreSQL connection", map[string]interface{}{
		"dsn": dsn,
	})
	
	repo, err := New(ctx, dsn, log)
	if err != nil {
		log.Error("Failed to create PostgreSQL connection", map[string]interface{}{
			"error": err.Error(),
			"dsn":   dsn,
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
