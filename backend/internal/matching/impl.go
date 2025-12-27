package matching

import (
	"fmt"
	"regexp"
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

	// Определяем тип продукта
	productType := req.Type
	if productType == "" {
		productType = "good" // По умолчанию товар
	}

	// Нормализуем данные для поиска
	normalizedName := s.normalizeName(req.Name, productType)
	normalizedBrand := s.normalizeBrand(req.Brand)

	// Ищем похожие товары или услуги
	similar, err := s.storage.FindSimilarProducts(normalizedName, normalizedBrand, productType, 10)
	if err != nil {
		s.logger.Error("Failed to find similar products", map[string]interface{}{
			"error": err,
		})
		return nil, fmt.Errorf("failed to find similar: %w", err)
	}

	// Рассчитываем схожесть для каждого найденного товара или услуги
	matches := make([]*ProductMatch, 0, len(similar))
	for _, product := range similar {
		// Для услуг используем более мягкий порог
		threshold := 0.5
		if productType == "service" {
			threshold = 0.4 // Более мягкий порог для услуг
		}

		similarity := s.calculateSimilarity(req, product, productType)

		if similarity > threshold {
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

// normalizeUnits нормализует единицы измерения
func (s *Service) normalizeUnits(text string) string {
	// Нормализация единиц веса
	text = regexp.MustCompile(`(?i)\b(\d+)\s*(kg|кило|килограм|килограма|килограма)\b`).ReplaceAllString(text, "${1}kg")
	text = regexp.MustCompile(`(?i)\b(\d+)\s*(g|gr|грам|грама|грама)\b`).ReplaceAllString(text, "${1}g")
	text = regexp.MustCompile(`(?i)\b(\d+)\s*(mg|милиграм|милиграма)\b`).ReplaceAllString(text, "${1}mg")

	// Нормализация единиц объема
	text = regexp.MustCompile(`(?i)\b(\d+)\s*(l|литр|литра|литре)\b`).ReplaceAllString(text, "${1}l")
	text = regexp.MustCompile(`(?i)\b(\d+)\s*(ml|мл|милилитр|милилитра)\b`).ReplaceAllString(text, "${1}ml")

	// Нормализация единиц времени (для услуг)
	text = regexp.MustCompile(`(?i)\b(\d+)\s*(час|часа|часа|h|hr|hours?)\b`).ReplaceAllString(text, "${1}h")
	text = regexp.MustCompile(`(?i)\b(\d+)\s*(мин|минут|минута|min|mins?)\b`).ReplaceAllString(text, "${1}min")
	text = regexp.MustCompile(`(?i)\b(\d+)\s*(сек|секунд|секунда|sec|secs?)\b`).ReplaceAllString(text, "${1}sec")

	return text
}

// normalizeName нормализует название товара или услуги для сравнения
func (s *Service) normalizeName(name string, productType string) string {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)

	if name == "" {
		return ""
	}

	// Нормализация единиц измерения
	name = s.normalizeUnits(name)

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
	
	// Стоп-слова для товаров
	stopWordsGoods := map[string]bool{
		"crni": true, "black": true, "white": true, "midnight": true,
		"gb": true, "mb": true, "tb": true,
		"pro": true, "max": true, "mini": true, "plus": true,
	}
	
	// Стоп-слова для услуг (менее агрессивные)
	stopWordsServices := map[string]bool{
		"usluga": true, "usluge": true, "service": true, "services": true,
		"cena": true, "cene": true, "price": true, "prices": true,
	}
	
	stopWords := stopWordsGoods
	if productType == "service" {
		stopWords = stopWordsServices
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

// calculateSimilarity рассчитывает схожесть между товарами или услугами
func (s *Service) calculateSimilarity(req *MatchRequest, product *Product, productType string) float64 {
	reqName := s.normalizeName(req.Name, productType)
	prodName := s.normalizeName(product.Name, productType)

	if reqName == prodName {
		// Для услуг бренд менее важен
		if productType == "service" {
			return 0.95 // Высокая схожесть даже без бренда
		}
		
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

		// Для услуг используем fuzzy matching (частичное совпадение слов)
		useFuzzy := productType == "service"

		for _, reqWord := range reqWords {
			if len(reqWord) <= 2 {
				continue
			}
			for _, prodWord := range prodWords {
				matched := false
				if reqWord == prodWord {
					matched = true
				} else if useFuzzy {
					// Fuzzy matching для услуг: проверяем, содержит ли одно слово другое
					if len(reqWord) >= 4 && len(prodWord) >= 4 {
						if strings.Contains(reqWord, prodWord) || strings.Contains(prodWord, reqWord) {
							matched = true
						}
					}
				}
				
				if matched {
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

	// Для услуг бренд менее важен (или вообще не важен)
	if !skipBrandBonus && productType != "service" && req.Brand != "" && product.Brand != "" {
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
