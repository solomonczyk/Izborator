package queue

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

// Client defines the queue operations used by the app.
type Client interface {
	Publish(topic string, data interface{}) error
	Consume(ctx context.Context, topic string, handler func([]byte) error) error
}

// New creates a queue client based on configuration.
func New(cfg *config.QueueConfig, redisClient *redis.Client, log *logger.Logger) (Client, error) {
	if cfg == nil {
		return nil, nil
	}

	queueType := strings.ToLower(strings.TrimSpace(cfg.Type))
	switch queueType {
	case "", "none", "disabled":
		return nil, nil
	case "redis":
		if redisClient == nil {
			return nil, fmt.Errorf("redis client is required for queue type redis")
		}
		return NewRedisQueue(redisClient, log), nil
	default:
		return nil, fmt.Errorf("unsupported queue type: %s", queueType)
	}
}
