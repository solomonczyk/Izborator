package pricehistory

import (
	"time"

	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с time-series хранилищем цен
type Storage interface {
	// SavePrice сохраняет точку цены
	SavePrice(point *PricePoint) error
	
	// GetHistory получает историю цен за период
	GetHistory(productID string, from, to time.Time) ([]*PricePoint, error)
	
	// GetPriceChart получает данные для графика цен
	GetPriceChart(productID string, period string, shopIDs []string) (*PriceChart, error)
	
	// CleanupOldData удаляет старые данные
	CleanupOldData(before time.Time) error
}

// Service сервис для работы с историей цен
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис истории цен
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}

