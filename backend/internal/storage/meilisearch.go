package storage

import (
	"fmt"

	"github.com/meilisearch/meilisearch-go"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

// Meilisearch клиент для работы с Meilisearch
type Meilisearch struct {
	client *meilisearch.Client
	logger *logger.Logger
}

// NewMeilisearch создаёт новый клиент Meilisearch
func NewMeilisearch(cfg *config.MeilisearchConfig, log *logger.Logger) (*Meilisearch, error) {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   cfg.Address(),
		APIKey: cfg.APIKey,
	})

	// Проверка подключения
	health, err := client.Health()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Meilisearch: %w", err)
	}

	log.Info("Meilisearch connection established", map[string]interface{}{
		"status": health.Status,
	})

	return &Meilisearch{
		client: client,
		logger: log,
	}, nil
}

// Client возвращает *meilisearch.Client для прямого доступа
func (m *Meilisearch) Client() *meilisearch.Client {
	return m.client
}

// Close закрывает соединение с Meilisearch (заглушка, т.к. клиент не требует закрытия)
func (m *Meilisearch) Close() error {
	// Meilisearch клиент не требует явного закрытия
	return nil
}
