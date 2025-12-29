package storage

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/producttypes"
)

// ProductTypesAdapter адаптер для работы с типами товаров
type ProductTypesAdapter struct {
	*BaseAdapter
}

// NewProductTypesAdapter создаёт новый адаптер для типов товаров
func NewProductTypesAdapter(pg *Postgres) producttypes.Storage {
	return &ProductTypesAdapter{
		BaseAdapter: NewBaseAdapter(pg, nil),
	}
}

// GetByID получает тип товара по ID
func (a *ProductTypesAdapter) GetByID(id string) (*producttypes.ProductType, error) {
	ptUUID, err := a.ParseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product type ID: %w", err)
	}

	query := `
		SELECT id, code, name_sr, is_active
		FROM product_types
		WHERE id = $1
	`

	var pt producttypes.ProductType
	err = a.pg.DB().QueryRow(a.GetContext(), query, ptUUID).Scan(
		&pt.ID,
		&pt.Code,
		&pt.NameSr,
		&pt.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product type not found")
		}
		return nil, fmt.Errorf("failed to get product type: %w", err)
	}

	return &pt, nil
}

// GetByCode получает тип товара по коду
func (a *ProductTypesAdapter) GetByCode(code string) (*producttypes.ProductType, error) {
	query := `
		SELECT id, code, name_sr, is_active
		FROM product_types
		WHERE code = $1 AND is_active = true
	`

	var pt producttypes.ProductType
	err := a.pg.DB().QueryRow(a.GetContext(), query, code).Scan(
		&pt.ID,
		&pt.Code,
		&pt.NameSr,
		&pt.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("product type not found")
		}
		return nil, fmt.Errorf("failed to get product type by code: %w", err)
	}

	return &pt, nil
}

// GetAllActive получает все активные типы товаров
func (a *ProductTypesAdapter) GetAllActive() ([]*producttypes.ProductType, error) {
	query := `
		SELECT id, code, name_sr, is_active
		FROM product_types
		WHERE is_active = true
		ORDER BY code
	`

	rows, err := a.pg.DB().Query(a.GetContext(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all active product types: %w", err)
	}
	defer rows.Close()

	var result []*producttypes.ProductType
	for rows.Next() {
		var pt producttypes.ProductType
		if err := rows.Scan(
			&pt.ID,
			&pt.Code,
			&pt.NameSr,
			&pt.IsActive,
		); err != nil {
			continue
		}
		result = append(result, &pt)
	}

	return result, nil
}

// GetByCategoryID получает типы товаров для категории
func (a *ProductTypesAdapter) GetByCategoryID(categoryID string) ([]*producttypes.ProductType, error) {
	categoryUUID, err := a.ParseUUID(categoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category ID: %w", err)
	}

	query := `
		SELECT pt.id, pt.code, pt.name_sr, pt.is_active
		FROM product_types pt
		INNER JOIN category_product_types cpt ON pt.id = cpt.product_type_id
		WHERE cpt.category_id = $1 AND pt.is_active = true
		ORDER BY pt.code
	`

	rows, err := a.pg.DB().Query(a.GetContext(), query, categoryUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product types by category: %w", err)
	}
	defer rows.Close()

	var result []*producttypes.ProductType
	for rows.Next() {
		var pt producttypes.ProductType
		if err := rows.Scan(
			&pt.ID,
			&pt.Code,
			&pt.NameSr,
			&pt.IsActive,
		); err != nil {
			continue
		}
		result = append(result, &pt)
	}

	return result, nil
}

