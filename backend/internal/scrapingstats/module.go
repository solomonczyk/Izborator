package scrapingstats

import (
	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с хранилищем статистики парсинга
type Storage interface {
	// SaveStat сохраняет статистику парсинга
	SaveStat(stat *ScrapingStat) error
	
	// GetShopStats получает статистику по магазину
	GetShopStats(shopID string, days int) (*ShopStats, error)
	
	// GetOverallStats получает общую статистику
	GetOverallStats(days int) (*OverallStats, error)
	
	// GetRecentStats получает последние N записей статистики
	GetRecentStats(limit int) ([]*ScrapingStat, error)
	
	// UpdateShopLastScraped обновляет last_scraped_at для магазина
	UpdateShopLastScraped(shopID string) error
}

// Service сервис для работы со статистикой парсинга
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис статистики
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}

// RecordStat записывает статистику парсинга
func (s *Service) RecordStat(stat *ScrapingStat) error {
	if err := s.storage.SaveStat(stat); err != nil {
		s.logger.Error("Failed to save scraping stat", map[string]interface{}{
			"error":   err.Error(),
			"shop_id": stat.ShopID,
		})
		return err
	}
	
	// Обновляем last_scraped_at для магазина
	if stat.Status == "success" || stat.Status == "partial" {
		if err := s.storage.UpdateShopLastScraped(stat.ShopID); err != nil {
			s.logger.Warn("Failed to update shop last_scraped_at", map[string]interface{}{
				"error":   err.Error(),
				"shop_id": stat.ShopID,
			})
		}
	}
	
	return nil
}

// GetShopStats получает статистику по магазину
func (s *Service) GetShopStats(shopID string, days int) (*ShopStats, error) {
	return s.storage.GetShopStats(shopID, days)
}

// GetOverallStats получает общую статистику
func (s *Service) GetOverallStats(days int) (*OverallStats, error) {
	return s.storage.GetOverallStats(days)
}

// GetRecentStats получает последние записи статистики
func (s *Service) GetRecentStats(limit int) ([]*ScrapingStat, error) {
	return s.storage.GetRecentStats(limit)
}

