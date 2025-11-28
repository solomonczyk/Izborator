package products

import (
	"context"
	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с хранилищем товаров
type Storage interface {
	// GetProduct получает товар по ID
	GetProduct(id string) (*Product, error)
	
	// SearchProducts ищет товары по запросу
	SearchProducts(query string, limit, offset int) ([]*Product, int, error)
	
	// Browse возвращает каталог товаров с фильтрами
	Browse(ctx context.Context, params BrowseParams) (*BrowseResult, error)
	
	// SaveProduct сохраняет товар
	SaveProduct(product *Product) error
	
	// GetProductPrices получает цены товара из разных магазинов
	GetProductPrices(productID string) ([]*ProductPrice, error)
	
	// SaveProductPrice сохраняет цену товара
	SaveProductPrice(price *ProductPrice) error
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

