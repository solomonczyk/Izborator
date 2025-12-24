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

	if name == "" {
		return ""
	}

	// Нормализуем тире и дефисы (заменяем на пробелы)
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "–", " ") // en-dash
	name = strings.ReplaceAll(name, "—", " ") // em-dash

	// Удаляем спецсимволы, оставляем только буквы (включая кириллицу), цифры, пробелы и слэш
	var normalized strings.Builder
	for _, r := range name {
		// Буквы (латиница, кириллица), цифры, пробелы, слэш
		if (r >= 'a' && r <= 'z') ||
			(r >= 'а' && r <= 'я') ||
			(r >= '0' && r <= '9') ||
			r == ' ' ||
			r == '/' ||
			r == 'ё' || r == 'Ё' {
			normalized.WriteRune(r)
		}
	}
	name = normalized.String()

	// Нормализуем пробелы (множественные пробелы -> один)
	words := strings.Fields(name)

	// Удаляем несущественные слова (цвета, описания) и нормализуем память
	filteredWords := make([]string, 0, len(words))
	stopWords := map[string]bool{
		"crni": true, "black": true, "white": true, "midnight": true,
		"gb": true, "mb": true, "tb": true,
		"pro": true, "max": true, "mini": true, "plus": true,
	}

	for _, word := range words {
		// Пропускаем стоп-слова
		if stopWords[word] {
			continue
		}
		// Пропускаем очень короткие слова (меньше 2 символов), кроме цифр
		if len(word) < 2 && !(word >= "0" && word <= "9") {
			continue
		}
		// Нормализуем память: "12/512gb" -> "512", "512" -> "512"
		if strings.Contains(word, "/") {
			parts := strings.Split(word, "/")
			if len(parts) > 0 {
				// Берем последнюю часть (обычно это память)
				lastPart := parts[len(parts)-1]
				// Удаляем единицы измерения если есть
				lastPart = strings.TrimSuffix(lastPart, "gb")
				lastPart = strings.TrimSuffix(lastPart, "mb")
				lastPart = strings.TrimSuffix(lastPart, "tb")
				if lastPart != "" {
					filteredWords = append(filteredWords, lastPart)
				}
			}
		} else {
			filteredWords = append(filteredWords, word)
		}
	}

	name = strings.Join(filteredWords, " ")
	return strings.TrimSpace(name)
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
	// Нормализуем названия и бренды для сравнения
	reqName := s.normalizeName(req.Name)
	prodName := s.normalizeName(product.Name)

	// Точное совпадение названий = 100% similarity
	if reqName == prodName {
		// Проверяем бренды для дополнительной уверенности
		if req.Brand != "" && product.Brand != "" {
			reqBrand := s.normalizeBrand(req.Brand)
			prodBrand := s.normalizeBrand(product.Brand)
			if reqBrand == prodBrand {
				return 1.0 // Полное совпадение
			}
			// Названия совпадают, но бренды разные - всё равно высокая схожесть
			return 0.95
		}
		// Названия совпадают, бренды не указаны или один пустой
		return 0.95
	}

	// Частичное совпадение названий
	similarity := 0.0
	if strings.Contains(reqName, prodName) || strings.Contains(prodName, reqName) {
		similarity = 0.7 // Увеличиваем для частичного совпадения
	} else {
		// Проверка на общие слова с улучшенным алгоритмом
		reqWords := strings.Fields(reqName)
		prodWords := strings.Fields(prodName)

		if len(reqWords) == 0 || len(prodWords) == 0 {
			return 0.0
		}

		commonWords := 0
		importantWords := 0 // Ключевые слова (бренд, модель)

		for _, reqWord := range reqWords {
			if len(reqWord) <= 2 {
				continue // Игнорируем короткие слова
			}
			for _, prodWord := range prodWords {
				if reqWord == prodWord {
					commonWords++
					// Ключевые слова (первые слова обычно важнее)
					if commonWords <= 3 {
						importantWords++
					}
					break
				}
			}
		}

		if commonWords > 0 {
			// Jaccard similarity с бонусом за ключевые слова
			totalWords := len(reqWords) + len(prodWords) - commonWords
			baseSimilarity := float64(commonWords) / float64(totalWords)

			// Бонус за совпадение ключевых слов (бренд, модель)
			importantBonus := float64(importantWords) * 0.15

			similarity = (baseSimilarity * 0.8) + importantBonus

			// Если совпало много слов - это точно один товар
			if commonWords >= 4 {
				similarity = 0.85
			}
		}
	}

	// Бонус за совпадение брендов
	if req.Brand != "" && product.Brand != "" {
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
