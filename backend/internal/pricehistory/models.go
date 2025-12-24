package pricehistory

import "time"

// PricePoint точка цены во времени
type PricePoint struct {
	ProductID string    `json:"product_id"`
	ShopID    string    `json:"shop_id"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	Timestamp time.Time `json:"timestamp"`
}

// PriceHistory история цен товара
type PriceHistory struct {
	ProductID string        `json:"product_id"`
	Points    []*PricePoint `json:"points"`
	Period    string        `json:"period"` // "day", "week", "month", "year"
}

// PriceChart данные для графика цен
type PriceChart struct {
	ProductID string                   `json:"product_id"`
	Shops     map[string][]*PricePoint `json:"shops"`      // shop_id -> points
	ShopNames map[string]string        `json:"shop_names"` // shop_id -> shop_name
	Period    string                   `json:"period"`
	From      time.Time                `json:"from"`
	To        time.Time                `json:"to"`
}
