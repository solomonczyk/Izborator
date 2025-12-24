package scraper

import (
	"testing"
)

func TestCleanPrice(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPrice   float64
		wantCurr    string
		wantErr     bool
	}{
		{
			name:      "standard RSD price with dots",
			input:     "1.234,56 RSD",
			wantPrice: 1234.56,
			wantCurr:  "RSD",
			wantErr:   false,
		},
		{
			name:      "price with multiple dots",
			input:     "12.345,67 RSD",
			wantPrice: 12345.67,
			wantCurr:  "RSD",
			wantErr:   false,
		},
		{
			name:      "price with spaces",
			input:     "1 234,56 RSD",
			wantPrice: 1234.56,
			wantCurr:  "RSD",
			wantErr:   false,
		},
		{
			name:      "price with DIN",
			input:     "1.234,56 DIN",
			wantPrice: 1234.56,
			wantCurr:  "RSD",
			wantErr:   false,
		},
		{
			name:      "price without currency",
			input:     "1234.56",
			wantPrice: 1234.56,
			wantCurr:  "RSD",
			wantErr:   false,
		},
		{
			name:      "price with lowercase",
			input:     "1.234,56 rsd",
			wantPrice: 1234.56,
			wantCurr:  "RSD",
			wantErr:   false,
		},
		{
			name:      "invalid price",
			input:     "abc",
			wantPrice: 0,
			wantCurr:  "",
			wantErr:   true,
		},
		{
			name:      "empty string",
			input:     "",
			wantPrice: 0,
			wantCurr:  "",
			wantErr:   true,
		},
		{
			name:      "price with text before",
			input:     "Cena: 1.234,56 RSD",
			wantPrice: 1234.56,
			wantCurr:  "RSD",
			wantErr:   false,
		},
		{
			name:      "price with EUR",
			input:     "123.45 EUR",
			wantPrice: 123.45,
			wantCurr:  "EUR",
			wantErr:   false,
		},
		{
			name:      "price with USD",
			input:     "99.99 USD",
			wantPrice: 99.99,
			wantCurr:  "USD",
			wantErr:   false,
		},
		{
			name:      "price with zero",
			input:     "0 RSD",
			wantPrice: 0,
			wantCurr:  "RSD",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrice, gotCurr, err := cleanPrice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrice != tt.wantPrice {
				t.Errorf("cleanPrice() price = %v, want %v", gotPrice, tt.wantPrice)
			}
			if gotCurr != tt.wantCurr {
				t.Errorf("cleanPrice() currency = %v, want %v", gotCurr, tt.wantCurr)
			}
		})
	}
}

func TestIsProductURL(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name      string
		url       string
		shopConfig *ShopConfig
		want      bool
	}{
		{
			name: "product URL",
			url:  "https://example.com/product/iphone-15",
			shopConfig: &ShopConfig{
				BaseURL: "https://example.com",
			},
			want: true,
		},
		{
			name: "category URL",
			url:  "https://example.com/category/phones",
			shopConfig: &ShopConfig{
				BaseURL: "https://example.com",
			},
			want: false,
		},
		{
			name: "search URL",
			url:  "https://example.com/search?q=iphone",
			shopConfig: &ShopConfig{
				BaseURL: "https://example.com",
			},
			want: false,
		},
		{
			name: "gigatron product URL",
			url:  "https://gigatron.rs/kategorija/mobilni-telefoni/iphone-15",
			shopConfig: &ShopConfig{
				BaseURL: "https://gigatron.rs",
			},
			want: true,
		},
		{
			name: "gigatron base URL",
			url:  "https://gigatron.rs",
			shopConfig: &ShopConfig{
				BaseURL: "https://gigatron.rs",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isProductURL(tt.url, tt.shopConfig)
			if got != tt.want {
				t.Errorf("isProductURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
