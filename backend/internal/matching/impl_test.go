package matching

import (
	"testing"
)

// TestNormalizeName тестирует нормализацию названия товара
func TestNormalizeName(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal name",
			input:    "iPhone 15 Pro Max",
			expected: "iphone 15",
		},
		{
			name:     "uppercase name",
			input:    "IPHONE 15 PRO MAX",
			expected: "iphone 15",
		},
		{
			name:     "name with spaces",
			input:    "  iPhone 15 Pro Max  ",
			expected: "iphone 15",
		},
		{
			name:     "name with dashes",
			input:    "iPhone-15-Pro-Max",
			expected: "iphone 15",
		},
		{
			name:     "name with memory",
			input:    "iPhone 15 Pro Max 256GB",
			expected: "iphone 15 256",
		},
		{
			name:     "name with memory slash",
			input:    "iPhone 15 12/512GB",
			expected: "iphone 15 512",
		},
		{
			name:     "name with cyrillic",
			input:    "Смартфон iPhone 15",
			expected: "смартфон iphone 15",
		},
		{
			name:     "name with stop words",
			input:    "iPhone 15 Pro Max Black",
			expected: "iphone 15",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.normalizeName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeBrand тестирует нормализацию бренда
func TestNormalizeBrand(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal brand",
			input:    "Apple",
			expected: "apple",
		},
		{
			name:     "uppercase brand",
			input:    "APPLE",
			expected: "apple",
		},
		{
			name:     "brand with spaces",
			input:    "  Apple  ",
			expected: "apple",
		},
		{
			name:     "brand with dash",
			input:    "Samsung-Galaxy",
			expected: "samsunggalaxy",
		},
		{
			name:     "brand alias Samsung",
			input:    "SAMSUNG",
			expected: "samsung",
		},
		{
			name:     "brand alias Xiaomi",
			input:    "XIAOMI",
			expected: "xiaomi",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
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

// TestCalculateSimilarity тестирует расчёт схожести товаров
func TestCalculateSimilarity(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		req      *MatchRequest
		product  *Product
		expected float64
	}{
		{
			name: "exact match with brand",
			req: &MatchRequest{
				Name:  "iPhone 15 Pro Max",
				Brand: "Apple",
			},
			product: &Product{
				Name:  "iPhone 15 Pro Max",
				Brand: "Apple",
			},
			expected: 1.0, // Полное совпадение
		},
		{
			name: "exact match without brand",
			req: &MatchRequest{
				Name:  "iPhone 15 Pro Max",
				Brand: "",
			},
			product: &Product{
				Name:  "iPhone 15 Pro Max",
				Brand: "",
			},
			expected: 0.95, // Высокая схожесть без бренда
		},
		{
			name: "partial match",
			req: &MatchRequest{
				Name:  "iPhone 15 Pro Max 256GB",
				Brand: "Apple",
			},
			product: &Product{
				Name:  "iPhone 15 Pro Max",
				Brand: "Apple",
			},
			expected: 0.7, // Частичное совпадение (содержит подстроку + совпадение бренда)
		},
		{
			name: "no match",
			req: &MatchRequest{
				Name:  "iPhone 15 Pro Max",
				Brand: "Apple",
			},
			product: &Product{
				Name:  "Samsung Galaxy S24",
				Brand: "Samsung",
			},
			expected: 0.0, // Нет совпадения
		},
		{
			name: "match with different colors",
			req: &MatchRequest{
				Name:  "iPhone 15 Pro Max Black",
				Brand: "Apple",
			},
			product: &Product{
				Name:  "iPhone 15 Pro Max White",
				Brand: "Apple",
			},
			expected: 1.0, // Цвета фильтруются, должно быть полное совпадение
		},
		{
			name: "match with memory normalization",
			req: &MatchRequest{
				Name:  "iPhone 15 12/512GB",
				Brand: "Apple",
			},
			product: &Product{
				Name:  "iPhone 15 512GB",
				Brand: "Apple",
			},
			expected: 1.0, // Память нормализуется, должно совпадать
		},
		{
			name: "match with dash normalization",
			req: &MatchRequest{
				Name:  "iPhone-15-Pro-Max",
				Brand: "Apple",
			},
			product: &Product{
				Name:  "iPhone 15 Pro Max",
				Brand: "Apple",
			},
			expected: 1.0, // Тире нормализуются, должно совпадать
		},
		{
			name: "match with brand alias",
			req: &MatchRequest{
				Name:  "Galaxy S24",
				Brand: "SAMSUNG",
			},
			product: &Product{
				Name:  "Galaxy S24",
				Brand: "Samsung",
			},
			expected: 1.0, // Бренды нормализуются, должно совпадать
		},
		{
			name: "common words match",
			req: &MatchRequest{
				Name:  "Samsung Galaxy S24 Ultra",
				Brand: "Samsung",
			},
			product: &Product{
				Name:  "Samsung Galaxy S24",
				Brand: "Samsung",
			},
			expected: 0.85, // Много общих слов + совпадение бренда
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculateSimilarity(tt.req, tt.product)
			// Проверяем с небольшой погрешностью для float
			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("calculateSimilarity() = %f, want ~%f", result, tt.expected)
			}
		})
	}
}
