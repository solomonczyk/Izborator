package scraper

import (
	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с хранилищем данных парсинга
type Storage interface {
	// SaveRawProduct сохраняет сырые данные товара
	SaveRawProduct(data *RawProduct) error
	
	// GetShopConfig получает конфигурацию магазина
	GetShopConfig(shopID string) (*ShopConfig, error)
	
	// ListShops получает список всех магазинов
	ListShops() ([]*ShopConfig, error)
}

// Queue интерфейс для отправки данных в очередь
type Queue interface {
	// Publish отправляет сообщение в очередь
	Publish(topic string, data interface{}) error
}

// Service сервис для парсинга данных с сайтов магазинов
type Service struct {
	storage Storage
	queue   Queue
	logger  *logger.Logger
}

// New создаёт новый сервис парсеров
func New(storage Storage, queue Queue, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		queue:   queue,
		logger:  log,
	}
}

