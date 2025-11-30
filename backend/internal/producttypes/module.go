package producttypes

import (
	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с хранилищем типов товаров
type Storage interface {
	// GetByID получает тип товара по ID
	GetByID(id string) (*ProductType, error)
	
	// GetByCode получает тип товара по коду
	GetByCode(code string) (*ProductType, error)
	
	// GetAllActive получает все активные типы товаров
	GetAllActive() ([]*ProductType, error)
	
	// GetByCategoryID получает типы товаров для категории
	GetByCategoryID(categoryID string) ([]*ProductType, error)
}

// Service сервис для работы с типами товаров
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис типов товаров
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}

// GetByID получает тип товара по ID
func (s *Service) GetByID(id string) (*ProductType, error) {
	return s.storage.GetByID(id)
}

// GetByCode получает тип товара по коду
func (s *Service) GetByCode(code string) (*ProductType, error) {
	return s.storage.GetByCode(code)
}

// GetAllActive получает все активные типы товаров
func (s *Service) GetAllActive() ([]*ProductType, error) {
	return s.storage.GetAllActive()
}

// GetByCategoryID получает типы товаров для категории
func (s *Service) GetByCategoryID(categoryID string) ([]*ProductType, error) {
	return s.storage.GetByCategoryID(categoryID)
}


