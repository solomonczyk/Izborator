package storage

import (
	"time"

	"github.com/solomonczyk/izborator/internal/pricehistory"
)

// PriceHistoryAdapter адаптер для работы с историей цен
type PriceHistoryAdapter struct {
	// TODO: добавить клиент для InfluxDB или другого time-series хранилища
}

// NewPriceHistoryAdapter создаёт новый адаптер для истории цен
func NewPriceHistoryAdapter() pricehistory.Storage {
	return &PriceHistoryAdapter{}
}

// SavePrice сохраняет точку цены
func (a *PriceHistoryAdapter) SavePrice(point *pricehistory.PricePoint) error {
	// TODO: реализовать сохранение в time-series БД (InfluxDB)
	return nil
}

// GetHistory получает историю цен за период
func (a *PriceHistoryAdapter) GetHistory(productID string, from, to time.Time) ([]*pricehistory.PricePoint, error) {
	// TODO: реализовать получение истории из time-series БД
	return nil, nil
}

// GetPriceChart получает данные для графика цен
func (a *PriceHistoryAdapter) GetPriceChart(productID string, period string, shopIDs []string) (*pricehistory.PriceChart, error) {
	// TODO: реализовать получение данных для графика
	return nil, nil
}

// CleanupOldData удаляет старые данные
func (a *PriceHistoryAdapter) CleanupOldData(before time.Time) error {
	// TODO: реализовать очистку старых данных
	return nil
}

