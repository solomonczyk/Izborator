package products

import "time"

// Product каноническая карточка товара
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Brand       string    `json:"brand"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"image_url"`
	Specs       map[string]string `json:"specs"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductPrice цена товара в конкретном магазине
type ProductPrice struct {
	ProductID string    `json:"product_id"`
	ShopID    string    `json:"shop_id"`
	ShopName  string    `json:"shop_name"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	URL       string    `json:"url"`
	InStock   bool      `json:"in_stock"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchResult результат поиска товаров
type SearchResult struct {
	Items []*Product `json:"items"`
	Total int        `json:"total"`
	Limit int        `json:"limit"`
	Offset int       `json:"offset"`
}

