package matching

import (
	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с хранилищем товаров для matching
type Storage interface {
	// FindSimilarProducts ищет похожие товары или услуги
	FindSimilarProducts(name, brand string, productType string, limit int) ([]*Product, error)

	// GetProductByID получает товар по ID
	GetProductByID(id string) (*Product, error)

	// SaveMatch сохраняет результат сопоставления
	SaveMatch(match *ProductMatch) error

	// GetMatches получает все сопоставления для товара
	GetMatches(productID string) ([]*ProductMatch, error)
}

// Product используется из пакета products
// Импортируется через интерфейс или копируется структура
type Product struct {
	ID    string
	Name  string
	Brand string
	Specs map[string]string
	Type  string // "good" | "service"
}

// Service сервис для сопоставления товаров между магазинами
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис сопоставления
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}
