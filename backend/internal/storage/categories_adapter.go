package storage

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/categories"
)

// CategoriesAdapter адаптер для работы с категориями
type CategoriesAdapter struct {
	*BaseAdapter
}

// NewCategoriesAdapter создаёт новый адаптер для категорий
func NewCategoriesAdapter(pg *Postgres) categories.Storage {
	return &CategoriesAdapter{
		BaseAdapter: NewBaseAdapter(pg, nil),
	}
}

// GetByID получает категорию по ID
func (a *CategoriesAdapter) GetByID(id string) (*categories.Category, error) {
	categoryUUID, err := a.ParseUUID(id)
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

	err = a.pg.DB().QueryRow(a.GetContext(), query, categoryUUID).Scan(
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

	err := a.pg.DB().QueryRow(a.GetContext(), query, slug).Scan(
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
	parentUUID, err := a.ParseUUID(parentID)
	if err != nil {
		return nil, fmt.Errorf("invalid parent category ID: %w", err)
	}

	query := `
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, level, is_active, sort_order
		FROM categories
		WHERE parent_id = $1 AND is_active = true
		ORDER BY sort_order, name_sr
	`

	rows, err := a.pg.DB().Query(a.GetContext(), query, parentUUID)
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
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, name_ru, name_en, name_hu, name_zh, level, is_active, sort_order
		FROM categories
		WHERE is_active = true
		ORDER BY sort_order, name_sr
	`

	rows, err := a.pg.DB().Query(a.GetContext(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all active categories: %w", err)
	}
	defer rows.Close()

	var result []*categories.Category
	for rows.Next() {
		var cat categories.Category
		var parentID *uuid.UUID
		var nameRu, nameEn, nameHu, nameZh *string

		if err := rows.Scan(
			&cat.ID,
			&parentID,
			&cat.Slug,
			&cat.Code,
			&cat.NameSr,
			&cat.NameSrLc,
			&nameRu,
			&nameEn,
			&nameHu,
			&nameZh,
			&cat.Level,
			&cat.IsActive,
			&cat.SortOrder,
		); err != nil {
			continue
		}

		cat.NameRu = nameRu
		cat.NameEn = nameEn
		cat.NameHu = nameHu
		cat.NameZh = nameZh

		if parentID != nil {
			parentIDStr := parentID.String()
			cat.ParentID = &parentIDStr
		}

		result = append(result, &cat)
	}

	return result, nil
}

// GetTree получает дерево категорий (все корневые + их дети)
// Оптимизировано: один запрос вместо двух
func (a *CategoriesAdapter) GetTree() ([]*categories.Category, error) {
	// Оптимизированный запрос: получаем все активные категории одним запросом
	// Используем UNION для объединения корневых и дочерних категорий
	query := `
		SELECT id, parent_id, slug, code, name_sr, name_sr_lc, name_ru, name_en, name_hu, name_zh, level, is_active, sort_order
		FROM categories
		WHERE is_active = true
		ORDER BY 
			CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END,
			parent_id NULLS FIRST,
			sort_order,
			name_sr
	`

	rows, err := a.pg.DB().Query(a.GetContext(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories tree: %w", err)
	}
	defer rows.Close()

	result := make([]*categories.Category, 0) // Явно инициализируем как пустой слайс, не nil
	for rows.Next() {
		var cat categories.Category
		var parentID *uuid.UUID
		var nameRu, nameEn, nameHu, nameZh *string

		if err := rows.Scan(
			&cat.ID,
			&parentID,
			&cat.Slug,
			&cat.Code,
			&cat.NameSr,
			&cat.NameSrLc,
			&nameRu,
			&nameEn,
			&nameHu,
			&nameZh,
			&cat.Level,
			&cat.IsActive,
			&cat.SortOrder,
		); err != nil {
			continue
		}

		cat.NameRu = nameRu
		cat.NameEn = nameEn
		cat.NameHu = nameHu
		cat.NameZh = nameZh

		if parentID != nil {
			parentIDStr := parentID.String()
			cat.ParentID = &parentIDStr
		}

		result = append(result, &cat)
	}

	return result, nil
}

