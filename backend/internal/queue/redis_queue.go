package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/solomonczyk/izborator/internal/logger"
)

const (
	defaultBlockTimeout = 5 * time.Second
	publishTimeout      = 5 * time.Second
)

// RedisQueue implements a simple Redis list-backed queue.
type RedisQueue struct {
	client    *redis.Client
	log       *logger.Logger
	keyPrefix string
}

// NewRedisQueue returns a Redis-backed queue client.
func NewRedisQueue(client *redis.Client, log *logger.Logger) *RedisQueue {
	if log == nil {
		log = logger.New("info")
	}
	return &RedisQueue{
		client:    client,
		log:       log,
		keyPrefix: "queue:",
	}
}

// Publish pushes a message to the queue.
func (q *RedisQueue) Publish(topic string, data interface{}) error {
	if topic == "" {
		return errors.New("topic is required")
	}
	if q.client == nil {
		return errors.New("redis client is nil")
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal queue payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()

	return q.client.RPush(ctx, q.key(topic), payload).Err()
}

// Consume reads messages in a loop and passes them to handler.
func (q *RedisQueue) Consume(ctx context.Context, topic string, handler func([]byte) error) error {
	if topic == "" {
		return errors.New("topic is required")
	}
	if handler == nil {
		return errors.New("handler is required")
	}
	if q.client == nil {
		return errors.New("redis client is nil")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		payload, err := q.pop(ctx, topic, defaultBlockTimeout)
		if err != nil {
			q.log.Warn("queue consume failed", map[string]interface{}{
				"topic": topic,
				"error": err.Error(),
			})
			continue
		}
		if len(payload) == 0 {
			continue
		}
		if err := handler(payload); err != nil {
			q.log.Error("queue handler failed", map[string]interface{}{
				"topic": topic,
				"error": err.Error(),
			})
		}
	}
}

func (q *RedisQueue) pop(ctx context.Context, topic string, timeout time.Duration) ([]byte, error) {
	if timeout <= 0 {
		timeout = defaultBlockTimeout
	}
	result, err := q.client.BRPop(ctx, timeout, q.key(topic)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if len(result) != 2 {
		return nil, fmt.Errorf("unexpected redis response for queue pop")
	}
	return []byte(result[1]), nil
}

func (q *RedisQueue) key(topic string) string {
	return q.keyPrefix + topic
}
