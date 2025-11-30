package attributes

import (
	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с хранилищем атрибутов
type Storage interface {
	// GetByID получает атрибут по ID
	GetByID(id string) (*Attribute, error)
	
	// GetByCode получает атрибут по коду
	GetByCode(code string) (*Attribute, error)
	
	// GetAllActive получает все активные атрибуты
	GetAllActive() ([]*Attribute, error)
	
	// GetByProductTypeID получает атрибуты для типа товара
	GetByProductTypeID(productTypeID string) ([]*ProductTypeAttribute, error)
}

// Service сервис для работы с атрибутами
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис атрибутов
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}

// GetByID получает атрибут по ID
func (s *Service) GetByID(id string) (*Attribute, error) {
	return s.storage.GetByID(id)
}

// GetByCode получает атрибут по коду
func (s *Service) GetByCode(code string) (*Attribute, error) {
	return s.storage.GetByCode(code)
}

// GetAllActive получает все активные атрибуты
func (s *Service) GetAllActive() ([]*Attribute, error) {
	return s.storage.GetAllActive()
}

// GetByProductTypeID получает атрибуты для типа товара
func (s *Service) GetByProductTypeID(productTypeID string) ([]*ProductTypeAttribute, error) {
	return s.storage.GetByProductTypeID(productTypeID)
}


