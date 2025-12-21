package autoconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/solomonczyk/izborator/internal/ai"
	"github.com/solomonczyk/izborator/internal/logger"
)

// Service сервис для автоматической генерации конфигов
type Service struct {
	storage Storage
	ai      *ai.Client
	log     *logger.Logger
}

// NewService создает новый сервис AutoConfig
func NewService(storage Storage, ai *ai.Client, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		ai:      ai,
		log:     log,
	}
}

// ProcessNextCandidate берет одного кандидата и пытается создать конфиг
func (s *Service) ProcessNextCandidate(ctx context.Context) error {
	if s.ai == nil {
		return fmt.Errorf("AI client is not available")
	}

	candidates, err := s.storage.GetClassifiedCandidates(1)
	if err != nil {
		return fmt.Errorf("failed to get candidates: %w", err)
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no candidates available")
	}
	candidate := candidates[0]

	s.log.Info("🤖 Auto-configuring shop", map[string]interface{}{
		"domain": candidate.Domain,
		"id":     candidate.ID,
	})

	// 1. Scout: Ищем страницу товара
	productURL, err := s.findProductPage(candidate.Domain)
	if err != nil {
		s.log.Error("Scout failed", map[string]interface{}{
			"domain": candidate.Domain,
			"error":  err.Error(),
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "scout_failed: "+err.Error())
		return fmt.Errorf("scout failed: %w", err)
	}
	s.log.Info("Found product page", map[string]interface{}{
		"url": productURL,
	})

	// 2. Fetch & Clean: Скачиваем HTML
	html, err := s.fetchHTML(productURL)
	if err != nil {
		s.log.Error("Failed to fetch HTML", map[string]interface{}{
			"url":   productURL,
			"error": err.Error(),
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "fetch_failed: "+err.Error())
		return fmt.Errorf("fetch failed: %w", err)
	}

	cleanHTML, err := CleanHTML(html)
	if err != nil {
		s.log.Warn("HTML cleaning failed, using raw HTML", map[string]interface{}{
			"error": err.Error(),
		})
		cleanHTML = html // Используем сырой HTML, если очистка не удалась
	}

	// 3. Brain: Спрашиваем AI
	s.log.Info("Asking AI for selectors", map[string]interface{}{
		"html_length": len(cleanHTML),
	})
	selectorsJSON, err := s.ai.GenerateSelectors(ctx, cleanHTML)
	if err != nil {
		s.log.Error("AI generation failed", map[string]interface{}{
			"error": err.Error(),
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "ai_failed: "+err.Error())
		return fmt.Errorf("AI generation failed: %w", err)
	}

	// 4. Parse JSON
	var selectors map[string]string
	if err := json.Unmarshal([]byte(selectorsJSON), &selectors); err != nil {
		s.log.Error("Invalid JSON from AI", map[string]interface{}{
			"json":  selectorsJSON,
			"error": err.Error(),
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "invalid_json: "+err.Error())
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Проверяем, что есть хотя бы name и price
	if selectors["name"] == "" || selectors["price"] == "" {
		s.log.Warn("Missing required selectors", map[string]interface{}{
			"selectors": selectors,
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "missing_required_selectors")
		return fmt.Errorf("missing required selectors (name or price)")
	}

	// 5. Validate: Проверяем, работают ли селекторы
	if err := s.validateSelectors(productURL, selectors); err != nil {
		s.log.Warn("Validation failed", map[string]interface{}{
			"error":    err.Error(),
			"selectors": selectors,
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "validation_failed: "+err.Error())
		return fmt.Errorf("validation failed: %w", err)
	}

	// 6. Success: Сохраняем!
	s.log.Info("✨ SUCCESS! Config generated", map[string]interface{}{
		"selectors": selectors,
		"domain":    candidate.Domain,
	})
	return s.storage.MarkAsConfigured(candidate.ID, ShopConfig{Selectors: selectors})
}

// --- Helpers ---

// findProductPage ищет ссылку на товар с главной страницы
func (s *Service) findProductPage(domain string) (string, error) {
	baseURL := domain
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	var bestLink string
	maxScore := 0

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
		colly.MaxDepth(1),
	)
	c.SetRequestTimeout(30 * time.Second)
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		link = e.Request.AbsoluteURL(link)

		// Игнорируем внешние ссылки
		if !strings.Contains(link, domain) {
			return
		}

		// Эвристика: ссылка на товар обычно длинная и содержит ключевые слова
		score := 0
		linkLower := strings.ToLower(link)

		// Ключевые слова для страниц товаров
		if strings.Contains(linkLower, "/proizvod/") || strings.Contains(linkLower, "/p/") ||
			strings.Contains(linkLower, "/product/") || strings.Contains(linkLower, "/artikal/") {
			score += 50
		}

		// Длинная ссылка (обычно товары имеют длинные URL)
		if len(link) > len(baseURL)+20 {
			score += 10
		}

		// Игнорируем мусор
		if strings.Contains(linkLower, "login") || strings.Contains(linkLower, "cart") ||
			strings.Contains(linkLower, "facebook") || strings.Contains(linkLower, "twitter") ||
			strings.Contains(linkLower, "instagram") || strings.Contains(linkLower, "contact") ||
			strings.Contains(linkLower, "about") || strings.Contains(linkLower, "privacy") {
			score = -100
		}

		// Игнорируем якоря и пустые ссылки
		if strings.HasPrefix(link, "#") || link == "" || link == baseURL || link == baseURL+"/" {
			return
		}

		if score > maxScore {
			maxScore = score
			bestLink = link
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		s.log.Warn("Error during scout", map[string]interface{}{
			"url":   r.Request.URL.String(),
			"error": err.Error(),
		})
	})

	err := c.Visit(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to visit domain: %w", err)
	}

	if bestLink == "" {
		return "", fmt.Errorf("no product link found on %s", baseURL)
	}

	return bestLink, nil
}

// fetchHTML скачивает HTML страницы
func (s *Service) fetchHTML(url string) (string, error) {
	var html string
	var fetchErr error

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)
	c.SetRequestTimeout(60 * time.Second)
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.OnResponse(func(r *colly.Response) {
		html = string(r.Body)
	})

	c.OnError(func(r *colly.Response, err error) {
		fetchErr = err
	})

	err := c.Visit(url)
	if err != nil {
		return "", fmt.Errorf("failed to visit URL: %w", err)
	}
	if fetchErr != nil {
		return "", fmt.Errorf("error during fetch: %w", fetchErr)
	}
	if html == "" {
		return "", fmt.Errorf("empty HTML response")
	}

	return html, nil
}

// validateSelectors проверяет, что селекторы работают и извлекают данные
func (s *Service) validateSelectors(url string, selectors map[string]string) error {
	var name, price string
	var validationErr error

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)
	c.SetRequestTimeout(60 * time.Second)
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	nameSel := selectors["name"]
	priceSel := selectors["price"]

	if nameSel == "" || priceSel == "" {
		return fmt.Errorf("missing required selectors: name=%s, price=%s", nameSel, priceSel)
	}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		name = strings.TrimSpace(e.ChildText(nameSel))
		price = strings.TrimSpace(e.ChildText(priceSel))
	})

	c.OnError(func(r *colly.Response, err error) {
		validationErr = err
	})

	if err := c.Visit(url); err != nil {
		return fmt.Errorf("failed to visit URL for validation: %w", err)
	}
	if validationErr != nil {
		return fmt.Errorf("error during validation: %w", validationErr)
	}

	// Проверяем, что данные извлечены
	if name == "" {
		return fmt.Errorf("name selector '%s' did not extract data", nameSel)
	}
	if price == "" {
		return fmt.Errorf("price selector '%s' did not extract data", priceSel)
	}

	// Проверяем, что цена содержит числа
	if !strings.ContainsAny(price, "0123456789") {
		return fmt.Errorf("price selector extracted non-numeric value: '%s'", price)
	}

	s.log.Info("Validation successful", map[string]interface{}{
		"name":  name,
		"price": price,
	})

	return nil
}

