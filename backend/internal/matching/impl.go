package matching

import (
	"fmt"
	"strings"
)

// MatchProduct сопоставляет товар с существующими
func (s *Service) MatchProduct(req *MatchRequest) (*MatchResult, error) {
	if req == nil {
		return nil, ErrInsufficientData
	}

	if req.Name == "" {
		return nil, ErrInsufficientData
	}

	// Нормализуем данные для поиска
	normalizedName := s.normalizeName(req.Name)
	normalizedBrand := s.normalizeBrand(req.Brand)

	// Ищем похожие товары
	similar, err := s.storage.FindSimilarProducts(normalizedName, normalizedBrand, 10)
	if err != nil {
		s.logger.Error("Failed to find similar products", map[string]interface{}{
			"error": err,
		})
		return nil, fmt.Errorf("failed to find similar: %w", err)
	}

	// Рассчитываем схожесть для каждого найденного товара
	matches := make([]*ProductMatch, 0, len(similar))
	for _, product := range similar {
		similarity := s.calculateSimilarity(req, product)
		
		if similarity > 0.5 { // Порог схожести
			matches = append(matches, &ProductMatch{
				ProductID:  req.ProductID,
				MatchedID:  product.ID,
				Similarity: similarity,
			})
		}
	}

	return &MatchResult{
		Matches: matches,
		Count:   len(matches),
	}, nil
}

// normalizeName нормализует название товара для сравнения
func (s *Service) normalizeName(name string) string {
	// Приводим к нижнему регистру и убираем лишние пробелы
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)
	
	// TODO: добавить более сложную нормализацию
	// - удаление спецсимволов
	// - замена сокращений
	// - удаление артиклей
	
	return name
}

// normalizeBrand нормализует бренд
func (s *Service) normalizeBrand(brand string) string {
	if brand == "" {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(brand))
}

// calculateSimilarity рассчитывает схожесть между товарами
func (s *Service) calculateSimilarity(req *MatchRequest, product *Product) float64 {
	// Простой алгоритм схожести на основе названия и бренда
	// TODO: улучшить алгоритм (Levenshtein, Jaro-Winkler и т.д.)
	
	similarity := 0.0
	
	// Сравнение названий
	reqName := s.normalizeName(req.Name)
	prodName := s.normalizeName(product.Name)
	
	if reqName == prodName {
		similarity += 0.6
	} else if strings.Contains(reqName, prodName) || strings.Contains(prodName, reqName) {
		similarity += 0.4
	}
	
	// Сравнение брендов
	if req.Brand != "" && product.Brand != "" {
		reqBrand := s.normalizeBrand(req.Brand)
		prodBrand := s.normalizeBrand(product.Brand)
		
		if reqBrand == prodBrand {
			similarity += 0.4
		}
	}
	
	if similarity > 1.0 {
		similarity = 1.0
	}
	
	return similarity
}

// SaveMatch сохраняет результат сопоставления
func (s *Service) SaveMatch(match *ProductMatch) error {
	if match == nil {
		return fmt.Errorf("match is nil")
	}

	if match.Similarity < 0.0 || match.Similarity > 1.0 {
		return ErrInvalidSimilarity
	}

	if err := s.storage.SaveMatch(match); err != nil {
		s.logger.Error("Failed to save match", map[string]interface{}{
			"error": err,
		})
		return fmt.Errorf("failed to save match: %w", err)
	}

	return nil
}
