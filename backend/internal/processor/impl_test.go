package processor

import (
	"context"
	"testing"

	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/scraper"
)

// MockStorage мок для тестирования
type mockRawStorage struct {
	rawProducts []*scraper.RawProduct
}

func (m *mockRawStorage) GetUnprocessedRawProducts(ctx context.Context, limit int) ([]*scraper.RawProduct, error) {
	if limit > len(m.rawProducts) {
		limit = len(m.rawProducts)
	}
	return m.rawProducts[:limit], nil
}

func (m *mockRawStorage) MarkAsProcessed(ctx context.Context, shopID, externalID string) error {
	return nil
}

type mockProcessedStorage struct {
	products []*products.Product
}

func (m *mockProcessedStorage) SaveProduct(product *products.Product) error {
	m.products = append(m.products, product)
	return nil
}

func (m *mockProcessedStorage) GetProduct(id string) (*products.Product, error) {
	for _, p := range m.products {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, products.ErrProductNotFound
}

type mockMatching struct {
	similarProducts []*products.Product
}

func (m *mockMatching) FindSimilar(ctx context.Context, name, brand string, limit int) ([]*products.Product, error) {
	return m.similarProducts, nil
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
	matching := &mockMatching{similarProducts: []*products.Product{}}

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
	matching := &mockMatching{similarProducts: []*products.Product{existingProduct}}

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
