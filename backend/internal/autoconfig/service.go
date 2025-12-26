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

// Service Ã‘ÂÃÂµÃ‘â‚¬ÃÂ²ÃÂ¸Ã‘Â ÃÂ´ÃÂ»Ã‘Â ÃÂ°ÃÂ²Ã‘â€šÃÂ¾ÃÂ¼ÃÂ°Ã‘â€šÃÂ¸Ã‘â€¡ÃÂµÃ‘ÂÃÂºÃÂ¾ÃÂ¹ ÃÂ³ÃÂµÃÂ½ÃÂµÃ‘â‚¬ÃÂ°Ã‘â€ ÃÂ¸ÃÂ¸ ÃÂºÃÂ¾ÃÂ½Ã‘â€žÃÂ¸ÃÂ³ÃÂ¾ÃÂ²
type Service struct {
	storage Storage
	ai      *ai.Client
	log     *logger.Logger
}

// NewService Ã‘ÂÃÂ¾ÃÂ·ÃÂ´ÃÂ°ÃÂµÃ‘â€š ÃÂ½ÃÂ¾ÃÂ²Ã‘â€¹ÃÂ¹ Ã‘ÂÃÂµÃ‘â‚¬ÃÂ²ÃÂ¸Ã‘Â AutoConfig
func NewService(storage Storage, ai *ai.Client, log *logger.Logger) *Service {
	return &Service{
		storage: storage,
		ai:      ai,
		log:     log,
	}
}

// ProcessNextCandidate ÃÂ±ÃÂµÃ‘â‚¬ÃÂµÃ‘â€š ÃÂ¾ÃÂ´ÃÂ½ÃÂ¾ÃÂ³ÃÂ¾ ÃÂºÃÂ°ÃÂ½ÃÂ´ÃÂ¸ÃÂ´ÃÂ°Ã‘â€šÃÂ° ÃÂ¸ ÃÂ¿Ã‘â€¹Ã‘â€šÃÂ°ÃÂµÃ‘â€šÃ‘ÂÃ‘Â Ã‘ÂÃÂ¾ÃÂ·ÃÂ´ÃÂ°Ã‘â€šÃ‘Å’ ÃÂºÃÂ¾ÃÂ½Ã‘â€žÃÂ¸ÃÂ³
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

	s.log.Info("Ã°Å¸Â¤â€“ Auto-configuring shop", map[string]interface{}{
		"domain": candidate.Domain,
		"id":     candidate.ID,
	})

	// 1. Scout: ÃËœÃ‘â€°ÃÂµÃÂ¼ Ã‘ÂÃ‘â€šÃ‘â‚¬ÃÂ°ÃÂ½ÃÂ¸Ã‘â€ Ã‘Æ’ Ã‘â€šÃÂ¾ÃÂ²ÃÂ°Ã‘â‚¬ÃÂ°
	productURL, err := s.findProductPage(candidate.Domain, siteType)
	if err != nil {
		s.log.Error("Scout failed", map[string]interface{}{
			"domain": candidate.Domain,
			"error":  err.Error(),
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "scout_failed: "+err.Error())
		return fmt.Errorf("scout failed: %w", err)
	}
	s.log.Info("Found page", map[string]interface{}{
		"url":       productURL,
		"site_type": siteType,
	})

	// 2. Fetch & Clean: ÃÂ¡ÃÂºÃÂ°Ã‘â€¡ÃÂ¸ÃÂ²ÃÂ°ÃÂµÃÂ¼ HTML
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
		cleanHTML = html // ÃËœÃ‘ÂÃÂ¿ÃÂ¾ÃÂ»Ã‘Å’ÃÂ·Ã‘Æ’ÃÂµÃÂ¼ Ã‘ÂÃ‘â€¹Ã‘â‚¬ÃÂ¾ÃÂ¹ HTML, ÃÂµÃ‘ÂÃÂ»ÃÂ¸ ÃÂ¾Ã‘â€¡ÃÂ¸Ã‘ÂÃ‘â€šÃÂºÃÂ° ÃÂ½ÃÂµ Ã‘Æ’ÃÂ´ÃÂ°ÃÂ»ÃÂ°Ã‘ÂÃ‘Å’
	}

	// 3. Brain: ÃÂ¡ÃÂ¿Ã‘â‚¬ÃÂ°Ã‘Ë†ÃÂ¸ÃÂ²ÃÂ°ÃÂµÃÂ¼ AI
	s.log.Info("Asking AI for selectors", map[string]interface{}{
		"html_length": len(cleanHTML),
		"site_type":   siteType,
	})
	selectorsJSON, err := s.ai.GenerateSelectors(ctx, cleanHTML, siteType)
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

	// ÃÅ¸Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃÂ¼, Ã‘â€¡Ã‘â€šÃÂ¾ ÃÂµÃ‘ÂÃ‘â€šÃ‘Å’ Ã‘â€¦ÃÂ¾Ã‘â€šÃ‘Â ÃÂ±Ã‘â€¹ name ÃÂ¸ price
	if selectors["name"] == "" || selectors["price"] == "" {
		s.log.Warn("Missing required selectors", map[string]interface{}{
			"selectors": selectors,
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "missing_required_selectors")
		return fmt.Errorf("missing required selectors (name or price)")
	}

	// 5. Validate: ÃÅ¸Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃÂ¼, Ã‘â‚¬ÃÂ°ÃÂ±ÃÂ¾Ã‘â€šÃÂ°Ã‘Å½Ã‘â€š ÃÂ»ÃÂ¸ Ã‘ÂÃÂµÃÂ»ÃÂµÃÂºÃ‘â€šÃÂ¾Ã‘â‚¬Ã‘â€¹
	if err := s.validateSelectors(productURL, selectors); err != nil {
		s.log.Warn("Validation failed", map[string]interface{}{
			"error":     err.Error(),
			"selectors": selectors,
		})
		_ = s.storage.MarkAsFailed(candidate.ID, "validation_failed: "+err.Error())
		return fmt.Errorf("validation failed: %w", err)
	}

	// 6. Success: ÃÂ¡ÃÂ¾Ã‘â€¦Ã‘â‚¬ÃÂ°ÃÂ½Ã‘ÂÃÂµÃÂ¼!
	s.log.Info("Ã¢Å“Â¨ SUCCESS! Config generated", map[string]interface{}{
		"selectors": selectors,
		"domain":    candidate.Domain,
	})
	return s.storage.MarkAsConfigured(candidate.ID, ShopConfig{Selectors: selectors})
}

// --- Helpers ---

// findProductPage ÃÂ¸Ã‘â€°ÃÂµÃ‘â€š Ã‘ÂÃ‘ÂÃ‘â€¹ÃÂ»ÃÂºÃ‘Æ’ ÃÂ½ÃÂ° Ã‘â€šÃÂ¾ÃÂ²ÃÂ°Ã‘â‚¬ Ã‘Â ÃÂ³ÃÂ»ÃÂ°ÃÂ²ÃÂ½ÃÂ¾ÃÂ¹ Ã‘ÂÃ‘â€šÃ‘â‚¬ÃÂ°ÃÂ½ÃÂ¸Ã‘â€ Ã‘â€¹
func (s *Service) findProductPage(domain string, siteType string) (string, error) {
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

		// ÃËœÃÂ³ÃÂ½ÃÂ¾Ã‘â‚¬ÃÂ¸Ã‘â‚¬Ã‘Æ’ÃÂµÃÂ¼ ÃÂ²ÃÂ½ÃÂµÃ‘Ë†ÃÂ½ÃÂ¸ÃÂµ Ã‘ÂÃ‘ÂÃ‘â€¹ÃÂ»ÃÂºÃÂ¸
		if !strings.Contains(link, domain) {
			return
		}

		// ÃÂ­ÃÂ²Ã‘â‚¬ÃÂ¸Ã‘ÂÃ‘â€šÃÂ¸ÃÂºÃÂ°: Ã‘ÂÃ‘ÂÃ‘â€¹ÃÂ»ÃÂºÃÂ° ÃÂ½ÃÂ° Ã‘â€šÃÂ¾ÃÂ²ÃÂ°Ã‘â‚¬ ÃÂ¾ÃÂ±Ã‘â€¹Ã‘â€¡ÃÂ½ÃÂ¾ ÃÂ´ÃÂ»ÃÂ¸ÃÂ½ÃÂ½ÃÂ°Ã‘Â ÃÂ¸ Ã‘ÂÃÂ¾ÃÂ´ÃÂµÃ‘â‚¬ÃÂ¶ÃÂ¸Ã‘â€š ÃÂºÃÂ»Ã‘Å½Ã‘â€¡ÃÂµÃÂ²Ã‘â€¹ÃÂµ Ã‘ÂÃÂ»ÃÂ¾ÃÂ²ÃÂ°
		score := 0
		linkLower := strings.ToLower(link)

		// ÃËœÃÂ³ÃÂ½ÃÂ¾Ã‘â‚¬ÃÂ¸Ã‘â‚¬Ã‘Æ’ÃÂµÃÂ¼ Ã‘ÂÃ‘â€šÃ‘â‚¬ÃÂ°ÃÂ½ÃÂ¸Ã‘â€ Ã‘â€¹ ÃÂºÃÂ¾ÃÂ»ÃÂ»ÃÂµÃÂºÃ‘â€ ÃÂ¸ÃÂ¹/ÃÂºÃÂ°Ã‘â€šÃÂµÃÂ³ÃÂ¾Ã‘â‚¬ÃÂ¸ÃÂ¹ (ÃÂ½ÃÂµ Ã‘â€šÃÂ¾ÃÂ²ÃÂ°Ã‘â‚¬Ã‘â€¹)
		if strings.Contains(linkLower, "/collections/") || strings.Contains(linkLower, "/collection/") ||
			strings.Contains(linkLower, "/category/") || strings.Contains(linkLower, "/kategorija/") ||
			strings.Contains(linkLower, "/kategorije/") || strings.Contains(linkLower, "/categories/") {
			return
		}

		// ÃÅ¡ÃÂ»Ã‘Å½Ã‘â€¡ÃÂµÃÂ²Ã‘â€¹ÃÂµ Ã‘ÂÃÂ»ÃÂ¾ÃÂ²ÃÂ° ÃÂ´ÃÂ»Ã‘Â Ã‘ÂÃ‘â€šÃ‘â‚¬ÃÂ°ÃÂ½ÃÂ¸Ã‘â€  Ã‘â€šÃÂ¾ÃÂ²ÃÂ°Ã‘â‚¬ÃÂ¾ÃÂ²
		if siteType == "ecommerce" {
			if strings.Contains(linkLower, "/proizvod/") || strings.Contains(linkLower, "/p/") ||
				strings.Contains(linkLower, "/product/") || strings.Contains(linkLower, "/artikal/") ||
				strings.Contains(linkLower, "/products/") || strings.Contains(linkLower, "/proizvodi/") {
				score += 50
			}
		}

		if siteType == "service_provider" {
			if strings.Contains(linkLower, "cenovnik") || strings.Contains(linkLower, "cene") ||
				strings.Contains(linkLower, "usluge") || strings.Contains(linkLower, "price") ||
				strings.Contains(linkLower, "pricelist") || strings.Contains(linkLower, "tabela") {
				score += 50
			}
		}

		// Ãâ€ÃÂ»ÃÂ¸ÃÂ½ÃÂ½ÃÂ°Ã‘Â Ã‘ÂÃ‘ÂÃ‘â€¹ÃÂ»ÃÂºÃÂ° (ÃÂ¾ÃÂ±Ã‘â€¹Ã‘â€¡ÃÂ½ÃÂ¾ Ã‘â€šÃÂ¾ÃÂ²ÃÂ°Ã‘â‚¬Ã‘â€¹ ÃÂ¸ÃÂ¼ÃÂµÃ‘Å½Ã‘â€š ÃÂ´ÃÂ»ÃÂ¸ÃÂ½ÃÂ½Ã‘â€¹ÃÂµ URL)
		if len(link) > len(baseURL)+20 {
			score += 10
		}

		// ÃËœÃÂ³ÃÂ½ÃÂ¾Ã‘â‚¬ÃÂ¸Ã‘â‚¬Ã‘Æ’ÃÂµÃÂ¼ ÃÂ¼Ã‘Æ’Ã‘ÂÃÂ¾Ã‘â‚¬
		if strings.Contains(linkLower, "login") || strings.Contains(linkLower, "cart") ||
			strings.Contains(linkLower, "facebook") || strings.Contains(linkLower, "twitter") ||
			strings.Contains(linkLower, "instagram") || strings.Contains(linkLower, "contact") ||
			strings.Contains(linkLower, "about") || strings.Contains(linkLower, "privacy") {
			score = -100
		}

		// ÃËœÃÂ³ÃÂ½ÃÂ¾Ã‘â‚¬ÃÂ¸Ã‘â‚¬Ã‘Æ’ÃÂµÃÂ¼ Ã‘ÂÃÂºÃÂ¾Ã‘â‚¬Ã‘Â ÃÂ¸ ÃÂ¿Ã‘Æ’Ã‘ÂÃ‘â€šÃ‘â€¹ÃÂµ Ã‘ÂÃ‘ÂÃ‘â€¹ÃÂ»ÃÂºÃÂ¸
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

// fetchHTML Ã‘ÂÃÂºÃÂ°Ã‘â€¡ÃÂ¸ÃÂ²ÃÂ°ÃÂµÃ‘â€š HTML Ã‘ÂÃ‘â€šÃ‘â‚¬ÃÂ°ÃÂ½ÃÂ¸Ã‘â€ Ã‘â€¹
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

// validateSelectors ÃÂ¿Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃ‘â€š, Ã‘â€¡Ã‘â€šÃÂ¾ Ã‘ÂÃÂµÃÂ»ÃÂµÃÂºÃ‘â€šÃÂ¾Ã‘â‚¬Ã‘â€¹ Ã‘â‚¬ÃÂ°ÃÂ±ÃÂ¾Ã‘â€šÃÂ°Ã‘Å½Ã‘â€š ÃÂ¸ ÃÂ¸ÃÂ·ÃÂ²ÃÂ»ÃÂµÃÂºÃÂ°Ã‘Å½Ã‘â€š ÃÂ´ÃÂ°ÃÂ½ÃÂ½Ã‘â€¹ÃÂµ
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

	// ÃÅ¸Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃÂ¼, Ã‘â€¡Ã‘â€šÃÂ¾ ÃÂ´ÃÂ°ÃÂ½ÃÂ½Ã‘â€¹ÃÂµ ÃÂ¸ÃÂ·ÃÂ²ÃÂ»ÃÂµÃ‘â€¡ÃÂµÃÂ½Ã‘â€¹
	if name == "" {
		return fmt.Errorf("name selector '%s' did not extract data", nameSel)
	}
	if price == "" {
		return fmt.Errorf("price selector '%s' did not extract data", priceSel)
	}

	// ÃÅ¸Ã‘â‚¬ÃÂ¾ÃÂ²ÃÂµÃ‘â‚¬Ã‘ÂÃÂµÃÂ¼, Ã‘â€¡Ã‘â€šÃÂ¾ Ã‘â€ ÃÂµÃÂ½ÃÂ° Ã‘ÂÃÂ¾ÃÂ´ÃÂµÃ‘â‚¬ÃÂ¶ÃÂ¸Ã‘â€š Ã‘â€¡ÃÂ¸Ã‘ÂÃÂ»ÃÂ°
	if !strings.ContainsAny(price, "0123456789") {
		return fmt.Errorf("price selector extracted non-numeric value: '%s'", price)
	}

	s.log.Info("Validation successful", map[string]interface{}{
		"name":  name,
		"price": price,
	})

	return nil
}
