package processor

import (
	"testing"

	"github.com/solomonczyk/izborator/internal/scraper"
)

// TestNormalizeBrand тестирует нормализацию бренда
func TestNormalizeBrand(t *testing.T) {
	// Создаём минимальный сервис для теста
	service := &Service{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal brand",
			input:    "apple",
			expected: "Apple",
		},
		{
			name:     "uppercase brand",
			input:    "APPLE",
			expected: "Apple",
		},
		{
			name:     "mixed case brand",
			input:    "ApPlE",
			expected: "Apple",
		},
		{
			name:     "brand with spaces",
			input:    "  apple  ",
			expected: "Apple",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
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

// TestNormalizeRawProduct тестирует нормализацию сырого товара
func TestNormalizeRawProduct(t *testing.T) {
	service := &Service{}

	// Тест с минимальными данными
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

