package classifier

import (
	"github.com/solomonczyk/izborator/internal/logger"
)

// ClassificationScore результат классификации
type ClassificationScore struct {
	KeywordsScore  float64 // 0.0 - 1.0 (наличие ключевых слов)
	PlatformScore  float64 // 0.0 - 1.0 (обнаружение платформы)
	StructureScore float64 // 0.0 - 1.0 (структура страницы)
	TotalScore     float64 // Weighted average
}

// ClassificationResult результат классификации домена
type ClassificationResult struct {
	IsShop           bool   // Является ли сайт магазином (e-commerce)
	IsService        bool   // Является ли сайт провайдером услуг
	SiteType         string // "ecommerce" | "service_provider" | "unknown"
	Score            ClassificationScore
	DetectedPlatform string   // Обнаруженная платформа (shopify, woocommerce, etc.)
	Reasons          []string // Причины решения
}

// PotentialShop кандидат на магазин
type PotentialShop struct {
	ID              string
	Domain          string
	Source          string
	Status          string
	ConfidenceScore float64
	DiscoveredAt    string
	Metadata        map[string]interface{}
}

// Storage интерфейс для работы с хранилищем
type Storage interface {
	// SavePotentialShop сохраняет кандидата на магазин
	SavePotentialShop(shop *PotentialShop) error

	// GetPotentialShopByDomain получает кандидата по домену
	GetPotentialShopByDomain(domain string) (*PotentialShop, error)

	// ListPotentialShopsByStatus получает список кандидатов по статусу
	ListPotentialShopsByStatus(status string, limit int) ([]*PotentialShop, error)

	// UpdatePotentialShop обновляет кандидата
	UpdatePotentialShop(shop *PotentialShop) error
}

// Service сервис для классификации магазинов
type Service struct {
	storage Storage
	logger  *logger.Logger
}

// New создаёт новый сервис классификации
func New(storage Storage, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  log,
	}
}
