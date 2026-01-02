package products

import (
	"context"
	"github.com/solomonczyk/izborator/internal/logger"
	"time"
)

// Storage интерфейс для работы с хранилищем товаров
type Storage interface {
	// GetProduct получает товар по ID
	GetProduct(id string) (*Product, error)

	// SearchProducts ищет товары по запросу
	SearchProducts(query string, limit, offset int) ([]*Product, int, error)

	// Browse возвращает каталог товаров с фильтрами
	Browse(ctx context.Context, params BrowseParams) (*BrowseResult, error)

	ListBrands(ctx context.Context, productType string) ([]string, error)

	// SaveProduct сохраняет товар
	SaveProduct(product *Product) error

	// GetProductPrices получает цены товара из разных магазинов
	GetProductPrices(productID string) ([]*ProductPrice, error)

	// SaveProductPrice сохраняет цену товара
	SaveProductPrice(price *ProductPrice) error

	// GetURLsForRescrape возвращает список URL и ID магазинов для товаров,
	// цена которых не обновлялась дольше указанного времени
	GetURLsForRescrape(ctx context.Context, olderThan time.Duration, limit int) ([]RescrapeItem, error)
}

// RescrapeItem элемент для перескрапинга
type RescrapeItem struct {
	URL    string
	ShopID string
}

// Service сервис для работы с товарами
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис товаров
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}

// GetURLsForRescrape возвращает список URL для перескрапинга
func (s *Service) GetURLsForRescrape(ctx context.Context, olderThan time.Duration, limit int) ([]RescrapeItem, error) {
	return s.storage.GetURLsForRescrape(ctx, olderThan, limit)
}
