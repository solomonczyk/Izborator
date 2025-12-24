package categories

import (
	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с хранилищем категорий
type Storage interface {
	// GetByID получает категорию по ID
	GetByID(id string) (*Category, error)

	// GetBySlug получает категорию по slug
	GetBySlug(slug string) (*Category, error)

	// GetByParentID получает все подкатегории родительской категории
	GetByParentID(parentID string) ([]*Category, error)

	// GetAllActive получает все активные категории
	GetAllActive() ([]*Category, error)

	// GetTree получает дерево категорий (все корневые + их дети)
	GetTree() ([]*Category, error)
}

// Service сервис для работы с категориями
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис категорий
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}

// GetByID получает категорию по ID
func (s *Service) GetByID(id string) (*Category, error) {
	return s.storage.GetByID(id)
}

// GetBySlug получает категорию по slug
func (s *Service) GetBySlug(slug string) (*Category, error) {
	return s.storage.GetBySlug(slug)
}

// GetByParentID получает подкатегории
func (s *Service) GetByParentID(parentID string) ([]*Category, error) {
	return s.storage.GetByParentID(parentID)
}

// GetAllActive получает все активные категории
func (s *Service) GetAllActive() ([]*Category, error) {
	return s.storage.GetAllActive()
}

// GetTree получает дерево категорий
func (s *Service) GetTree() ([]*Category, error) {
	return s.storage.GetTree()
}
