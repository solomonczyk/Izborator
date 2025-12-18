package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/classifier"
)

// classifierAdapter реализация Storage для classifier
type classifierAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewClassifierAdapter создаёт новый адаптер для classifier
func NewClassifierAdapter(pg *Postgres) classifier.Storage {
	return &classifierAdapter{
		pg:  pg,
		ctx: pg.Context(),
	}
}

// SavePotentialShop сохраняет кандидата на магазин
func (a *classifierAdapter) SavePotentialShop(shop *classifier.PotentialShop) error {
	query := `
		INSERT INTO potential_shops (id, domain, source, status, confidence_score, discovered_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (domain) DO UPDATE SET
			source = EXCLUDED.source,
			updated_at = NOW()
	`

	var metadataJSON []byte
	if shop.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(shop.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	var discoveredAt time.Time
	if shop.DiscoveredAt != "" {
		var err error
		discoveredAt, err = time.Parse(time.RFC3339, shop.DiscoveredAt)
		if err != nil {
			discoveredAt = time.Now()
		}
	} else {
		discoveredAt = time.Now()
	}

	_, err := a.pg.DB().Exec(a.ctx, query,
		shop.ID,
		shop.Domain,
		shop.Source,
		shop.Status,
		shop.ConfidenceScore,
		discoveredAt,
		metadataJSON,
	)

	return err
}

// GetPotentialShopByDomain получает кандидата по домену
func (a *classifierAdapter) GetPotentialShopByDomain(domain string) (*classifier.PotentialShop, error) {
	query := `
		SELECT id, domain, source, status, confidence_score, discovered_at, metadata, classified_at
		FROM potential_shops
		WHERE domain = $1
	`

	var shop classifier.PotentialShop
	var discoveredAt time.Time
	var classifiedAt *time.Time
	var metadataJSON []byte

	err := a.pg.DB().QueryRow(a.ctx, query, domain).Scan(
		&shop.ID,
		&shop.Domain,
		&shop.Source,
		&shop.Status,
		&shop.ConfidenceScore,
		&discoveredAt,
		&metadataJSON,
		&classifiedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if !discoveredAt.IsZero() {
		shop.DiscoveredAt = discoveredAt.Format(time.RFC3339)
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &shop.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &shop, nil
}

// ListPotentialShopsByStatus получает список кандидатов по статусу
func (a *classifierAdapter) ListPotentialShopsByStatus(status string, limit int) ([]*classifier.PotentialShop, error) {
	query := `
		SELECT id, domain, source, status, confidence_score, discovered_at, metadata, classified_at
		FROM potential_shops
		WHERE status = $1
		ORDER BY discovered_at DESC
		LIMIT $2
	`

	rows, err := a.pg.DB().Query(a.ctx, query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shops []*classifier.PotentialShop
	for rows.Next() {
		var shop classifier.PotentialShop
		var discoveredAt time.Time
		var classifiedAt *time.Time
		var metadataJSON []byte

		if err := rows.Scan(
			&shop.ID,
			&shop.Domain,
			&shop.Source,
			&shop.Status,
			&shop.ConfidenceScore,
			&discoveredAt,
			&metadataJSON,
			&classifiedAt,
		); err != nil {
			return nil, err
		}

		if !discoveredAt.IsZero() {
			shop.DiscoveredAt = discoveredAt.Format(time.RFC3339)
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &shop.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		shops = append(shops, &shop)
	}

	return shops, rows.Err()
}

// UpdatePotentialShop обновляет кандидата
func (a *classifierAdapter) UpdatePotentialShop(shop *classifier.PotentialShop) error {
	query := `
		UPDATE potential_shops
		SET status = $1,
		    confidence_score = $2,
		    classified_at = CASE WHEN $1 IN ('classified', 'configured') THEN COALESCE(classified_at, NOW()) ELSE classified_at END,
		    metadata = COALESCE($3::jsonb, metadata),
		    updated_at = NOW()
		WHERE id = $4
	`

	var metadataJSON []byte
	var err error
	if shop.Metadata != nil && len(shop.Metadata) > 0 {
		metadataJSON, err = json.Marshal(shop.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	result, err := a.pg.DB().Exec(a.ctx, query,
		shop.Status,
		shop.ConfidenceScore,
		metadataJSON,
		shop.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update potential_shop (id=%s, domain=%s): %w", shop.ID, shop.Domain, err)
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated for potential_shop (id=%s, domain=%s) - record may not exist", shop.ID, shop.Domain)
	}

	return nil
}

