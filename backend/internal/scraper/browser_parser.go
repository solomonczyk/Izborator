package scraper

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// ParseProductWithBrowser парсит страницу с использованием headless браузера (для JS-страниц)
func (s *Service) ParseProductWithBrowser(ctx context.Context, url string, shopConfig *ShopConfig) (*RawProduct, error) {
	s.logger.Info("Starting browser scraping", map[string]interface{}{
		"url":     url,
		"shop_id": shopConfig.ID,
	})

	var product RawProduct
	product.URL = url
	product.ShopID = shopConfig.ID
	product.ShopName = shopConfig.Name

	// Извлекаем external_id из URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		product.ExternalID = parts[len(parts)-1]
	}

	// Запускаем браузер с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Пробуем использовать существующий Chrome/Chromium вместо скачивания
	// Это обходит проблему с Windows Defender, блокирующим leakless
	l := launcher.New().
		Headless(true).
		NoSandbox(true).
		Leakless(false) // Отключаем leakless для обхода блокировки антивирусом

	// Пробуем найти Chrome в стандартных местах
	chromePaths := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Users\` + os.Getenv("USERNAME") + `\AppData\Local\Google\Chrome\Application\chrome.exe`,
	}

	var chromeFound bool
	for _, path := range chromePaths {
		if _, err := os.Stat(path); err == nil {
			l = l.Bin(path)
			chromeFound = true
			s.logger.Debug("Using existing Chrome", map[string]interface{}{
				"path": path,
			})
			break
		}
	}

	if !chromeFound {
		s.logger.Info("Chrome not found in standard paths, will download Chromium", map[string]interface{}{})
	}

	s.logger.Debug("Launching browser", map[string]interface{}{})
	browserURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().Context(ctx).ControlURL(browserURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := browser.Context(closeCtx).Close(); err != nil {
			// Игнорируем ошибки закрытия, если контекст истек
			if closeCtx.Err() == nil {
				s.logger.Warn("Failed to close browser", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}()

	// Создаем страницу
	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := page.Context(closeCtx).Close(); err != nil {
			// Игнорируем ошибки закрытия, если контекст истек
			if closeCtx.Err() == nil {
				s.logger.Warn("Failed to close page", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}()

	// Устанавливаем User-Agent
	page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	})

	// Переходим на страницу
	s.logger.Debug("Navigating to page", map[string]interface{}{
		"url": url,
	})

	err = page.Navigate(url)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate: %w", err)
	}

	// Ждем загрузки страницы
	page.WaitLoad()

	// Дополнительная задержка для загрузки JS-контента
	time.Sleep(3 * time.Second)

	// Получаем селекторы
	nameSelector := shopConfig.Selectors["name"]
	priceSelector := shopConfig.Selectors["price"]
	imageSelector := shopConfig.Selectors["image"]
	descriptionSelector := shopConfig.Selectors["description"]
	categorySelector := shopConfig.Selectors["category"]
	brandSelector := shopConfig.Selectors["brand"]

	// Парсинг названия
	if nameSelector != "" {
		selectors := strings.Split(nameSelector, ",")
		for _, sel := range selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			elem, err := page.Element(sel)
			if err == nil {
				text, err := elem.Text()
				if err == nil && text != "" && product.Name == "" {
					product.Name = strings.TrimSpace(text)
					s.logger.Debug("Found name from browser", map[string]interface{}{
						"name":     product.Name,
						"selector": sel,
					})
					break
				}
			}
		}
	}

	// Fallback: парсинг из title
	if product.Name == "" {
		titleElem, err := page.Element("title")
		if err == nil {
			title, err := titleElem.Text()
			if err == nil && title != "" {
				title = strings.TrimSuffix(title, " | Tehnomanija")
				title = strings.TrimSpace(title)
				if title != "" {
					product.Name = title
					s.logger.Debug("Found name from title", map[string]interface{}{
						"name": product.Name,
					})
				}
			}
		}
	}

	// Парсинг цены из JSON-LD
	jsonLDScripts, err := page.Elements("script[type='application/ld+json']")
	if err == nil {
		for _, script := range jsonLDScripts {
			text, err := script.Text()
			if err == nil && strings.Contains(text, `"@type":"Product"`) {
				// Ищем цену в JSON-LD
				offersIdx := strings.Index(text, `"offers"`)
				if offersIdx >= 0 {
					afterOffers := text[offersIdx:]
					priceIdx := strings.Index(afterOffers, `"price":`)
					if priceIdx > 0 {
						afterPrice := afterOffers[priceIdx+8:]
						endIdx := strings.IndexAny(afterPrice, ",}\n\r\t ")
						if endIdx > 0 {
							priceStr := strings.TrimSpace(afterPrice[:endIdx])
							if price, err := strconv.ParseFloat(priceStr, 64); err == nil && price > 0 && price < 10000000 {
								product.Price = price
								product.Currency = "RSD"
								s.logger.Debug("Found price from JSON-LD (browser)", map[string]interface{}{
									"price":    price,
									"currency": product.Currency,
								})
								break
							}
						}
					}
				}
			}
		}
	}

	// Парсинг цены из селекторов
	if product.Price == 0 && priceSelector != "" {
		selectors := strings.Split(priceSelector, ",")
		for _, sel := range selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			elem, err := page.Element(sel)
			if err == nil {
				text, err := elem.Text()
				if err == nil && text != "" {
					price, currency, err := cleanPrice(text)
					if err == nil && price > 0 {
						product.Price = price
						product.Currency = currency
						s.logger.Debug("Found price from browser", map[string]interface{}{
							"price":    price,
							"currency": currency,
							"selector": sel,
						})
						break
					}
				}
			}
		}
	}

	// Парсинг изображений
	if imageSelector != "" {
		selectors := strings.Split(imageSelector, ",")
		for _, sel := range selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			elems, err := page.Elements(sel)
			if err == nil {
				for _, elem := range elems {
					img, err := elem.Attribute("src")
					if err != nil || img == nil || *img == "" {
						img, err = elem.Attribute("data-src")
					}
					if err == nil && img != nil && *img != "" {
						imgURL := *img
						if !strings.HasPrefix(imgURL, "http") {
							imgURL = shopConfig.BaseURL + imgURL
						}
						// Проверяем, нет ли уже такого URL
						found := false
						for _, existing := range product.ImageURLs {
							if existing == imgURL {
								found = true
								break
							}
						}
						if !found {
							product.ImageURLs = append(product.ImageURLs, imgURL)
						}
					}
				}
			}
		}
	}

	// Парсинг описания
	if descriptionSelector != "" {
		selectors := strings.Split(descriptionSelector, ",")
		for _, sel := range selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			elem, err := page.Element(sel)
			if err == nil {
				text, err := elem.Text()
				if err == nil && text != "" && product.Description == "" {
					product.Description = strings.TrimSpace(text)
					break
				}
			}
		}
	}

	// Парсинг категории
	if categorySelector != "" {
		selectors := strings.Split(categorySelector, ",")
		for _, sel := range selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			elem, err := page.Element(sel)
			if err == nil {
				text, err := elem.Text()
				if err == nil && text != "" && product.Category == "" {
					product.Category = strings.TrimSpace(text)
					break
				}
			}
		}
	}

	// Парсинг бренда
	if brandSelector != "" {
		selectors := strings.Split(brandSelector, ",")
		for _, sel := range selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			elem, err := page.Element(sel)
			if err == nil {
				text, err := elem.Text()
				if err == nil && text != "" && product.Brand == "" {
					product.Brand = strings.TrimSpace(text)
					break
				}
			}
		}
	}

	// Валидация результата
	if product.Name == "" || product.Price == 0 {
		return nil, fmt.Errorf("failed to extract essential data from %s: name='%s', price=%.2f", url, product.Name, product.Price)
	}

	product.ParsedAt = time.Now()
	product.ScrapedAt = product.ParsedAt
	product.InStock = true

	s.logger.Info("Browser parsing completed", map[string]interface{}{
		"name":        product.Name,
		"price":       product.Price,
		"currency":    product.Currency,
		"brand":       product.Brand,
		"category":    product.Category,
		"description": product.Description,
		"images":      len(product.ImageURLs),
	})

	return &product, nil
}

