package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/matching"
)

// MatchingAdapter адаптер для работы с matching
type MatchingAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewMatchingAdapter создаёт новый адаптер для matching
func NewMatchingAdapter(pg *Postgres) matching.Storage {
	return &MatchingAdapter{
		pg:  pg,
		ctx: pg.Context(),
	}
}

// FindSimilarProducts ищет похожие товары
func (a *MatchingAdapter) FindSimilarProducts(name, brand string, limit int) ([]*matching.Product, error) {
	var query string
	var args []interface{}

	// Поиск по названию и бренду
	if brand != "" {
		query = `
			SELECT id, name, brand, specs
			FROM products
			WHERE (name ILIKE $1 OR name ILIKE $2)
			  AND (brand ILIKE $3 OR brand = '')
			ORDER BY 
				CASE WHEN name ILIKE $1 THEN 1 ELSE 2 END,
				CASE WHEN brand ILIKE $3 THEN 1 ELSE 2 END
			LIMIT $4
		`
		nameExact := fmt.Sprintf("%%%s%%", name)
		nameWords := fmt.Sprintf("%%%s%%", strings.Join(strings.Fields(name), "%"))
		args = []interface{}{nameExact, nameWords, fmt.Sprintf("%%%s%%", brand), limit}
	} else {
		query = `
			SELECT id, name, brand, specs
			FROM products
			WHERE name ILIKE $1 OR name ILIKE $2
			ORDER BY 
				CASE WHEN name ILIKE $1 THEN 1 ELSE 2 END
			LIMIT $3
		`
		nameExact := fmt.Sprintf("%%%s%%", name)
		nameWords := fmt.Sprintf("%%%s%%", strings.Join(strings.Fields(name), "%"))
		args = []interface{}{nameExact, nameWords, limit}
	}

	rows, err := a.pg.DB().Query(a.ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar products: %w", err)
	}
	defer rows.Close()

	var products []*matching.Product

	for rows.Next() {
		var product matching.Product
		var specsJSON []byte

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Brand,
			&specsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		// Парсим specs для matching (упрощённо, только ключи)
		if len(specsJSON) > 0 {
			// Для matching нам нужны только ключи specs, не значения
			// Упрощённая версия - можно улучшить
			product.Specs = make(map[string]string)
		} else {
			product.Specs = make(map[string]string)
		}

		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// GetProductByID получает товар по ID
func (a *MatchingAdapter) GetProductByID(id string) (*matching.Product, error) {
	productUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		SELECT id, name, brand, specs
		FROM products
		WHERE id = $1
	`

	var product matching.Product
	var specsJSON []byte

	err = a.pg.DB().QueryRow(a.ctx, query, productUUID).Scan(
		&product.ID,
		&product.Name,
		&product.Brand,
		&specsJSON,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, matching.ErrMatchNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Упрощённая обработка specs для matching
	if len(specsJSON) > 0 {
		product.Specs = make(map[string]string)
	} else {
		product.Specs = make(map[string]string)
	}

	return &product, nil
}

// SaveMatch сохраняет результат сопоставления
func (a *MatchingAdapter) SaveMatch(match *matching.ProductMatch) error {
	productUUID, err := uuid.Parse(match.ProductID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	matchedUUID, err := uuid.Parse(match.MatchedID)
	if err != nil {
		return fmt.Errorf("invalid matched ID: %w", err)
	}

	query := `
		INSERT INTO product_matches (product_id, matched_id, similarity, confidence, matched_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (product_id, matched_id) DO UPDATE SET
			similarity = EXCLUDED.similarity,
			confidence = EXCLUDED.confidence,
			matched_at = EXCLUDED.matched_at
	`

	if match.MatchedAt.IsZero() {
		match.MatchedAt = time.Now()
	}

	if match.Confidence == "" {
		match.Confidence = "medium"
	}

	_, err = a.pg.DB().Exec(a.ctx, query,
		productUUID,
		matchedUUID,
		match.Similarity,
		match.Confidence,
		match.MatchedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save match: %w", err)
	}

	return nil
}

// GetMatches получает все сопоставления для товара
func (a *MatchingAdapter) GetMatches(productID string) ([]*matching.ProductMatch, error) {
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		SELECT product_id, matched_id, similarity, confidence, matched_at
		FROM product_matches
		WHERE product_id = $1
		ORDER BY similarity DESC, matched_at DESC
	`

	rows, err := a.pg.DB().Query(a.ctx, query, productUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}
	defer rows.Close()

	var matches []*matching.ProductMatch

	for rows.Next() {
		var match matching.ProductMatch
		var matchedAt time.Time

		err := rows.Scan(
			&match.ProductID,
			&match.MatchedID,
			&match.Similarity,
			&match.Confidence,
			&matchedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %w", err)
		}

		match.MatchedAt = matchedAt
		matches = append(matches, &match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating matches: %w", err)
	}

	return matches, nil
}
