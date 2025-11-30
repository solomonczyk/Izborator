package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/pricehistory"
)

// PriceHistoryAdapter адаптер для работы с историей цен через PostgreSQL
type PriceHistoryAdapter struct {
	pg  *Postgres
	ctx context.Context
}

// NewPriceHistoryAdapter создаёт новый адаптер для истории цен
func NewPriceHistoryAdapter(pg *Postgres) pricehistory.Storage {
	return &PriceHistoryAdapter{
		pg:  pg,
		ctx: context.Background(),
	}
}

// SavePrice сохраняет точку цены
// В текущей реализации используем product_prices, где updated_at показывает время обновления
// Для полноценной истории нужно будет создать отдельную таблицу price_history
func (a *PriceHistoryAdapter) SavePrice(point *pricehistory.PricePoint) error {
	// Пока что сохранение происходит через ProductsAdapter.SavePrice
	// Здесь можно добавить логику для сохранения в отдельную таблицу истории
	return nil
}

// GetHistory получает историю цен за период из product_prices
// Используем updated_at как timestamp изменения цены
func (a *PriceHistoryAdapter) GetHistory(productID string, from, to time.Time) ([]*pricehistory.PricePoint, error) {
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	query := `
		SELECT shop_id, price, currency, updated_at
		FROM product_prices
		WHERE product_id = $1
		  AND updated_at >= $2
		  AND updated_at <= $3
		ORDER BY updated_at ASC, shop_id
	`

	rows, err := a.pg.DB().Query(a.ctx, query, productUUID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query price history: %w", err)
	}
	defer rows.Close()

	var points []*pricehistory.PricePoint
	for rows.Next() {
		var point pricehistory.PricePoint
		var timestamp time.Time

		err := rows.Scan(
			&point.ShopID,
			&point.Price,
			&point.Currency,
			&timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price point: %w", err)
		}

		point.ProductID = productID
		point.Timestamp = timestamp
		points = append(points, &point)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price history: %w", err)
	}

	return points, nil
}

// GetPriceChart получает данные для графика цен
// Группирует цены по магазинам и возвращает структуру для отображения графика
func (a *PriceHistoryAdapter) GetPriceChart(productID string, period string, shopIDs []string) (*pricehistory.PriceChart, error) {
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// Определяем период
	now := time.Now()
	var from time.Time
	switch period {
	case "day":
		from = now.AddDate(0, 0, -1)
	case "week":
		from = now.AddDate(0, 0, -7)
	case "month":
		from = now.AddDate(0, -1, 0)
	case "year":
		from = now.AddDate(-1, 0, 0)
	default:
		from = now.AddDate(0, 0, -30) // По умолчанию 30 дней
	}

	// Строим запрос с shop_name
	query := `
		SELECT pp.shop_id, pp.price, pp.currency, pp.updated_at, pp.shop_name
		FROM product_prices pp
		WHERE pp.product_id = $1
		  AND pp.updated_at >= $2
	`
	args := []interface{}{productUUID, from}

	// Фильтр по магазинам (если указаны)
	if len(shopIDs) > 0 {
		query += " AND pp.shop_id = ANY($3)"
		args = append(args, shopIDs)
	}

	query += " ORDER BY pp.updated_at ASC, pp.shop_id"

	rows, err := a.pg.DB().Query(a.ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query price chart: %w", err)
	}
	defer rows.Close()

	// Группируем по магазинам и собираем названия
	shops := make(map[string][]*pricehistory.PricePoint)
	shopNames := make(map[string]string)

	for rows.Next() {
		var point pricehistory.PricePoint
		var timestamp time.Time
		var shopName string

		err := rows.Scan(
			&point.ShopID,
			&point.Price,
			&point.Currency,
			&timestamp,
			&shopName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price point: %w", err)
		}

		point.ProductID = productID
		point.Timestamp = timestamp
		shops[point.ShopID] = append(shops[point.ShopID], &point)
		shopNames[point.ShopID] = shopName
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating price chart: %w", err)
	}

	return &pricehistory.PriceChart{
		ProductID: productID,
		Shops:     shops,
		ShopNames: shopNames,
		Period:    period,
		From:      from,
		To:        now,
	}, nil
}

// CleanupOldData удаляет старые данные
// В текущей реализации не требуется, т.к. используем product_prices
func (a *PriceHistoryAdapter) CleanupOldData(before time.Time) error {
	// TODO: реализовать очистку старых данных из отдельной таблицы истории
	return nil
}

