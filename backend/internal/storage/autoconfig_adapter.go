package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/autoconfig"
)

// autoconfigAdapter реализация Storage для autoconfig
type autoconfigAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewAutoconfigAdapter создаёт новый адаптер для autoconfig
func NewAutoconfigAdapter(pg *Postgres) autoconfig.Storage {
	return &autoconfigAdapter{
		pg:  pg,
		ctx: pg.Context(),
	}
}

// GetClassifiedCandidates получает кандидатов со статусом "classified" для авто-конфигурации
func (a *autoconfigAdapter) GetClassifiedCandidates(limit int) ([]autoconfig.Candidate, error) {
	query := `
		SELECT id, domain
		FROM potential_shops
		WHERE status = 'classified'
		ORDER BY confidence_score DESC, discovered_at ASC
		LIMIT $1
	`

	rows, err := a.pg.DB().Query(a.ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query candidates: %w", err)
	}
	defer rows.Close()

	var candidates []autoconfig.Candidate
	for rows.Next() {
		var c autoconfig.Candidate
		if err := rows.Scan(&c.ID, &c.Domain); err != nil {
			return nil, fmt.Errorf("failed to scan candidate: %w", err)
		}
		candidates = append(candidates, c)
	}

	return candidates, rows.Err()
}

// MarkAsConfigured создает магазин в таблице shops и обновляет статус в potential_shops
func (a *autoconfigAdapter) MarkAsConfigured(id string, config autoconfig.ShopConfig) error {
	// Начинаем транзакцию
	tx, err := a.pg.DB().Begin(a.ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(a.ctx)

	// Получаем данные кандидата
	var domain string
	var metadataJSON []byte
	err = tx.QueryRow(a.ctx, `
		SELECT domain, metadata
		FROM potential_shops
		WHERE id = $1
	`, id).Scan(&domain, &metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to get candidate: %w", err)
	}

	// Извлекаем название магазина из метаданных или используем домен
	shopName := domain
	if metadataJSON != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
			if title, ok := metadata["title"].(string); ok && title != "" {
				shopName = title
			}
		}
	}

	// Генерируем ID для магазина (используем UUID)
	shopID := uuid.New().String()

	// Генерируем code из name (как в миграции 0002)
	// Удаляем все символы кроме букв и цифр, приводим к нижнему регистру
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	shopCode := strings.ToLower(re.ReplaceAllString(shopName, ""))
	// Если code пустой (только спецсимволы в названии), используем домен
	if shopCode == "" {
		shopCode = strings.ToLower(re.ReplaceAllString(domain, ""))
		// Удаляем точку из домена для code
		shopCode = strings.ReplaceAll(shopCode, ".", "")
	}

	// Формируем base_url
	baseURL := domain
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	// Сериализуем селекторы
	selectorsJSON, err := json.Marshal(config.Selectors)
	if err != nil {
		return fmt.Errorf("failed to marshal selectors: %w", err)
	}

	// Создаем магазин в таблице shops
	_, err = tx.Exec(a.ctx, `
		INSERT INTO shops (id, name, code, base_url, selectors, rate_limit, is_active, is_auto_configured, ai_config_model, discovery_source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`, shopID, shopName, shopCode, baseURL, selectorsJSON, 1, true, true, "gpt-4o-mini", "google_search")
	if err != nil {
		return fmt.Errorf("failed to insert shop: %w", err)
	}

	// Обновляем статус в potential_shops
	_, err = tx.Exec(a.ctx, `
		UPDATE potential_shops
		SET status = 'configured',
		    updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to update potential_shop status: %w", err)
	}

	// Сохраняем попытку конфигурации в shop_config_attempts
	_, err = tx.Exec(a.ctx, `
		INSERT INTO shop_config_attempts (potential_shop_id, shop_id, ai_response, validation_result, status, created_at)
		VALUES ($1, $2, $3, $4, 'success', NOW())
	`, id, shopID, selectorsJSON, json.RawMessage(`{"validated": true}`))
	if err != nil {
		// Не критично, продолжаем
	}

	// Коммитим транзакцию
	if err := tx.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MarkAsFailed обновляет статус кандидата на "rejected" и сохраняет причину ошибки
func (a *autoconfigAdapter) MarkAsFailed(id string, reason string) error {
	// Обновляем статус и метаданные с ошибкой
	query := `
		UPDATE potential_shops
		SET status = 'rejected',
		    metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('autoconfig_error', $2, 'autoconfig_failed_at', $3),
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := a.pg.DB().Exec(a.ctx, query, id, reason, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to mark as failed: %w", err)
	}

	// Сохраняем попытку конфигурации в shop_config_attempts
	_, err = a.pg.DB().Exec(a.ctx, `
		INSERT INTO shop_config_attempts (potential_shop_id, status, error_message, created_at)
		VALUES ($1, 'failed', $2, NOW())
	`, id, reason)
	if err != nil {
		// Не критично, продолжаем
	}

	return nil
}

