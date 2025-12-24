package pricehistory

import (
	"fmt"
	"time"
)

// SavePrice сохраняет цену товара
func (s *Service) SavePrice(productID, shopID string, price float64, currency string) error {
	if productID == "" {
		return fmt.Errorf("product ID is required")
	}
	if shopID == "" {
		return fmt.Errorf("shop ID is required")
	}
	if price < 0 {
		return fmt.Errorf("price cannot be negative")
	}

	point := &PricePoint{
		ProductID: productID,
		ShopID:    shopID,
		Price:     price,
		Currency:  currency,
		Timestamp: time.Now(),
	}

	if err := s.storage.SavePrice(point); err != nil {
		s.logger.Error("Failed to save price", map[string]interface{}{
			"error":      err,
			"product_id": productID,
			"shop_id":    shopID,
		})
		return fmt.Errorf("failed to save price: %w", err)
	}

	return nil
}

// GetHistory получает историю цен товара
func (s *Service) GetHistory(productID string, from, to time.Time) (*PriceHistory, error) {
	if productID == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	if from.After(to) {
		return nil, ErrInvalidTimeRange
	}

	points, err := s.storage.GetHistory(productID, from, to)
	if err != nil {
		s.logger.Error("Failed to get price history", map[string]interface{}{
			"error":      err,
			"product_id": productID,
		})
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	return &PriceHistory{
		ProductID: productID,
		Points:    points,
		Period:    calculatePeriod(from, to),
	}, nil
}

// GetPriceChart получает данные для графика цен
func (s *Service) GetPriceChart(productID string, period string, shopIDs []string) (*PriceChart, error) {
	if productID == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	validPeriods := map[string]bool{
		"day":   true,
		"week":  true,
		"month": true,
		"year":  true,
	}

	if !validPeriods[period] {
		return nil, ErrInvalidPeriod
	}

	chart, err := s.storage.GetPriceChart(productID, period, shopIDs)
	if err != nil {
		s.logger.Error("Failed to get price chart", map[string]interface{}{
			"error":      err,
			"product_id": productID,
			"period":     period,
		})
		return nil, fmt.Errorf("failed to get chart: %w", err)
	}

	return chart, nil
}

// calculatePeriod определяет период на основе временного диапазона
func calculatePeriod(from, to time.Time) string {
	duration := to.Sub(from)

	if duration <= 24*time.Hour {
		return "day"
	} else if duration <= 7*24*time.Hour {
		return "week"
	} else if duration <= 30*24*time.Hour {
		return "month"
	}
	return "year"
}
