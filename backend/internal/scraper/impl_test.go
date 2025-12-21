package scraper

import (
	"testing"
)

// TestCleanPrice тестирует функцию cleanPrice для парсинга цен
func TestCleanPrice(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  float64
		currency  string
		hasError  bool
	}{
		{
			name:     "standard RSD price with dots",
			input:    "15.999 RSD",
			expected: 15999.0,
			currency: "RSD",
			hasError: false,
		},
		{
			name:     "price with multiple dots",
			input:    "1.000.000 RSD",
			expected: 1000000.0,
			currency: "RSD",
			hasError: false,
		},
		{
			name:     "price with spaces",
			input:    "  16.999 RSD  ",
			expected: 16999.0,
			currency: "RSD",
			hasError: false,
		},
		{
			name:     "price with DIN",
			input:    "15.999 DIN",
			expected: 15999.0,
			currency: "RSD",
			hasError: false,
		},
		{
			name:     "price without currency",
			input:    "15999",
			expected: 15999.0,
			currency: "RSD",
			hasError: false,
		},
		{
			name:     "price with lowercase",
			input:    "15.999 rsd",
			expected: 15999.0,
			currency: "RSD",
			hasError: false,
		},
		{
			name:     "invalid price",
			input:    "invalid",
			expected: 0,
			currency: "",
			hasError: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			currency: "",
			hasError: true,
		},
		{
			name:     "price with text before",
			input:    "Cena: 15.999 RSD",
			expected: 15999.0,
			currency: "RSD",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, currency, err := cleanPrice(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("cleanPrice(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("cleanPrice(%q) unexpected error: %v", tt.input, err)
				}
				if price != tt.expected {
					t.Errorf("cleanPrice(%q) price = %f, want %f", tt.input, price, tt.expected)
				}
				if currency != tt.currency {
					t.Errorf("cleanPrice(%q) currency = %q, want %q", tt.input, currency, tt.currency)
				}
			}
		})
	}
}

