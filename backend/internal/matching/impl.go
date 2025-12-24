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
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)

	if name == "" {
		return ""
	}

	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "\u2013", " ")
	name = strings.ReplaceAll(name, "\u2014", " ")

	var normalized strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') ||
			(r >= '\u0430' && r <= '\u044f') || r == '\u0451' ||
			(r >= '0' && r <= '9') || r == ' ' || r == '/' {
			normalized.WriteRune(r)
		}
	}
	name = normalized.String()

	words := strings.Fields(name)
	filteredWords := make([]string, 0, len(words))
	stopWords := map[string]bool{
		"crni": true, "black": true, "white": true, "midnight": true,
		"gb": true, "mb": true, "tb": true,
		"pro": true, "max": true, "mini": true, "plus": true,
	}

	for _, word := range words {
		if word == "" {
			continue
		}
		if strings.Contains(word, "/") {
			parts := strings.Split(word, "/")
			if len(parts) > 0 {
				lastPart := stripMemorySuffix(parts[len(parts)-1])
				if lastPart != "" && !stopWords[lastPart] {
					filteredWords = append(filteredWords, lastPart)
				}
			}
			continue
		}

		word = stripMemorySuffix(word)
		if stopWords[word] {
			continue
		}
		if len(word) < 2 && !isNumericWord(word) {
			continue
		}
		filteredWords = append(filteredWords, word)
	}

	name = strings.Join(filteredWords, " ")
	return strings.TrimSpace(name)
}

func stripMemorySuffix(word string) string {
	for _, suffix := range []string{"gb", "mb", "tb"} {
		if strings.HasSuffix(word, suffix) {
			trimmed := strings.TrimSuffix(word, suffix)
			if trimmed != "" && isNumericWord(trimmed) {
				return trimmed
			}
			break
		}
	}
	return word
}

func isNumericWord(word string) bool {
	if word == "" {
		return false
	}
	for _, r := range word {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func diffWords(longer, shorter []string) []string {
	counts := make(map[string]int, len(shorter))
	for _, word := range shorter {
		counts[word]++
	}
	extra := make([]string, 0)
	for _, word := range longer {
		if counts[word] > 0 {
			counts[word]--
			continue
		}
		extra = append(extra, word)
	}
	return extra
}

func allNumericWords(words []string) bool {
	if len(words) == 0 {
		return false
	}
	for _, word := range words {
		if !isNumericWord(word) {
			return false
		}
	}
	return true
}


// normalizeBrand нормализует бренд
func (s *Service) normalizeBrand(brand string) string {
	if brand == "" {
		return ""
	}

	brand = strings.ToLower(strings.TrimSpace(brand))

	// Нормализуем тире и дефисы
	brand = strings.ReplaceAll(brand, "-", "")
	brand = strings.ReplaceAll(brand, "_", "")
	brand = strings.ReplaceAll(brand, " ", "")

	// Известные варианты написания брендов
	brandAliases := map[string]string{
		"samsung":  "samsung",
		"apple":    "apple",
		"xiaomi":   "xiaomi",
		"huawei":   "huawei",
		"motorola": "motorola",
		"lg":       "lg",
		"sony":     "sony",
		"nokia":    "nokia",
		"oneplus":  "oneplus",
		"oppo":     "oppo",
		"vivo":     "vivo",
		"realme":   "realme",
	}

	if normalized, ok := brandAliases[brand]; ok {
		return normalized
	}

	return brand
}

// calculateSimilarity рассчитывает схожесть между товарами
func (s *Service) calculateSimilarity(req *MatchRequest, product *Product) float64 {
	reqName := s.normalizeName(req.Name)
	prodName := s.normalizeName(product.Name)

	if reqName == prodName {
		if req.Brand != "" && product.Brand != "" {
			reqBrand := s.normalizeBrand(req.Brand)
			prodBrand := s.normalizeBrand(product.Brand)
			if reqBrand == prodBrand {
				return 1.0
			}
			return 0.95
		}
		return 0.95
	}

	reqWords := strings.Fields(reqName)
	prodWords := strings.Fields(prodName)

	similarity := 0.0
	skipBrandBonus := false

	if strings.Contains(reqName, prodName) || strings.Contains(prodName, reqName) {
		longer := reqWords
		shorter := prodWords
		if len(prodWords) > len(reqWords) {
			longer = prodWords
			shorter = reqWords
		}
		extraWords := diffWords(longer, shorter)
		if len(extraWords) > 0 && allNumericWords(extraWords) {
			similarity = 0.7
			skipBrandBonus = true
		} else {
			similarity = 0.65
		}
	} else {
		if len(reqWords) == 0 || len(prodWords) == 0 {
			return 0.0
		}

		commonWords := 0
		importantWords := 0

		for _, reqWord := range reqWords {
			if len(reqWord) <= 2 {
				continue
			}
			for _, prodWord := range prodWords {
				if reqWord == prodWord {
					commonWords++
					if commonWords <= 3 {
						importantWords++
					}
					break
				}
			}
		}

		if commonWords > 0 {
			totalWords := len(reqWords) + len(prodWords) - commonWords
			baseSimilarity := float64(commonWords) / float64(totalWords)

			importantBonus := float64(importantWords) * 0.15
			similarity = (baseSimilarity * 0.8) + importantBonus

			if commonWords >= 4 {
				similarity = 0.85
			}
		}
	}

	if !skipBrandBonus && req.Brand != "" && product.Brand != "" {
		reqBrand := s.normalizeBrand(req.Brand)
		prodBrand := s.normalizeBrand(product.Brand)
		if reqBrand == prodBrand {
			similarity += 0.2
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
