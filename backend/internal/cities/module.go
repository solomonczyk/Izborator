package cities

import (
	"github.com/solomonczyk/izborator/internal/logger"
)

// Storage интерфейс для работы с хранилищем городов
type Storage interface {
	// GetByID получает город по ID
	GetByID(id string) (*City, error)
	
	// GetBySlug получает город по slug
	GetBySlug(slug string) (*City, error)
	
	// GetAllActive получает все активные города
	GetAllActive() ([]*City, error)
}

// Service сервис для работы с городами
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис городов
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}

// GetByID получает город по ID
func (s *Service) GetByID(id string) (*City, error) {
	return s.storage.GetByID(id)
}

// GetBySlug получает город по slug
func (s *Service) GetBySlug(slug string) (*City, error) {
	return s.storage.GetBySlug(slug)
}

// GetAllActive получает все активные города
func (s *Service) GetAllActive() ([]*City, error) {
	return s.storage.GetAllActive()
}


