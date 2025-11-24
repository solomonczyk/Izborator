package products

import (
	"fmt"
)

// Search ищет товары по запросу
func (s *Service) Search(query string, limit, offset int) (*SearchResult, error) {
	if query == "" {
		return nil, ErrInvalidSearchQuery
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	products, total, err := s.storage.SearchProducts(query, limit, offset)
	if err != nil {
		s.logger.Error("Failed to search products", map[string]interface{}{
			"error": err,
			"query": query,
		})
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return &SearchResult{
		Items:  products,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// GetByID получает товар по ID
func (s *Service) GetByID(id string) (*Product, error) {
	if id == "" {
		return nil, ErrInvalidProductID
	}

	product, err := s.storage.GetProduct(id)
	if err != nil {
		s.logger.Error("Failed to get product", map[string]interface{}{
			"error":    err,
			"product_id": id,
		})
		return nil, ErrProductNotFound
	}

	return product, nil
}

// GetPrices получает цены товара из разных магазинов
func (s *Service) GetPrices(productID string) ([]*ProductPrice, error) {
	if productID == "" {
		return nil, ErrInvalidProductID
	}

	prices, err := s.storage.GetProductPrices(productID)
	if err != nil {
		s.logger.Error("Failed to get product prices", map[string]interface{}{
			"error":     err,
			"product_id": productID,
		})
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	return prices, nil
}

// SaveProduct сохраняет товар
func (s *Service) SaveProduct(product *Product) error {
	if product == nil {
		return fmt.Errorf("product is nil")
	}

	if err := s.storage.SaveProduct(product); err != nil {
		s.logger.Error("Failed to save product", map[string]interface{}{
			"error": err,
			"product_id": product.ID,
		})
		return fmt.Errorf("failed to save product: %w", err)
	}

	return nil
}

// SavePrice сохраняет цену товара
func (s *Service) SavePrice(price *ProductPrice) error {
	if price == nil {
		return fmt.Errorf("price is nil")
	}

	if err := s.storage.SaveProductPrice(price); err != nil {
		s.logger.Error("Failed to save product price", map[string]interface{}{
			"error":     err,
			"product_id": price.ProductID,
			"shop_id":    price.ShopID,
		})
		return fmt.Errorf("failed to save price: %w", err)
	}

	return nil
}
