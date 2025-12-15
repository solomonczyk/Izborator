package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/categories"
)

// CategoriesAdapter адаптер для работы с категориями
type CategoriesAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewCategoriesAdapter создаёт новый адаптер для категорий
func NewCategoriesAdapter(pg *Postgres) categories.Storage {
	return &CategoriesAdapter{
		pg:  pg,
		ctx: pg.Context(),
	}
}

// GetByID получает категорию по ID
func (a *CategoriesAdapter) GetByID(id string) (*categories.Category, error) {
	categoryUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid category ID: %w", err)
	}

	query := `
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, level, is_active, sort_order
		FROM categories
		WHERE id = $1
	`

	var cat categories.Category
	var parentID *uuid.UUID

	err = a.pg.DB().QueryRow(a.ctx, query, categoryUUID).Scan(
		&cat.ID,
		&parentID,
		&cat.Slug,
		&cat.Code,
		&cat.NameSr,
		&cat.NameSrLc,
		&cat.Level,
		&cat.IsActive,
		&cat.SortOrder,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if parentID != nil {
		parentIDStr := parentID.String()
		cat.ParentID = &parentIDStr
	}

	return &cat, nil
}

// GetBySlug получает категорию по slug
func (a *CategoriesAdapter) GetBySlug(slug string) (*categories.Category, error) {
	query := `
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, level, is_active, sort_order
		FROM categories
		WHERE slug = $1 AND is_active = true
	`

	var cat categories.Category
	var parentID *uuid.UUID

	err := a.pg.DB().QueryRow(a.ctx, query, slug).Scan(
		&cat.ID,
		&parentID,
		&cat.Slug,
		&cat.Code,
		&cat.NameSr,
		&cat.NameSrLc,
		&cat.Level,
		&cat.IsActive,
		&cat.SortOrder,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	if parentID != nil {
		parentIDStr := parentID.String()
		cat.ParentID = &parentIDStr
	}

	return &cat, nil
}

// GetByParentID получает все подкатегории родительской категории
func (a *CategoriesAdapter) GetByParentID(parentID string) ([]*categories.Category, error) {
	parentUUID, err := uuid.Parse(parentID)
	if err != nil {
		return nil, fmt.Errorf("invalid parent category ID: %w", err)
	}

	query := `
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, level, is_active, sort_order
		FROM categories
		WHERE parent_id = $1 AND is_active = true
		ORDER BY sort_order, name_sr
	`

	rows, err := a.pg.DB().Query(a.ctx, query, parentUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories by parent: %w", err)
	}
	defer rows.Close()

	var result []*categories.Category
	for rows.Next() {
		var cat categories.Category
		var parentID *uuid.UUID

		if err := rows.Scan(
			&cat.ID,
			&parentID,
			&cat.Slug,
			&cat.Code,
			&cat.NameSr,
			&cat.NameSrLc,
			&cat.Level,
			&cat.IsActive,
			&cat.SortOrder,
		); err != nil {
			continue
		}

		if parentID != nil {
			parentIDStr := parentID.String()
			cat.ParentID = &parentIDStr
		}

		result = append(result, &cat)
	}

	return result, nil
}

// GetAllActive получает все активные категории
func (a *CategoriesAdapter) GetAllActive() ([]*categories.Category, error) {
	query := `
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, level, is_active, sort_order
		FROM categories
		WHERE is_active = true
		ORDER BY sort_order, name_sr
	`

	rows, err := a.pg.DB().Query(a.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all active categories: %w", err)
	}
	defer rows.Close()

	var result []*categories.Category
	for rows.Next() {
		var cat categories.Category
		var parentID *uuid.UUID

		if err := rows.Scan(
			&cat.ID,
			&parentID,
			&cat.Slug,
			&cat.Code,
			&cat.NameSr,
			&cat.NameSrLc,
			&cat.Level,
			&cat.IsActive,
			&cat.SortOrder,
		); err != nil {
			continue
		}

		if parentID != nil {
			parentIDStr := parentID.String()
			cat.ParentID = &parentIDStr
		}

		result = append(result, &cat)
	}

	return result, nil
}

// GetTree получает дерево категорий (все корневые + их дети)
func (a *CategoriesAdapter) GetTree() ([]*categories.Category, error) {
	// Сначала получаем все корневые категории
	rootQuery := `
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, level, is_active, sort_order
		FROM categories
		WHERE parent_id IS NULL AND is_active = true
		ORDER BY sort_order, name_sr
	`

	rows, err := a.pg.DB().Query(a.ctx, rootQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get root categories: %w", err)
	}
	defer rows.Close()

	result := make([]*categories.Category, 0) // Явно инициализируем как пустой слайс, не nil
	for rows.Next() {
		var cat categories.Category
		var parentID *uuid.UUID

		if err := rows.Scan(
			&cat.ID,
			&parentID,
			&cat.Slug,
			&cat.Code,
			&cat.NameSr,
			&cat.NameSrLc,
			&cat.Level,
			&cat.IsActive,
			&cat.SortOrder,
		); err != nil {
			continue
		}

		// Для корневых категорий parentID должен быть nil
		if parentID != nil {
			parentIDStr := parentID.String()
			cat.ParentID = &parentIDStr
		}

		result = append(result, &cat)
	}

	// Затем получаем все подкатегории
	childrenQuery := `
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, level, is_active, sort_order
		FROM categories
		WHERE parent_id IS NOT NULL AND is_active = true
		ORDER BY parent_id, sort_order, name_sr
	`

	rows, err = a.pg.DB().Query(a.ctx, childrenQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get child categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cat categories.Category
		var parentID *uuid.UUID

		if err := rows.Scan(
			&cat.ID,
			&parentID,
			&cat.Slug,
			&cat.Code,
			&cat.NameSr,
			&cat.NameSrLc,
			&cat.Level,
			&cat.IsActive,
			&cat.SortOrder,
		); err != nil {
			continue
		}

		if parentID != nil {
			parentIDStr := parentID.String()
			cat.ParentID = &parentIDStr
		}

		result = append(result, &cat)
	}

	return result, nil
}

