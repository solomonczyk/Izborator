package processor

import (
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/matching"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/scraper"
	"github.com/solomonczyk/izborator/internal/semantic"
)

// RawStorage интерфейс для чтения сырых данных
type RawStorage interface {
	GetUnprocessedRawProducts(limit int) ([]*scraper.RawProduct, error)
	MarkRawProductAsProcessed(shopID, externalID string) error
	GetShopDefaultCityID(shopID string) (*string, error)
}

// ProcessedStorage интерфейс для записи обработанных данных
type ProcessedStorage interface {
	SaveProduct(product *products.Product) error
	SavePrice(price *products.ProductPrice) error
	IndexProduct(product *products.Product) error // Индексация в Meilisearch
}

// Matching интерфейс для сопоставления товаров
type Matching interface {
	MatchProduct(req *matching.MatchRequest) (*matching.MatchResult, error)
}

type SemanticValidationRecorder interface {
	RecordSemanticValidation(result semantic.SemanticValidationResult)
}

// Service сервис для обработки сырых данных
type Service struct {
	rawStorage       RawStorage
	processedStorage ProcessedStorage
	matching         Matching
	semanticRecorder SemanticValidationRecorder
	logger           *logger.Logger
}

// New создаёт новый сервис обработки
func New(
	rawStorage RawStorage,
	processedStorage ProcessedStorage,
	matching Matching,
	semanticRecorder SemanticValidationRecorder,
	log *logger.Logger,
) *Service {
	if log == nil {
		log = logger.New("info")
	}
	return &Service{
		rawStorage:       rawStorage,
		processedStorage: processedStorage,
		matching:         matching,
		semanticRecorder: semanticRecorder,
		logger:           log,
	}
}
