package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/matching"
)

// MatchingAdapter адаптер для работы с matching
type MatchingAdapter struct {
	*BaseAdapter
}

// NewMatchingAdapter создаёт новый адаптер для matching
func NewMatchingAdapter(pg *Postgres) matching.Storage {
	return &MatchingAdapter{
		BaseAdapter: NewBaseAdapter(pg, nil),
	}
}

// FindSimilarProducts ищет похожие товары или услуги
func (a *MatchingAdapter) FindSimilarProducts(name, brand string, productType string, limit int) ([]*matching.Product, error) {
	var query string
	var args []interface{}

	// Нормализуем входные данные для поиска
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	normalizedBrand := strings.ToLower(strings.TrimSpace(brand))

	// Определяем тип продукта для фильтрации
	if productType == "" {
		productType = "good" // По умолчанию товар
	}

	// Поиск по названию и бренду
	if normalizedBrand != "" {
		query = `
			SELECT id, name, brand, specs, COALESCE(type, 'good') as type
			FROM products
			WHERE (
				LOWER(TRIM(name)) = $1
				OR LOWER(TRIM(name)) LIKE $2
				OR LOWER(TRIM(name)) LIKE $3
			)
			AND (
				brand = '' 
				OR LOWER(TRIM(brand)) = $4
				OR LOWER(TRIM(brand)) LIKE $5
			)
			AND COALESCE(type, 'good') = $6
			ORDER BY 
				CASE WHEN LOWER(TRIM(name)) = $1 THEN 1 ELSE 2 END,
				CASE WHEN LOWER(TRIM(brand)) = $4 THEN 1 ELSE 2 END
			LIMIT $7
		`
		nameExact := normalizedName
		namePrefix := normalizedName + "%"
		nameContains := "%" + normalizedName + "%"
		brandExact := normalizedBrand
		brandContains := "%" + normalizedBrand + "%"
		args = []interface{}{nameExact, namePrefix, nameContains, brandExact, brandContains, productType, limit}
	} else {
		// Улучшенный поиск: ищем по ключевым словам из названия
		// Разбиваем название на слова и ищем товары, содержащие несколько ключевых слов
		words := strings.Fields(normalizedName)
		keyWords := make([]string, 0)
		for _, word := range words {
			// Берем только значимые слова (длиннее 2 символов, не числа меньше 10)
			if len(word) > 2 {
				keyWords = append(keyWords, word)
			}
		}

		// Если есть ключевые слова, ищем товары, содержащие хотя бы 2 из них
		if len(keyWords) >= 2 {
			// Строим условие: товар должен содержать несколько ключевых слов
			conditions := make([]string, 0)
			queryArgs := make([]interface{}, 0)
			argIndex := 1

			// Точное совпадение (приоритет)
			conditions = append(conditions, fmt.Sprintf("LOWER(TRIM(name)) = $%d", argIndex))
			queryArgs = append(queryArgs, normalizedName)
			argIndex++

			// Поиск по ключевым словам (хотя бы 2 совпадения)
			for i := 0; i < len(keyWords) && i < 5; i++ { // Берем до 5 ключевых слов
				conditions = append(conditions, fmt.Sprintf("LOWER(TRIM(name)) LIKE $%d", argIndex))
				queryArgs = append(queryArgs, "%"+keyWords[i]+"%")
				argIndex++
			}

			// Добавляем фильтр по типу продукта
			typeCondition := fmt.Sprintf(" AND COALESCE(type, 'good') = $%d", argIndex)
			queryArgs = append(queryArgs, productType)
			argIndex++
			
			query = fmt.Sprintf(`
				SELECT id, name, brand, specs, COALESCE(type, 'good') as type
				FROM products
				WHERE (%s)%s
				ORDER BY 
					CASE WHEN LOWER(TRIM(name)) = $1 THEN 1 ELSE 2 END,
					name
				LIMIT $%d
			`, strings.Join(conditions, " OR "), typeCondition, argIndex)
			queryArgs = append(queryArgs, limit)
			args = queryArgs
		} else {
			// Fallback на старый поиск
			query = `
				SELECT id, name, brand, specs, COALESCE(type, 'good') as type
				FROM products
				WHERE (
					LOWER(TRIM(name)) = $1
					OR LOWER(TRIM(name)) LIKE $2
					OR LOWER(TRIM(name)) LIKE $3
				)
				AND COALESCE(type, 'good') = $4
				ORDER BY 
					CASE WHEN LOWER(TRIM(name)) = $1 THEN 1 ELSE 2 END
				LIMIT $5
			`
			nameExact := normalizedName
			namePrefix := normalizedName + "%"
			nameContains := "%" + normalizedName + "%"
			args = []interface{}{nameExact, namePrefix, nameContains, productType, limit}
		}
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
			&product.Type,
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
	productUUID, err := a.ParseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		SELECT id, name, brand, specs, COALESCE(type, 'good') as type
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
		&product.Type,
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
	productUUID, err := a.ParseUUID(match.ProductID)
	if err != nil {
		return fmt.Errorf("invalid product ID: %w", err)
	}

	matchedUUID, err := a.ParseUUID(match.MatchedID)
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
	productUUID, err := a.ParseUUID(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// Оптимизированный запрос: использует индекс idx_product_matches_similarity
	query := `
		SELECT product_id, matched_id, similarity, confidence, matched_at
		FROM product_matches
		WHERE product_id = $1
		ORDER BY similarity DESC, matched_at DESC
		LIMIT 100
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
