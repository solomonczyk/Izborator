package products

import (
	"context"
	"fmt"
)

// Search ищет товары по запросу (простой поиск без пагинации для /api/v1/products/search)
func (s *Service) Search(ctx context.Context, query string) ([]*Product, error) {
	if query == "" {
		return nil, ErrInvalidSearchQuery
	}

	products, _, err := s.storage.SearchProducts(query, 20, 0)
	if err != nil {
		s.logger.Error("Failed to search products", map[string]interface{}{
			"error": err,
			"query": query,
		})
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return products, nil
}

// SearchWithPagination ищет товары по запросу с пагинацией (старый формат)
func (s *Service) SearchWithPagination(ctx context.Context, query string, limit, offset int) (*SearchResult, error) {
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

// Browse возвращает каталог товаров с фильтрами
func (s *Service) Browse(ctx context.Context, params BrowseParams) (*BrowseResult, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PerPage <= 0 {
		params.PerPage = 20
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	result, err := s.storage.Browse(ctx, params)
	if err != nil {
		s.logger.Error("Failed to browse products", map[string]interface{}{
			"error": err,
			"query": params.Query,
		})
		return nil, fmt.Errorf("browse failed: %w", err)
	}

	return result, nil
}

// GetByID получает товар по ID
func (s *Service) GetByID(id string) (*Product, error) {
	if id == "" {
		return nil, ErrInvalidProductID
	}

	product, err := s.storage.GetProduct(id)
	if err != nil {
		s.logger.Error("Failed to get product", map[string]interface{}{
			"error":      err,
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
			"error":      err,
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
			"error":      err,
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
			"error":      err,
			"product_id": price.ProductID,
			"shop_id":    price.ShopID,
		})
		return fmt.Errorf("failed to save price: %w", err)
	}

	return nil
}
