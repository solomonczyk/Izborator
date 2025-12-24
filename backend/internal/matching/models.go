package matching

import "time"

// ProductMatch результат сопоставления товаров
type ProductMatch struct {
	ProductID  string    `json:"product_id"`
	MatchedID  string    `json:"matched_id"`
	Similarity float64   `json:"similarity"` // 0.0 - 1.0
	MatchedAt  time.Time `json:"matched_at"`
	Confidence string    `json:"confidence"` // "high", "medium", "low"
}

// MatchRequest запрос на сопоставление товара
type MatchRequest struct {
	ProductID string            `json:"product_id"`
	Name      string            `json:"name"`
	Brand     string            `json:"brand"`
	Specs     map[string]string `json:"specs"`
}

// MatchResult результат поиска похожих товаров
type MatchResult struct {
	Matches []*ProductMatch `json:"matches"`
	Count   int             `json:"count"`
}
