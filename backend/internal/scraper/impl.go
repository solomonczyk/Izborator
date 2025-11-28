package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

// ParseProduct скачивает страницу и извлекает данные по селекторам
func (s *Service) ParseProduct(ctx context.Context, url string, shopConfig *ShopConfig) (*RawProduct, error) {
	s.logger.Info("Starting scraping", map[string]interface{}{
		"url":     url,
		"shop_id": shopConfig.ID,
	})

	var product RawProduct
	product.URL = url
	product.ShopID = shopConfig.ID
	product.ShopName = shopConfig.Name
	
	// Извлекаем external_id из URL (последняя часть после последнего слеша)
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		product.ExternalID = parts[len(parts)-1]
	}

	// Инициализация Colly
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)

	// Настройка тайм-аутов
	c.SetRequestTimeout(30 * time.Second)

	// Рандомизация User-Agent и Referer (защита от бана)
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	// Получаем селекторы из конфига
	nameSelector := shopConfig.Selectors["name"]
	priceSelector := shopConfig.Selectors["price"]
	imageSelector := shopConfig.Selectors["image"]
	descriptionSelector := shopConfig.Selectors["description"]
	categorySelector := shopConfig.Selectors["category"]
	brandSelector := shopConfig.Selectors["brand"]

	// 0. Парсинг цены из JSON-LD (schema.org) - приоритетный метод, выполняется первым
	// На странице может быть несколько JSON-LD блоков в одном script теге
	c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		s.logger.Debug("Found JSON-LD script tag", map[string]interface{}{
			"text_length": len(e.Text),
		})
		if product.Price == 0 {
			jsonText := strings.TrimSpace(e.Text)
			
			// Сначала пробуем найти "price":число в offers напрямую (fallback метод)
			// Ищем паттерн "offers" -> "price":число
			offersIdx := strings.Index(jsonText, `"offers"`)
			if offersIdx >= 0 {
				afterOffers := jsonText[offersIdx:]
				priceIdx := strings.Index(afterOffers, `"price":`)
				if priceIdx > 0 {
					// Ищем число после "price":
					afterPrice := afterOffers[priceIdx+8:]
					// Ищем конец числа (запятая, скобка, пробел, новая строка)
					endIdx := strings.IndexAny(afterPrice, ",}\n\r\t ")
					if endIdx > 0 {
						priceStr := strings.TrimSpace(afterPrice[:endIdx])
						if price, err := strconv.ParseFloat(priceStr, 64); err == nil && price > 0 && price < 10000000 {
							// Проверяем, что цена разумная (меньше 10 миллионов RSD)
							product.Price = price
							product.Currency = "RSD"
							s.logger.Debug("Found price from JSON-LD (fallback)", map[string]interface{}{
								"price":    price,
								"currency": product.Currency,
							})
							return
						}
					}
				}
			}
			
			// Пробуем распарсить как объект
			var schemaData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonText), &schemaData); err == nil {
				// Проверяем, что это Product
				if schemaType, ok := schemaData["@type"].(string); ok && schemaType == "Product" {
					// Ищем offers.price в schema.org данных
					if offers, ok := schemaData["offers"].(map[string]interface{}); ok {
						// price может быть float64
						var price float64
						if priceFloat, ok := offers["price"].(float64); ok {
							price = priceFloat
						}
						if price > 0 && price < 10000000 {
							product.Price = price
							if currency, ok := offers["priceCurrency"].(string); ok {
								product.Currency = currency
							} else {
								product.Currency = "RSD"
							}
							s.logger.Debug("Found price from JSON-LD", map[string]interface{}{
								"price":    price,
								"currency": product.Currency,
							})
						}
					}
				}
			}
		}
	})

	// 1. Парсинг Названия
	if nameSelector != "" {
		c.OnHTML(nameSelector, func(e *colly.HTMLElement) {
			if product.Name == "" {
				product.Name = strings.TrimSpace(e.Text)
				s.logger.Debug("Found name", map[string]interface{}{
					"name":     product.Name,
					"selector": nameSelector,
				})
			}
		})
	}

	// 2. Парсинг Цены
	if priceSelector != "" {
		c.OnHTML(priceSelector, func(e *colly.HTMLElement) {
			if product.Price == 0 {
				rawPrice := strings.TrimSpace(e.Text)
				s.logger.Debug("Found price text", map[string]interface{}{
					"raw":     rawPrice,
					"selector": priceSelector,
				})
				price, currency, err := cleanPrice(rawPrice)
				if err == nil {
					product.Price = price
					product.Currency = currency
					s.logger.Debug("Parsed price", map[string]interface{}{
						"price":    price,
						"currency": currency,
					})
				} else {
					s.logger.Warn("Failed to parse price", map[string]interface{}{
						"raw":     rawPrice,
						"error":   err.Error(),
						"selector": priceSelector,
					})
				}
			}
		})
	}

	// 3. Парсинг Картинки
	if imageSelector != "" {
		c.OnHTML(imageSelector, func(e *colly.HTMLElement) {
			img := e.Attr("src")
			if img == "" {
				img = e.Attr("data-src") // Для lazy loading
			}
			if img != "" {
				if !strings.HasPrefix(img, "http") {
					img = e.Request.AbsoluteURL(img)
				}
				// Добавляем в массив изображений
				if len(product.ImageURLs) == 0 {
					product.ImageURLs = []string{img}
				} else {
					// Проверяем, нет ли уже такого URL
					found := false
					for _, existing := range product.ImageURLs {
						if existing == img {
							found = true
							break
						}
					}
					if !found {
						product.ImageURLs = append(product.ImageURLs, img)
					}
				}
			}
		})
	}

	// 4. Парсинг Описания
	if descriptionSelector != "" {
		c.OnHTML(descriptionSelector, func(e *colly.HTMLElement) {
			if product.Description == "" {
				product.Description = strings.TrimSpace(e.Text)
			}
		})
	}

	// 5. Парсинг Категории
	if categorySelector != "" {
		c.OnHTML(categorySelector, func(e *colly.HTMLElement) {
			if product.Category == "" {
				product.Category = strings.TrimSpace(e.Text)
			}
		})
	}

	// 6. Парсинг Бренда
	if brandSelector != "" {
		c.OnHTML(brandSelector, func(e *colly.HTMLElement) {
			if product.Brand == "" {
				product.Brand = strings.TrimSpace(e.Text)
			}
		})
	}

	// Парсинг цены из JSON-LD (schema.org) - приоритетный метод
	// На странице может быть несколько JSON-LD блоков в одном script теге
	c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		s.logger.Debug("Found JSON-LD script tag", map[string]interface{}{
			"text_length": len(e.Text),
		})
		if product.Price == 0 {
			jsonText := strings.TrimSpace(e.Text)
			
			// Сначала пробуем найти "price":число в offers напрямую (fallback метод)
			// Ищем паттерн "offers" -> "price":число
			offersIdx := strings.Index(jsonText, `"offers"`)
			s.logger.Debug("Checking JSON-LD for price", map[string]interface{}{
				"has_offers": offersIdx >= 0,
				"offers_idx": offersIdx,
				"text_preview": func() string {
					if len(jsonText) > 200 {
						return jsonText[:200] + "..."
					}
					return jsonText
				}(),
			})
			if offersIdx >= 0 {
				afterOffers := jsonText[offersIdx:]
				priceIdx := strings.Index(afterOffers, `"price":`)
				if priceIdx > 0 {
					// Ищем число после "price":
					afterPrice := afterOffers[priceIdx+8:]
					// Ищем конец числа (запятая, скобка, пробел, новая строка)
					endIdx := strings.IndexAny(afterPrice, ",}\n\r\t ")
					if endIdx > 0 {
						priceStr := strings.TrimSpace(afterPrice[:endIdx])
						s.logger.Debug("Trying to parse price from JSON-LD", map[string]interface{}{
							"price_str": priceStr,
							"offers_idx": offersIdx,
							"price_idx": priceIdx,
						})
						if price, err := strconv.ParseFloat(priceStr, 64); err == nil && price > 0 && price < 10000000 {
							// Проверяем, что цена разумная (меньше 10 миллионов RSD)
							product.Price = price
							product.Currency = "RSD"
							s.logger.Debug("Found price from JSON-LD (fallback)", map[string]interface{}{
								"price":    price,
								"currency": product.Currency,
							})
							return
						}
					}
				}
			}
			
			// Пробуем распарсить как объект
			var schemaData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonText), &schemaData); err == nil {
				// Проверяем, что это Product
				if schemaType, ok := schemaData["@type"].(string); ok && schemaType == "Product" {
					// Ищем offers.price в schema.org данных
					if offers, ok := schemaData["offers"].(map[string]interface{}); ok {
						// price может быть float64
						var price float64
						if priceFloat, ok := offers["price"].(float64); ok {
							price = priceFloat
						}
						if price > 0 && price < 10000000 {
							product.Price = price
							if currency, ok := offers["priceCurrency"].(string); ok {
								product.Currency = currency
							} else {
								product.Currency = "RSD"
							}
							s.logger.Debug("Found price from JSON-LD", map[string]interface{}{
								"price":    price,
								"currency": product.Currency,
							})
						}
					}
				}
			}
		}
	})

	// Логирование запроса
	c.OnRequest(func(r *colly.Request) {
		s.logger.Debug("Visiting", map[string]interface{}{
			"url": r.URL.String(),
		})
	})

	// Обработка ошибок
	c.OnError(func(r *colly.Response, err error) {
		s.logger.Error("Scraping failed", map[string]interface{}{
			"url":    r.Request.URL.String(),
			"error":  err.Error(),
			"status": r.StatusCode,
		})
	})

	// Запуск
	err := c.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("colly visit error: %w", err)
	}

	// Логируем что было найдено
	s.logger.Debug("Parsing completed", map[string]interface{}{
		"name":        product.Name,
		"price":       product.Price,
		"currency":    product.Currency,
		"brand":       product.Brand,
		"category":    product.Category,
		"description": product.Description,
		"images":      len(product.ImageURLs),
	})

	// Валидация результата
	if product.Name == "" || product.Price == 0 {
		return nil, fmt.Errorf("failed to extract essential data from %s: name='%s', price=%.2f", url, product.Name, product.Price)
	}

	product.ParsedAt = time.Now()
	product.ScrapedAt = product.ParsedAt // для обратной совместимости
	product.InStock = true // По умолчанию считаем, что товар в наличии

	return &product, nil
}

// cleanPrice превращает строку "120.000 RSD" -> (120000.0, "RSD", nil)
// Ищет паттерн "число.число RSD" в тексте
func cleanPrice(raw string) (float64, string, error) {
	// Ищем паттерн "число.число RSD" или "число RSD"
	// Примеры: "15.999 RSD", "16.999 RSD", "1.000.000 RSD"
	
	// Сначала пробуем найти паттерн с точками и RSD
	text := strings.ToUpper(raw)
	
	// Ищем "число.число RSD" - самый распространённый формат
	// Паттерн: 1-3 цифры, точка, 3 цифры, пробел, RSD
	idx := strings.Index(text, "RSD")
	if idx > 0 {
		// Ищем число перед RSD
		beforeRSD := text[:idx]
		// Ищем последнее число с точкой перед RSD
		parts := strings.Fields(beforeRSD)
		for i := len(parts) - 1; i >= 0; i-- {
			part := strings.TrimSpace(parts[i])
			// Проверяем, содержит ли часть точки и цифры
			if strings.Contains(part, ".") && strings.ContainsAny(part, "0123456789") {
				// Убираем все точки и пробуем распарсить
				clean := strings.ReplaceAll(part, ".", "")
				if val, err := strconv.ParseFloat(clean, 64); err == nil && val > 0 {
					return val, "RSD", nil
				}
			}
		}
	}
	
	// Если не нашли, пробуем старый метод
	clean := strings.ReplaceAll(raw, ".", "")
	clean = strings.ReplaceAll(clean, "RSD", "")
	clean = strings.ReplaceAll(clean, "DIN", "")
	clean = strings.TrimSpace(clean)
	
	// Убираем все нецифровые символы
	var builder strings.Builder
	for _, r := range clean {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	clean = builder.String()
	
	if clean == "" {
		return 0, "", fmt.Errorf("no numbers found in price string: '%s'", raw)
	}

	val, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse price '%s' (cleaned: '%s'): %w", raw, clean, err)
	}

	return val, "RSD", nil
}

// ParseProductByShopID парсит товар по shopID (получает конфиг из storage)
func (s *Service) ParseProductByShopID(ctx context.Context, url string, shopID string) (*RawProduct, error) {
	if url == "" {
		return nil, ErrInvalidURL
	}

	config, err := s.storage.GetShopConfig(shopID)
	if err != nil {
		s.logger.Error("Failed to get shop config", map[string]interface{}{
			"error":  err,
			"shop_id": shopID,
		})
		return nil, ErrShopNotFound
	}

	if !config.Enabled {
		return nil, ErrShopDisabled
	}

	return s.ParseProduct(ctx, url, config)
}

// SaveRawProduct сохраняет сырые данные товара
func (s *Service) SaveRawProduct(ctx context.Context, product *RawProduct) error {
	if product == nil {
		return fmt.Errorf("product is nil")
	}

	if err := s.storage.SaveRawProduct(product); err != nil {
		s.logger.Error("Failed to save raw product", map[string]interface{}{
			"error":  err,
			"shop_id": product.ShopID,
		})
		return fmt.Errorf("failed to save raw product: %w", err)
	}

	// Отправляем в очередь для дальнейшей обработки
	if s.queue != nil {
		if err := s.queue.Publish("scraping_tasks", product); err != nil {
			s.logger.Error("Failed to publish to queue", map[string]interface{}{
				"error": err,
			})
			// Не возвращаем ошибку, так как данные уже сохранены
		}
	}

	return nil
}

// ListShops получает список всех магазинов
func (s *Service) ListShops(ctx context.Context) ([]*ShopConfig, error) {
	shops, err := s.storage.ListShops()
	if err != nil {
		s.logger.Error("Failed to list shops", map[string]interface{}{
			"error": err,
		})
		return nil, fmt.Errorf("failed to list shops: %w", err)
	}

	return shops, nil
}
