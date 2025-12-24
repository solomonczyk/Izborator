package processor

import (
	"context"
	"testing"

	"github.com/solomonczyk/izborator/internal/matching"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/scraper"
)

// MockStorage мок для тестирования
type mockRawStorage struct {
	rawProducts []*scraper.RawProduct
}

func (m *mockRawStorage) GetUnprocessedRawProducts(limit int) ([]*scraper.RawProduct, error) {
	if limit > len(m.rawProducts) {
		limit = len(m.rawProducts)
	}
	return m.rawProducts[:limit], nil
}

func (m *mockRawStorage) MarkRawProductAsProcessed(shopID, externalID string) error {
	return nil
}

type mockProcessedStorage struct {
	products []*products.Product
}

func (m *mockProcessedStorage) SaveProduct(product *products.Product) error {
	m.products = append(m.products, product)
	return nil
}

func (m *mockProcessedStorage) SavePrice(price *products.ProductPrice) error {
	return nil
}

func (m *mockProcessedStorage) IndexProduct(product *products.Product) error {
	return nil
}

type mockMatching struct {
	matchResult *matching.MatchResult
}

func (m *mockMatching) MatchProduct(req *matching.MatchRequest) (*matching.MatchResult, error) {
	if m.matchResult != nil {
		return m.matchResult, nil
	}
	return &matching.MatchResult{
		Matches: []*matching.ProductMatch{},
		Count:   0,
	}, nil
}

func TestNormalizeBrand(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal brand", "apple", "Apple"},
		{"uppercase brand", "APPLE", "Apple"},
		{"mixed case brand", "ApPlE", "Apple"},
		{"brand with spaces", "  apple  ", "Apple"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
		{"single character", "a", "A"},
		{"two characters", "ab", "Ab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.normalizeBrand(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeBrand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeRawProduct(t *testing.T) {
	service := &Service{}

	raw := &scraper.RawProduct{
		Name:     "  iPhone 15 Pro Max  ",
		Brand:    "  apple  ",
		Category: "Mobilni telefoni",
		Specs: map[string]string{
			"storage": "256GB",
		},
		ImageURLs: []string{"https://example.com/image.jpg"},
	}

	normalized := service.normalizeRawProduct(raw)

	if normalized.Name != "iPhone 15 Pro Max" {
		t.Errorf("Name not trimmed: got %q, want %q", normalized.Name, "iPhone 15 Pro Max")
	}

	if normalized.Brand != "Apple" {
		t.Errorf("Brand not normalized: got %q, want %q", normalized.Brand, "Apple")
	}

	if normalized.ImageURL != "https://example.com/image.jpg" {
		t.Errorf("ImageURL not set: got %q, want %q", normalized.ImageURL, "https://example.com/image.jpg")
	}

	if len(normalized.Specs) != 1 {
		t.Errorf("Specs not copied: got %d, want 1", len(normalized.Specs))
	}
}

func TestProcessRawProducts_NoMatches(t *testing.T) {
	rawStorage := &mockRawStorage{
		rawProducts: []*scraper.RawProduct{
			{
				Name:     "Test Product",
				Brand:    "Test Brand",
				Category: "Test Category",
				Price:    100.0,
				Currency: "RSD",
			},
		},
	}
	processedStorage := &mockProcessedStorage{}
	matching := &mockMatching{
		matchResult: &matching.MatchResult{
			Matches: []*matching.ProductMatch{},
			Count:   0,
		},
	}

	service := New(rawStorage, processedStorage, matching, nil)

	ctx := context.Background()
	count, err := service.ProcessRawProducts(ctx, 10)

	if err != nil {
		t.Fatalf("ProcessRawProducts failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 product processed, got %d", count)
	}

	if len(processedStorage.products) != 1 {
		t.Errorf("Expected 1 product saved, got %d", len(processedStorage.products))
	}
}

func TestProcessRawProducts_WithMatches(t *testing.T) {
	rawStorage := &mockRawStorage{
		rawProducts: []*scraper.RawProduct{
			{
				Name:     "iPhone 15",
				Brand:    "Apple",
				Category: "Mobilni telefoni",
				Price:    1000.0,
				Currency: "RSD",
			},
		},
	}
	processedStorage := &mockProcessedStorage{}
	existingProduct := &products.Product{
		ID:    "existing-id",
		Name:  "iPhone 15",
		Brand: "Apple",
	}
	matching := &mockMatching{
		matchResult: &matching.MatchResult{
			Matches: []*matching.ProductMatch{
				{
					MatchedID: existingProduct.ID,
					Similarity: 0.95,
				},
			},
			Count: 1,
		},
	}

	service := New(rawStorage, processedStorage, matching, nil)

	ctx := context.Background()
	count, err := service.ProcessRawProducts(ctx, 10)

	if err != nil {
		t.Fatalf("ProcessRawProducts failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 product processed, got %d", count)
	}
}

func TestProcessRawProducts_EmptyBatch(t *testing.T) {
	rawStorage := &mockRawStorage{
		rawProducts: []*scraper.RawProduct{},
	}
	processedStorage := &mockProcessedStorage{}
	matching := &mockMatching{}

	service := New(rawStorage, processedStorage, matching, nil)

	ctx := context.Background()
	count, err := service.ProcessRawProducts(ctx, 10)

	if err != nil {
		t.Fatalf("ProcessRawProducts failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 products processed, got %d", count)
	}
}

func TestProcessRawProducts_BatchSizeLimit(t *testing.T) {
	rawStorage := &mockRawStorage{
		rawProducts: make([]*scraper.RawProduct, 150), // Больше лимита
	}
	for i := range rawStorage.rawProducts {
		rawStorage.rawProducts[i] = &scraper.RawProduct{
			Name:     "Test Product",
			Brand:    "Test Brand",
			Category: "Test Category",
			Price:    100.0,
			Currency: "RSD",
		}
	}
	processedStorage := &mockProcessedStorage{}
	matching := &mockMatching{}

	service := New(rawStorage, processedStorage, matching, nil)

	ctx := context.Background()
	// Запрос с batchSize > 100 должен быть ограничен до 100
	count, err := service.ProcessRawProducts(ctx, 150)

	if err != nil {
		t.Fatalf("ProcessRawProducts failed: %v", err)
	}

	// Должно обработать максимум 100 товаров
	if count > 100 {
		t.Errorf("Expected max 100 products processed, got %d", count)
	}
}

func TestProcessRawProducts_InvalidBatchSize(t *testing.T) {
	rawStorage := &mockRawStorage{
		rawProducts: []*scraper.RawProduct{
			{
				Name:     "Test Product",
				Brand:    "Test Brand",
				Category: "Test Category",
				Price:    100.0,
				Currency: "RSD",
			},
		},
	}
	processedStorage := &mockProcessedStorage{}
	matching := &mockMatching{}

	service := New(rawStorage, processedStorage, matching, nil)

	ctx := context.Background()
	// Отрицательный batchSize должен быть заменен на 10
	count, err := service.ProcessRawProducts(ctx, -5)

	if err != nil {
		t.Fatalf("ProcessRawProducts failed: %v", err)
	}

	// Должно обработать товар (batchSize заменен на 10)
	if count != 1 {
		t.Errorf("Expected 1 product processed, got %d", count)
	}
}
