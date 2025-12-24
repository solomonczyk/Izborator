package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

// Redis клиент для работы с Redis
type Redis struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedis создаёт новый клиент Redis
func NewRedis(cfg *config.RedisConfig, log *logger.Logger) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Проверка подключения с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Redis connection established", map[string]interface{}{})

	return &Redis{
		client: client,
		logger: log,
	}, nil
}

// Close закрывает соединение с Redis
func (r *Redis) Close() error {
	return r.client.Close()
}

// Client возвращает *redis.Client для прямого доступа
func (r *Redis) Client() *redis.Client {
	return r.client
}
