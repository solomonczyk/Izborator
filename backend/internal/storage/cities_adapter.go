package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/solomonczyk/izborator/internal/cities"
)

// CitiesAdapter адаптер для работы с городами
type CitiesAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewCitiesAdapter создаёт новый адаптер для городов
func NewCitiesAdapter(pg *Postgres) cities.Storage {
	return &CitiesAdapter{
		pg:  pg,
		ctx: pg.Context(),
	}
}

// GetByID получает город по ID
func (a *CitiesAdapter) GetByID(id string) (*cities.City, error) {
	cityUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid city ID: %w", err)
	}

	query := `
		SELECT id, slug, name_sr, region_sr, sort_order, is_active
		FROM cities
		WHERE id = $1
	`

	var city cities.City
	err = a.pg.DB().QueryRow(a.ctx, query, cityUUID).Scan(
		&city.ID,
		&city.Slug,
		&city.NameSr,
		&city.RegionSr,
		&city.SortOrder,
		&city.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("city not found")
		}
		return nil, fmt.Errorf("failed to get city: %w", err)
	}

	return &city, nil
}

// GetBySlug получает город по slug
func (a *CitiesAdapter) GetBySlug(slug string) (*cities.City, error) {
	query := `
		SELECT id, slug, name_sr, region_sr, sort_order, is_active
		FROM cities
		WHERE slug = $1 AND is_active = true
	`

	var city cities.City
	err := a.pg.DB().QueryRow(a.ctx, query, slug).Scan(
		&city.ID,
		&city.Slug,
		&city.NameSr,
		&city.RegionSr,
		&city.SortOrder,
		&city.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("city not found")
		}
		return nil, fmt.Errorf("failed to get city by slug: %w", err)
	}

	return &city, nil
}

// GetAllActive получает все активные города
func (a *CitiesAdapter) GetAllActive() ([]*cities.City, error) {
	query := `
		SELECT id, slug, name_sr, region_sr, sort_order, is_active
		FROM cities
		WHERE is_active = true
		ORDER BY sort_order, name_sr
	`

	rows, err := a.pg.DB().Query(a.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all active cities: %w", err)
	}
	defer rows.Close()

	var result []*cities.City
	for rows.Next() {
		var city cities.City
		if err := rows.Scan(
			&city.ID,
			&city.Slug,
			&city.NameSr,
			&city.RegionSr,
			&city.SortOrder,
			&city.IsActive,
		); err != nil {
			continue
		}
		result = append(result, &city)
	}

	return result, nil
}


