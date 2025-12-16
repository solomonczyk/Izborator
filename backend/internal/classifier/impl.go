package classifier

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	// Пороги для классификации
	thresholdShop      = 0.8  // Если TotalScore > 0.8, это магазин
	thresholdReview   = 0.5  // Если TotalScore > 0.5, требуется ручная проверка

	// Веса для подсчета TotalScore
	weightKeywords  = 0.35
	weightPlatform  = 0.35  // Увеличиваем вес платформы
	weightStructure = 0.30
)

// Ключевые слова для определения магазинов (сербский язык)
var shopKeywords = []string{
	"korpa", "korpa za kupovinu", "dodaj u korpu", "dodaj u korpu za kupovinu",
	"cena", "cijena", "rsd", "din", "dinara",
	"kupi", "kupi odmah", "naruci", "naruci odmah",
	"proizvod", "proizvodi", "katalog", "katalog proizvoda",
	"akcija", "akcije", "popust", "snizenje",
	"dostava", "isporuka", "placanje", "na rate",
	"shop", "store", "prodavnica", "online shop",
	"checkout", "kosarica", "narudzba", "porudzbina",
	"cena sa pdv", "cena bez pdv", "ukupno", "total",
}

// Платформы E-commerce
var ecommercePlatforms = map[string][]string{
	"shopify":     {"shopify", "cdn.shopify.com", "myshopify.com"},
	"woocommerce": {"woocommerce", "wp-content/plugins/woocommerce", "wc-"},
	"magento":    {"magento", "mage/", "static/version", "magento/"},
	"opencart":   {"opencart", "index.php?route=", "opencart"},
	"prestashop": {"prestashop", "prestashop.com"},
	"next.js":    {"__next", "next.js", "_next/static", "next/"},
	"react":      {"react", "react-dom", "__react"},
}

// Classify анализирует домен и определяет, является ли он магазином
func (s *Service) Classify(ctx context.Context, domain string) (*ClassificationResult, error) {
	s.logger.Info("Starting classification", map[string]interface{}{
		"domain": domain,
	})

	// Добавляем протокол, если его нет
	url := domain
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Скачиваем главную страницу
	html, err := s.fetchPage(ctx, url)
	if err != nil {
		s.logger.Error("Failed to fetch page", map[string]interface{}{
			"domain": domain,
			"error":  err.Error(),
		})
		return &ClassificationResult{
			IsShop: false,
			Reasons: []string{fmt.Sprintf("Failed to fetch page: %v", err)},
		}, nil
	}

	// Нормализуем HTML (нижний регистр для поиска)
	htmlLower := strings.ToLower(html)

	// 1. Анализ ключевых слов
	keywordsScore := s.analyzeKeywords(htmlLower)

	// 2. Анализ платформы
	platformScore, detectedPlatform := s.analyzePlatform(htmlLower)

	// 3. Анализ структуры
	structureScore := s.analyzeStructure(htmlLower, html)

	// 4. Подсчет общего скора
	totalScore := (keywordsScore * weightKeywords) +
		(platformScore * weightPlatform) +
		(structureScore * weightStructure)

	score := ClassificationScore{
		KeywordsScore:  keywordsScore,
		PlatformScore:  platformScore,
		StructureScore: structureScore,
		TotalScore:     totalScore,
	}

	// 5. Принятие решения
	isShop := totalScore >= thresholdShop
	reasons := s.generateReasons(score, detectedPlatform, isShop)

	result := &ClassificationResult{
		IsShop:          isShop,
		Score:           score,
		DetectedPlatform: detectedPlatform,
		Reasons:         reasons,
	}

	s.logger.Info("Classification completed", map[string]interface{}{
		"domain":           domain,
		"is_shop":          isShop,
		"total_score":       totalScore,
		"detected_platform": detectedPlatform,
	})

	return result, nil
}

// fetchPage скачивает HTML страницы
func (s *Service) fetchPage(ctx context.Context, url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	// Устанавливаем User-Agent, чтобы не выглядеть как бот
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "sr-RS,sr;q=0.9,en;q=0.8")
	// Не указываем Accept-Encoding, чтобы Go автоматически распаковал ответ
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Обрабатываем gzip, если есть
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// analyzeKeywords анализирует наличие ключевых слов
func (s *Service) analyzeKeywords(htmlLower string) float64 {
	foundCount := 0
	for _, keyword := range shopKeywords {
		if strings.Contains(htmlLower, keyword) {
			foundCount++
		}
	}

	// Нормализуем: если найдено больше половины ключевых слов, это 1.0
	// Иначе пропорционально
	maxScore := float64(len(shopKeywords)) * 0.5
	if maxScore == 0 {
		return 0
	}

	score := float64(foundCount) / maxScore
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// analyzePlatform анализирует наличие признаков E-commerce платформ
func (s *Service) analyzePlatform(htmlLower string) (float64, string) {
	maxScore := 0.0
	detectedPlatform := ""

	for platform, patterns := range ecommercePlatforms {
		score := 0.0
		for _, pattern := range patterns {
			if strings.Contains(htmlLower, pattern) {
				score += 0.5
			}
		}
		if score > maxScore {
			maxScore = score
			detectedPlatform = platform
		}
	}

	// Нормализуем до 1.0
	if maxScore > 1.0 {
		maxScore = 1.0
	}

	return maxScore, detectedPlatform
}

// analyzeStructure анализирует структуру страницы
func (s *Service) analyzeStructure(htmlLower, html string) float64 {
	score := 0.0

	// Проверка наличия иконки корзины (расширенный поиск)
	cartPatterns := []string{
		`class="[^"]*cart[^"]*"`,
		`class="[^"]*korpa[^"]*"`,
		`id="[^"]*cart[^"]*"`,
		`id="[^"]*korpa[^"]*"`,
		`data-[^=]*cart`,
		`aria-label="[^"]*korpa`,
		`aria-label="[^"]*cart`,
		`title="[^"]*korpa`,
		`title="[^"]*cart`,
		`shopping[_-]?cart`,
		`basket`,
	}
	for _, pattern := range cartPatterns {
		matched, _ := regexp.MatchString(pattern, htmlLower)
		if matched {
			score += 0.2
			break
		}
	}

	// Проверка наличия цен в формате RSD (расширенный поиск)
	pricePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\d+[\s,.]?\d*\s*(?:rsd|din|dinara)`),
		regexp.MustCompile(`\d+[\s,.]?\d*\s*д[иі]н`), // Кириллица
		regexp.MustCompile(`cena[:\s]+\d+`),
		regexp.MustCompile(`price[:\s]+\d+`),
		regexp.MustCompile(`\d+\s*€`), // Евро тоже может быть
	}
	for _, pattern := range pricePatterns {
		if pattern.MatchString(htmlLower) {
			score += 0.2
			break
		}
	}

	// Проверка наличия кнопок "Купить" (расширенный поиск)
	buyPatterns := []string{
		`kupi`,
		`dodaj`,
		`naruci`,
		`buy now`,
		`add to cart`,
		`dodaj u korpu`,
		`dodaj u korpu za kupovinu`,
		`kupi odmah`,
		`naruci odmah`,
		`button[^>]*kupi`,
		`button[^>]*dodaj`,
	}
	for _, pattern := range buyPatterns {
		if strings.Contains(htmlLower, pattern) {
			score += 0.2
			break
		}
	}

	// Проверка наличия структурированных данных (schema.org)
	if strings.Contains(htmlLower, "schema.org/product") || 
		strings.Contains(htmlLower, "itemtype=\"http://schema.org/product\"") ||
		strings.Contains(htmlLower, "itemtype=\"https://schema.org/product\"") {
		score += 0.2
	}

	// Проверка наличия каталога/товаров в URL или структуре
	catalogPatterns := []string{
		`/proizvod/`,
		`/product/`,
		`/katalog/`,
		`/catalog/`,
		`product-`,
		`proizvod-`,
		`class="[^"]*product[^"]*"`,
		`class="[^"]*proizvod[^"]*"`,
	}
	for _, pattern := range catalogPatterns {
		if strings.Contains(htmlLower, pattern) {
			score += 0.2
			break
		}
	}

	// Ограничиваем максимальный score до 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// generateReasons генерирует список причин решения
func (s *Service) generateReasons(score ClassificationScore, platform string, isShop bool) []string {
	reasons := []string{}

	if score.KeywordsScore > 0.5 {
		reasons = append(reasons, fmt.Sprintf("Найдены ключевые слова магазина (score: %.2f)", score.KeywordsScore))
	}

	if score.PlatformScore > 0.5 && platform != "" {
		reasons = append(reasons, fmt.Sprintf("Обнаружена платформа: %s (score: %.2f)", platform, score.PlatformScore))
	}

	if score.StructureScore > 0.5 {
		reasons = append(reasons, fmt.Sprintf("Структура страницы похожа на магазин (score: %.2f)", score.StructureScore))
	}

	if isShop {
		reasons = append(reasons, fmt.Sprintf("Общий score: %.2f (>= %.2f) → Это магазин", score.TotalScore, thresholdShop))
	} else if score.TotalScore >= thresholdReview {
		reasons = append(reasons, fmt.Sprintf("Общий score: %.2f (требуется ручная проверка)", score.TotalScore))
	} else {
		reasons = append(reasons, fmt.Sprintf("Общий score: %.2f (< %.2f) → Не магазин", score.TotalScore, thresholdReview))
	}

	return reasons
}

