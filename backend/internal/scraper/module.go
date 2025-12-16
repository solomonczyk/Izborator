package scraper

import (
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/scrapingstats"
)

const (
	defaultRetryLimit     = 3
	maxRetryLimit         = 5
	defaultRetryBackoffMs = 3000
	maxRetryBackoffMs     = 60000
)

// Storage интерфейс для работы с хранилищем данных парсинга
type Storage interface {
	// SaveRawProduct сохраняет сырые данные товара
	SaveRawProduct(data *RawProduct) error

	// GetShopConfig получает конфигурацию магазина
	GetShopConfig(shopID string) (*ShopConfig, error)

	// ListShops получает список всех магазинов
	ListShops() ([]*ShopConfig, error)

	// GetUnprocessedRawProducts получает необработанные сырые данные товаров
	GetUnprocessedRawProducts(limit int) ([]*RawProduct, error)

	// MarkRawProductAsProcessed помечает сырой товар как обработанный
	MarkRawProductAsProcessed(shopID, externalID string) error
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
	stats   *scrapingstats.Service
}

// CatalogResult результат парсинга каталога
type CatalogResult struct {
	ProductURLs []string `json:"product_urls"`
	NextPageURL string   `json:"next_page_url,omitempty"`
	TotalFound  int      `json:"total_found"`
}

// New создаёт новый сервис парсеров
func New(storage Storage, queue Queue, stats *scrapingstats.Service, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		queue:   queue,
		logger:  log,
		stats:   stats,
	}
}
