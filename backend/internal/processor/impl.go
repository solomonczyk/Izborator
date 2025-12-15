package processor

import (
	"context"
	"fmt"
	"strings"

	"github.com/solomonczyk/izborator/internal/matching"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/scraper"
)

// ProcessRawProducts обрабатывает необработанные сырые данные
func (s *Service) ProcessRawProducts(ctx context.Context, batchSize int) (int, error) {
	// Проверка и установка значения по умолчанию
	if batchSize <= 0 {
		batchSize = 10
	}
	// Защита от перегрузки - ограничиваем максимальный размер батча
	if batchSize > 100 {
		batchSize = 100
	}

	s.logger.Info("processor: loading raw products", map[string]interface{}{
		"batch_size": batchSize,
	})

	// Получаем необработанные записи
	rawProducts, err := s.rawStorage.GetUnprocessedRawProducts(batchSize)
	if err != nil {
		s.logger.Error("processor: failed to load raw products", map[string]interface{}{
			"error": err.Error(),
		})
		return 0, fmt.Errorf("failed to get unprocessed raw products: %w", err)
	}

	if len(rawProducts) == 0 {
		s.logger.Info("processor: no raw products to process", map[string]interface{}{})
		return 0, nil
	}

	s.logger.Info("processor: starting batch processing", map[string]interface{}{
		"count": len(rawProducts),
	})

	processedCount := 0

	for _, raw := range rawProducts {
		// Safety: чтобы один проблемный товар не валил весь батч
		if err := s.processRawProduct(ctx, raw); err != nil {
			s.logger.Error("processor: failed to process raw product", map[string]interface{}{
				"shop_id":     raw.ShopID,
				"external_id": raw.ExternalID,
				"name":        raw.Name,
				"error":       err.Error(),
			})
			// Продолжаем обработку других товаров
			continue
		}

		// Помечаем как обработанный
		if err := s.rawStorage.MarkRawProductAsProcessed(raw.ShopID, raw.ExternalID); err != nil {
			s.logger.Error("processor: failed to mark raw product as processed", map[string]interface{}{
				"shop_id":     raw.ShopID,
				"external_id": raw.ExternalID,
				"error":       err.Error(),
			})
			// Не считаем это критической ошибкой, продолжаем
		}

		processedCount++
	}

	s.logger.Info("processor: batch processed", map[string]interface{}{
		"processed": processedCount,
		"total":     len(rawProducts),
	})

	return processedCount, nil
}

// processRawProduct обрабатывает один сырой товар
func (s *Service) processRawProduct(ctx context.Context, raw *scraper.RawProduct) error {
	// Нормализуем данные
	normalized := s.normalizeRawProduct(raw)

	// 1. Ищем кандидатов через matching
	matchReq := &matching.MatchRequest{
		Name:  normalized.Name,
		Brand: normalized.Brand,
		Specs: normalized.Specs,
	}

	matchResult, err := s.matching.MatchProduct(matchReq)
	if err != nil {
		s.logger.Warn("processor: matching failed, creating new product", map[string]interface{}{
			"shop_id":     raw.ShopID,
			"external_id": raw.ExternalID,
			"name":        normalized.Name,
			"error":       err.Error(),
		})
		// Решение: создаём новый товар, если matching не сработал
		if err := s.createNewProduct(ctx, raw, normalized); err != nil {
			return err
		}
		// Сохраняем цену для нового товара
		return s.savePriceForProduct(normalized.ID, raw)
	}

	s.logger.Info("processor: matching result", map[string]interface{}{
		"name":         normalized.Name,
		"matches_count": matchResult.Count,
		"matches":      matchResult.Matches,
	})

	var (
		targetProductID string
		isNewProduct    bool
	)

	// 2. Выбираем лучший кандидат или создаём новый товар
	if matchResult.Count == 0 {
		// Нет кандидатов - создаём новый товар
		isNewProduct = true
		s.logger.Debug("processor: no matches found, creating new product", map[string]interface{}{
			"name": normalized.Name,
		})
	} else {
		// Есть кандидаты - выбираем лучший по similarity
		best := matchResult.Matches[0]
		
		// Проверяем точное совпадение (similarity >= 0.95) - это почти 100% уверенность
		if best.Similarity >= 0.95 {
			// Точное совпадение - используем существующий товар
			targetProductID = best.MatchedID
			isNewProduct = false
			s.logger.Info("processor: found exact match", map[string]interface{}{
				"matched_id": targetProductID,
				"similarity": best.Similarity,
				"name":       normalized.Name,
			})
		} else if best.Similarity >= 0.7 {
			// Высокая уверенность - используем существующий товар
			targetProductID = best.MatchedID
			isNewProduct = false
			s.logger.Debug("processor: found matching product", map[string]interface{}{
				"matched_id": targetProductID,
				"similarity": best.Similarity,
				"name":       normalized.Name,
			})
		} else {
			// Низкая уверенность - создаём новый товар
			isNewProduct = true
			s.logger.Debug("processor: similarity too low, creating new product", map[string]interface{}{
				"similarity": best.Similarity,
				"name":       normalized.Name,
			})
		}
	}

	// 3. Если товара ещё нет - создаём новый Product
	if isNewProduct {
		if err := s.createNewProduct(ctx, raw, normalized); err != nil {
			return err
		}
		targetProductID = normalized.ID
	}

	// 4. Сохраняем цену для товара (нового или существующего)
	if err := s.savePriceForProduct(targetProductID, raw); err != nil {
		return fmt.Errorf("failed to save price: %w", err)
	}

	return nil
}

// normalizeRawProduct нормализует сырые данные товара
func (s *Service) normalizeRawProduct(raw *scraper.RawProduct) *products.Product {
	normalized := &products.Product{
		Name:        strings.TrimSpace(raw.Name),
		Description: strings.TrimSpace(raw.Description),
		Brand:       strings.TrimSpace(raw.Brand),
		Category:    strings.TrimSpace(raw.Category),
		Specs:       raw.Specs,
	}

	// Выбираем первое изображение как основное
	if len(raw.ImageURLs) > 0 {
		normalized.ImageURL = raw.ImageURLs[0]
	}

	// Нормализация бренда (убираем лишние пробелы, приводим к правильному регистру)
	if normalized.Brand != "" {
		normalized.Brand = s.normalizeBrand(normalized.Brand)
	}

	return normalized
}

// normalizeBrand нормализует название бренда
func (s *Service) normalizeBrand(brand string) string {
	brand = strings.TrimSpace(brand)
	if brand == "" {
		return ""
	}
	return strings.ToUpper(brand[:1]) + strings.ToLower(brand[1:])
}

// createNewProduct создаёт новый товар из сырых данных
func (s *Service) createNewProduct(ctx context.Context, raw *scraper.RawProduct, normalized *products.Product) error {
	// Сохраняем товар
	// ID будет сгенерирован на стороне storage (в ProcessorAdapter)
	if err := s.processedStorage.SaveProduct(normalized); err != nil {
		return fmt.Errorf("failed to save product: %w", err)
	}

	// Индексируем товар в Meilisearch
	if err := s.processedStorage.IndexProduct(normalized); err != nil {
		s.logger.Error("processor: failed to index product", map[string]interface{}{
			"product_id": normalized.ID,
			"error":      err.Error(),
		})
		// Не критично - продолжаем работу
	}

	s.logger.Info("processor: created new product", map[string]interface{}{
		"product_id": normalized.ID,
		"name":       normalized.Name,
		"brand":      normalized.Brand,
		"shop_id":    raw.ShopID,
	})

	return nil
}

// savePriceForProduct сохраняет цену товара
func (s *Service) savePriceForProduct(productID string, raw *scraper.RawProduct) error {
	price := &products.ProductPrice{
		ProductID: productID,
		ShopID:    raw.ShopID,
		ShopName:  raw.ShopName,
		Price:     raw.Price,
		Currency:  raw.Currency,
		URL:       raw.URL,
		InStock:   raw.InStock,
	}

	if err := s.processedStorage.SavePrice(price); err != nil {
		return fmt.Errorf("failed to save price: %w", err)
	}

	s.logger.Debug("Saved product price", map[string]interface{}{
		"product_id": productID,
		"shop_id":    raw.ShopID,
		"price":      raw.Price,
		"currency":   raw.Currency,
	})

	return nil
}

