package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/attributes"
)

// AttributesAdapter адаптер для работы с атрибутами
type AttributesAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewAttributesAdapter создаёт новый адаптер для атрибутов
func NewAttributesAdapter(pg *Postgres) attributes.Storage {
	return &AttributesAdapter{
		pg:  pg,
		ctx: pg.Context(),
	}
}

// GetByID получает атрибут по ID
func (a *AttributesAdapter) GetByID(id string) (*attributes.Attribute, error) {
	attrUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute ID: %w", err)
	}

	query := `
		SELECT id, code, name_sr, data_type, unit_sr, is_filterable, is_sortable
		FROM attributes
		WHERE id = $1
	`

	var attr attributes.Attribute
	err = a.pg.DB().QueryRow(a.ctx, query, attrUUID).Scan(
		&attr.ID,
		&attr.Code,
		&attr.NameSr,
		&attr.DataType,
		&attr.UnitSr,
		&attr.IsFilterable,
		&attr.IsSortable,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attribute not found")
		}
		return nil, fmt.Errorf("failed to get attribute: %w", err)
	}

	return &attr, nil
}

// GetByCode получает атрибут по коду
func (a *AttributesAdapter) GetByCode(code string) (*attributes.Attribute, error) {
	query := `
		SELECT id, code, name_sr, data_type, unit_sr, is_filterable, is_sortable
		FROM attributes
		WHERE code = $1
	`

	var attr attributes.Attribute
	err := a.pg.DB().QueryRow(a.ctx, query, code).Scan(
		&attr.ID,
		&attr.Code,
		&attr.NameSr,
		&attr.DataType,
		&attr.UnitSr,
		&attr.IsFilterable,
		&attr.IsSortable,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attribute not found")
		}
		return nil, fmt.Errorf("failed to get attribute by code: %w", err)
	}

	return &attr, nil
}

// GetAllActive получает все активные атрибуты
func (a *AttributesAdapter) GetAllActive() ([]*attributes.Attribute, error) {
	query := `
		SELECT id, code, name_sr, data_type, unit_sr, is_filterable, is_sortable
		FROM attributes
		ORDER BY code
	`

	rows, err := a.pg.DB().Query(a.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all attributes: %w", err)
	}
	defer rows.Close()

	var result []*attributes.Attribute
	for rows.Next() {
		var attr attributes.Attribute
		if err := rows.Scan(
			&attr.ID,
			&attr.Code,
			&attr.NameSr,
			&attr.DataType,
			&attr.UnitSr,
			&attr.IsFilterable,
			&attr.IsSortable,
		); err != nil {
			continue
		}
		result = append(result, &attr)
	}

	return result, nil
}

// GetByProductTypeID получает атрибуты для типа товара
func (a *AttributesAdapter) GetByProductTypeID(productTypeID string) ([]*attributes.ProductTypeAttribute, error) {
	ptUUID, err := uuid.Parse(productTypeID)
	if err != nil {
		return nil, fmt.Errorf("invalid product type ID: %w", err)
	}

	query := `
		SELECT product_type_id, attribute_id, is_required, sort_order
		FROM product_type_attributes
		WHERE product_type_id = $1
		ORDER BY sort_order
	`

	rows, err := a.pg.DB().Query(a.ctx, query, ptUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attributes by product type: %w", err)
	}
	defer rows.Close()

	var result []*attributes.ProductTypeAttribute
	for rows.Next() {
		var pta attributes.ProductTypeAttribute
		if err := rows.Scan(
			&pta.ProductTypeID,
			&pta.AttributeID,
			&pta.IsRequired,
			&pta.SortOrder,
		); err != nil {
			continue
		}
		result = append(result, &pta)
	}

	return result, nil
}
