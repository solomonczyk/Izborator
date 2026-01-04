package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/solomonczyk/izborator/internal/scrapingstats"
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

	// Настройка тайм-аутов (увеличено для медленных соединений)
	c.SetRequestTimeout(60 * time.Second)

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

	s.logger.Debug("Loaded selectors", map[string]interface{}{
		"name":        nameSelector,
		"price":       priceSelector,
		"image":       imageSelector,
		"description": descriptionSelector,
		"category":    categorySelector,
		"brand":       brandSelector,
	})

	// Логируем используемые селекторы
	s.logger.Debug("Using selectors", map[string]interface{}{
		"name":        nameSelector,
		"price":       priceSelector,
		"image":       imageSelector,
		"description": descriptionSelector,
		"category":    categorySelector,
		"brand":       brandSelector,
	})

	// 0. Парсинг цены из JSON-LD (schema.org) - приоритетный метод, выполняется первым
	// На странице может быть несколько JSON-LD блоков в одном script теге
	c.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		s.logger.Debug("Found JSON-LD script tag", map[string]interface{}{
			"text_length": len(e.Text),
		})
		if product.Price == 0 {
			jsonText := strings.TrimSpace(e.Text)

			// Пробуем распарсить как объект
			var schemaData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonText), &schemaData); err == nil {
				// Проверяем, что это Product
				if schemaType, ok := schemaData["@type"].(string); ok && schemaType == "Product" {
					// Ищем offers.price в schema.org данных
					if offers, ok := schemaData["offers"].(map[string]interface{}); ok {
						// price может быть float64 или string
						var price float64
						if priceFloat, ok := offers["price"].(float64); ok {
							price = priceFloat
						} else if priceStr, ok := offers["price"].(string); ok {
							if parsedPrice, err := strconv.ParseFloat(priceStr, 64); err == nil {
								price = parsedPrice
							}
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
					// Если offers это массив, берем первый элемент
					if offersArr, ok := schemaData["offers"].([]interface{}); ok && len(offersArr) > 0 {
						if offer, ok := offersArr[0].(map[string]interface{}); ok {
							var price float64
							if priceFloat, ok := offer["price"].(float64); ok {
								price = priceFloat
							} else if priceStr, ok := offer["price"].(string); ok {
								if parsedPrice, err := strconv.ParseFloat(priceStr, 64); err == nil {
									price = parsedPrice
								}
							}
							if price > 0 && price < 10000000 {
								product.Price = price
								if currency, ok := offer["priceCurrency"].(string); ok {
									product.Currency = currency
								} else {
									product.Currency = "RSD"
								}
								s.logger.Debug("Found price from JSON-LD (array)", map[string]interface{}{
									"price":    price,
									"currency": product.Currency,
								})
							}
						}
					}
				}
			} else {
				// Fallback: ищем цену в тексте JSON напрямую
				if product.Price == 0 {
					offersIdx := strings.Index(jsonText, `"offers"`)
					if offersIdx >= 0 {
						afterOffers := jsonText[offersIdx:]
						priceIdx := strings.Index(afterOffers, `"price":`)
						if priceIdx > 0 {
							afterPrice := afterOffers[priceIdx+8:]
							endIdx := strings.IndexAny(afterPrice, ",}\n\r\t ")
							if endIdx > 0 {
								priceStr := strings.TrimSpace(afterPrice[:endIdx])
								if price, err := strconv.ParseFloat(priceStr, 64); err == nil && price > 0 && price < 10000000 {
									product.Price = price
									product.Currency = "RSD"
									s.logger.Debug("Found price from JSON-LD (fallback)", map[string]interface{}{
										"price":    price,
										"currency": product.Currency,
									})
								}
							}
						}
					}
				}
			}
		}
	})

	// 1. Парсинг Названия
	if nameSelector != "" {
		// Пробуем каждый селектор из списка (разделены запятыми)
		selectors := strings.Split(nameSelector, ",")
		for _, sel := range selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			c.OnHTML(sel, func(e *colly.HTMLElement) {
				if product.Name == "" {
					product.Name = strings.TrimSpace(e.Text)
					s.logger.Debug("Found name", map[string]interface{}{
						"name":     product.Name,
						"selector": sel,
					})
				}
			})
		}
	}

	// Fallback: парсинг названия из title страницы
	c.OnHTML("title", func(e *colly.HTMLElement) {
		if product.Name == "" {
			title := strings.TrimSpace(e.Text)
			// Убираем " | Tehnomanija" из конца
			if strings.Contains(title, " | Tehnomanija") {
				title = strings.Split(title, " | Tehnomanija")[0]
			}
			product.Name = title
			s.logger.Debug("Found name from title", map[string]interface{}{
				"name": product.Name,
			})
		}
	})

	// 2. Парсинг Цены
	if priceSelector != "" {
		// Пробуем каждый селектор из списка (разделены запятыми)
		selectors := strings.Split(priceSelector, ",")
		for _, sel := range selectors {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}
			c.OnHTML(sel, func(e *colly.HTMLElement) {
				if product.Price == 0 {
					rawPrice := strings.TrimSpace(e.Text)
					s.logger.Debug("Found price text", map[string]interface{}{
						"raw":      rawPrice,
						"selector": sel,
					})
					price, currency, err := cleanPrice(rawPrice)
					if err == nil {
						product.Price = price
						product.Currency = currency
						s.logger.Debug("Parsed price", map[string]interface{}{
							"price":    price,
							"currency": currency,
							"selector": sel,
						})
					} else {
						s.logger.Warn("Failed to parse price", map[string]interface{}{
							"raw":      rawPrice,
							"error":    err.Error(),
							"selector": sel,
						})
					}
				}
			})
		}
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
							"price_str":  priceStr,
							"offers_idx": offersIdx,
							"price_idx":  priceIdx,
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

	// Логирование запроса и настройка заголовков
	c.OnRequest(func(r *colly.Request) {
		s.logger.Debug("Visiting", map[string]interface{}{
			"url": r.URL.String(),
		})

		// Добавляем заголовки для обхода защиты от ботов
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("Accept-Language", "sr-RS,sr;q=0.9,en-US;q=0.8,en;q=0.7")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Cache-Control", "max-age=0")

		// Если Referer не установлен расширением, устанавливаем базовый URL магазина
		if r.Headers.Get("Referer") == "" {
			r.Headers.Set("Referer", shopConfig.BaseURL+"/")
		}
	})

	// Временное логирование HTML для отладки селекторов (только для tehnomanija)
	c.OnResponse(func(r *colly.Response) {
		if shopConfig.ID == "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22" {
			htmlLen := len(r.Body)
			s.logger.Info("Response received", map[string]interface{}{
				"url":         r.Request.URL.String(),
				"status":      r.StatusCode,
				"html_length": htmlLen,
			})

			// Сохраняем HTML в файл для анализа
			htmlStr := string(r.Body)
			if err := os.WriteFile("tehnomanija_debug.html", []byte(htmlStr), 0644); err == nil {
				s.logger.Info("HTML saved to tehnomanija_debug.html", map[string]interface{}{})
			}

			// Ищем название и цену в HTML напрямую для отладки
			if strings.Contains(htmlStr, "Dell Laptop XPS") {
				idx := strings.Index(htmlStr, "Dell Laptop XPS")
				start := idx - 200
				if start < 0 {
					start = 0
				}
				end := idx + 300
				if end > len(htmlStr) {
					end = len(htmlStr)
				}
				context := htmlStr[start:end]
				s.logger.Debug("Found product name in HTML", map[string]interface{}{
					"context": context,
				})
			}

			// Ищем цену в HTML
			if strings.Contains(htmlStr, "RSD") || strings.Contains(htmlStr, "din") {
				// Ищем паттерн цены
				rsdIdx := strings.Index(htmlStr, "RSD")
				if rsdIdx > 0 {
					start := rsdIdx - 50
					if start < 0 {
						start = 0
					}
					end := rsdIdx + 10
					if end > len(htmlStr) {
						end = len(htmlStr)
					}
					context := htmlStr[start:end]
					s.logger.Debug("Found price indicator in HTML", map[string]interface{}{
						"context": context,
					})
				}
			}
		}
	})

	// Обработка ошибок
	c.OnError(func(r *colly.Response, err error) {
		s.logger.Error("Scraping failed", map[string]interface{}{
			"url":    r.Request.URL.String(),
			"error":  err.Error(),
			"status": r.StatusCode,
		})
	})

	// Сохраняем HTML для отладки (только для tehnomanija)
	var savedHTML []byte
	if shopConfig.ID == "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22" {
		c.OnResponse(func(r *colly.Response) {
			savedHTML = r.Body
			// Сохраняем в файл для анализа
			_ = os.WriteFile("tehnomanija_debug.html", r.Body, 0644)
			s.logger.Info("Saved HTML to tehnomanija_debug.html", map[string]interface{}{
				"size": len(r.Body),
			})
		})
	}

	// Запуск
	err := c.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("colly visit error: %w", err)
	}

	// Сохраняем HTML для анализа (только для tehnomanija)
	if shopConfig.ID == "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22" && len(savedHTML) > 0 {
		htmlStr := string(savedHTML)
		// Ищем название в HTML напрямую
		if product.Name == "" {
			// Пробуем найти в title
			if titleIdx := strings.Index(htmlStr, "<title>"); titleIdx >= 0 {
				titleEnd := strings.Index(htmlStr[titleIdx:], "</title>")
				if titleEnd > 0 {
					title := strings.TrimSpace(htmlStr[titleIdx+7 : titleIdx+titleEnd])
					title = strings.TrimSuffix(title, " | Tehnomanija")
					if title != "" {
						product.Name = title
						s.logger.Debug("Extracted name from HTML title tag", map[string]interface{}{
							"name": product.Name,
						})
					}
				}
			}
		}
		// Ищем цену в JSON-LD в HTML
		if product.Price == 0 {
			jsonLDIdx := strings.Index(htmlStr, `"@type":"Product"`)
			if jsonLDIdx >= 0 {
				// Ищем цену после "@type":"Product"
				priceIdx := strings.Index(htmlStr[jsonLDIdx:], `"price":`)
				if priceIdx > 0 {
					afterPrice := htmlStr[jsonLDIdx+priceIdx+8:]
					endIdx := strings.IndexAny(afterPrice, ",}\n\r\t ")
					if endIdx > 0 {
						priceStr := strings.TrimSpace(afterPrice[:endIdx])
						if price, err := strconv.ParseFloat(priceStr, 64); err == nil && price > 0 && price < 10000000 {
							product.Price = price
							product.Currency = "RSD"
							s.logger.Debug("Extracted price from HTML JSON-LD", map[string]interface{}{
								"price":    price,
								"currency": product.Currency,
							})
						}
					}
				}
			}
		}
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
	product.InStock = true               // По умолчанию считаем, что товар в наличии

	return &product, nil
}

// cleanPrice превращает строку "120.000 RSD" -> (120000.0, "RSD", nil)
// Ищет паттерн "число.число RSD" в тексте
func cleanPrice(raw string) (float64, string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return 0, "", fmt.Errorf("empty price string")
	}

	upper := strings.ToUpper(text)
	currency := "RSD"
	if strings.Contains(upper, "EUR") {
		currency = "EUR"
	} else if strings.Contains(upper, "USD") {
		currency = "USD"
	} else if strings.Contains(upper, "DIN") || strings.Contains(upper, "RSD") {
		currency = "RSD"
	}

	var numBuilder strings.Builder
	for _, r := range text {
		if (r >= '0' && r <= '9') || r == '.' || r == ',' || r == ' ' {
			numBuilder.WriteRune(r)
		}
	}

	numStr := strings.ReplaceAll(numBuilder.String(), " ", "")
	if numStr == "" {
		return 0, "", fmt.Errorf("no numbers found in price string: '%s'", raw)
	}

	lastDot := strings.LastIndex(numStr, ".")
	lastComma := strings.LastIndex(numStr, ",")
	decimalSep := ""
	if lastDot >= 0 || lastComma >= 0 {
		if lastDot > lastComma {
			decimalSep = "."
		} else {
			decimalSep = ","
		}
	}

	var normalized string
	if decimalSep == "" {
		normalized = strings.ReplaceAll(strings.ReplaceAll(numStr, ".", ""), ",", "")
	} else {
		idx := strings.LastIndex(numStr, decimalSep)
		intPart := numStr[:idx]
		fracPart := numStr[idx+1:]
		intPart = strings.ReplaceAll(intPart, ".", "")
		intPart = strings.ReplaceAll(intPart, ",", "")
		fracPart = strings.ReplaceAll(fracPart, ".", "")
		fracPart = strings.ReplaceAll(fracPart, ",", "")
		if fracPart == "" {
			normalized = intPart
		} else {
			normalized = intPart + "." + fracPart
		}
	}

	if normalized == "" {
		return 0, "", fmt.Errorf("no numbers found in price string: '%s'", raw)
	}

	val, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse price '%s' (cleaned: '%s'): %w", raw, normalized, err)
	}

	return val, currency, nil
}


// recordScrapingStat сохраняет статистику парсинга, если сервис подключён
func (s *Service) recordScrapingStat(stat *scrapingstats.ScrapingStat) {
	if s.stats == nil || stat == nil {
		return
	}

	if stat.ScrapedAt.IsZero() {
		stat.ScrapedAt = time.Now()
	}
	stat.CreatedAt = time.Now()

	if stat.Status == "" {
		stat.Status = "error"
	}

	if err := s.stats.RecordStat(stat); err != nil {
		s.logger.Warn("Failed to record scraping stat", map[string]interface{}{
			"error":   err.Error(),
			"shop_id": stat.ShopID,
		})
	}
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	const maxLen = 512
	msg := err.Error()
	if len(msg) > maxLen {
		return msg[:maxLen] + "..."
	}
	return msg
}

func (s *Service) getRetryConfig(config *ShopConfig) (int, time.Duration) {
	limit := config.RetryLimit
	if limit <= 0 {
		limit = defaultRetryLimit
	}
	if limit > maxRetryLimit {
		limit = maxRetryLimit
	}
	backoffMs := config.RetryBackoffMs
	if backoffMs <= 0 {
		backoffMs = defaultRetryBackoffMs
	}
	if backoffMs > maxRetryBackoffMs {
		backoffMs = maxRetryBackoffMs
	}
	return limit, time.Duration(backoffMs) * time.Millisecond
}

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	maxBackoff := time.Duration(maxRetryBackoffMs) * time.Millisecond
	if next > maxBackoff {
		return maxBackoff
	}
	return next
}

// ParseProductByShopID парсит товар по shopID (получает конфиг из storage)
func (s *Service) ParseProductByShopID(ctx context.Context, url string, shopID string) (*RawProduct, error) {
	if url == "" {
		return nil, ErrInvalidURL
	}

	config, err := s.storage.GetShopConfig(shopID)
	if err != nil {
		s.logger.Error("Failed to get shop config", map[string]interface{}{
			"error":   err,
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
			"error":   err,
			"shop_id": product.ShopID,
		})
		return fmt.Errorf("failed to save raw product: %w", err)
	}

	// Отправляем в очередь для дальнейшей обработки
	if s.queue != nil {
		if err := s.queue.Publish(s.queueTopic, product); err != nil {
			s.logger.Error("Failed to publish to queue", map[string]interface{}{
				"error": err,
			})
			// Не возвращаем ошибку, так как данные уже сохранены
		}
	}

	return nil
}

// ScrapeAndSave выполняет полный цикл парсинга и сохранения товара с записью статистики
// Автоматически выбирает между обычным парсером (Colly) и browser парсером (rod) в зависимости от магазина
func (s *Service) ScrapeAndSave(ctx context.Context, url string, shopConfig *ShopConfig) (*RawProduct, error) {
	// Магазины, требующие JS-рендеринг (headless браузер)
	jsRenderingShops := map[string]bool{
		"b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22": true, // Tehnomanija
		"shop-001":                             true, // Gigatron
	}

	// Используем browser парсер для магазинов, требующих JS-рендеринг
	if jsRenderingShops[shopConfig.ID] {
		s.logger.Info("Using browser parser for JS-rendered page", map[string]interface{}{
			"shop_id":   shopConfig.ID,
			"shop_name": shopConfig.Name,
		})
		return s.scrapeAndSaveWithBrowser(ctx, url, shopConfig)
	}

	// Используем обычный парсер (Colly)
	return s.scrapeAndSaveWithColly(ctx, url, shopConfig)
}

// scrapeAndSaveWithColly использует Colly для парсинга
func (s *Service) scrapeAndSaveWithColly(ctx context.Context, url string, shopConfig *ShopConfig) (*RawProduct, error) {
	start := time.Now()
	stat := &scrapingstats.ScrapingStat{
		ShopID:    shopConfig.ID,
		ShopName:  shopConfig.Name,
		ScrapedAt: start,
		Status:    "error",
	}

	maxAttempts, initialBackoff := s.getRetryConfig(shopConfig)
	currentBackoff := initialBackoff
	status := "error"
	var lastErr error
	var rawProduct *RawProduct

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			lastErr = ctx.Err()
			break
		}

		s.logger.Info("Scraping attempt", map[string]interface{}{
			"attempt":      attempt,
			"max_attempts": maxAttempts,
			"shop_id":      shopConfig.ID,
			"url":          url,
		})

		rawProduct, lastErr = s.ParseProduct(ctx, url, shopConfig)
		if lastErr == nil {
			stat.ProductsFound = 1
			if err := s.SaveRawProduct(ctx, rawProduct); err == nil {
				status = "success"
				stat.ProductsSaved = 1
				stat.ErrorMessage = ""
				break
			} else {
				lastErr = err
				status = "partial"
			}
		}

		stat.ErrorsCount++
		stat.ErrorMessage = truncateError(lastErr)

		if attempt < maxAttempts {
			s.logger.Warn("Scraping attempt failed, retrying", map[string]interface{}{
				"attempt":      attempt,
				"max_attempts": maxAttempts,
				"shop_id":      shopConfig.ID,
				"error":        lastErr.Error(),
				"backoff_ms":   currentBackoff.Milliseconds(),
			})
			if !sleepWithContext(ctx, currentBackoff) {
				lastErr = ctx.Err()
				break
			}
			currentBackoff = nextBackoff(currentBackoff)
			continue
		}
	}

	if status != "success" && status != "partial" && stat.ErrorsCount > 0 {
		stat.ErrorMessage = truncateError(lastErr)
	}

	stat.Status = status
	stat.DurationMs = int(time.Since(start) / time.Millisecond)
	s.recordScrapingStat(stat)

	if status != "success" {
		if lastErr == nil {
			lastErr = fmt.Errorf("scraping failed after %d attempts", maxAttempts)
		}
		return nil, lastErr
	}

	return rawProduct, nil
}

// scrapeAndSaveWithBrowser использует rod (headless браузер) для парсинга JS-страниц
func (s *Service) scrapeAndSaveWithBrowser(ctx context.Context, url string, shopConfig *ShopConfig) (*RawProduct, error) {
	start := time.Now()
	stat := &scrapingstats.ScrapingStat{
		ShopID:    shopConfig.ID,
		ShopName:  shopConfig.Name,
		ScrapedAt: start,
		Status:    "error",
	}

	maxAttempts, initialBackoff := s.getRetryConfig(shopConfig)
	currentBackoff := initialBackoff
	status := "error"
	var lastErr error
	var rawProduct *RawProduct

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			lastErr = ctx.Err()
			break
		}

		s.logger.Info("Browser scraping attempt", map[string]interface{}{
			"attempt":      attempt,
			"max_attempts": maxAttempts,
			"shop_id":      shopConfig.ID,
			"url":          url,
		})

		rawProduct, lastErr = s.ParseProductWithBrowser(ctx, url, shopConfig)
		if lastErr == nil {
			stat.ProductsFound = 1
			if err := s.SaveRawProduct(ctx, rawProduct); err == nil {
				status = "success"
				stat.ProductsSaved = 1
				stat.ErrorMessage = ""
				break
			} else {
				lastErr = err
				status = "partial"
			}
		}

		stat.ErrorsCount++
		stat.ErrorMessage = truncateError(lastErr)

		if attempt < maxAttempts {
			s.logger.Warn("Browser scraping attempt failed, retrying", map[string]interface{}{
				"attempt":      attempt,
				"max_attempts": maxAttempts,
				"shop_id":      shopConfig.ID,
				"error":        lastErr.Error(),
				"backoff_ms":   currentBackoff.Milliseconds(),
			})
			if !sleepWithContext(ctx, currentBackoff) {
				lastErr = ctx.Err()
				break
			}
			currentBackoff = nextBackoff(currentBackoff)
			continue
		}
	}

	if status != "success" && status != "partial" && stat.ErrorsCount > 0 {
		stat.ErrorMessage = truncateError(lastErr)
	}

	stat.Status = status
	stat.DurationMs = int(time.Since(start) / time.Millisecond)
	s.recordScrapingStat(stat)

	if status != "success" {
		if lastErr == nil {
			lastErr = fmt.Errorf("browser scraping failed after %d attempts", maxAttempts)
		}
		return nil, lastErr
	}

	return rawProduct, nil
}

// ScrapeAndSaveByShopID выполняет цикл парсинга по shopID
func (s *Service) ScrapeAndSaveByShopID(ctx context.Context, url, shopID string) (*RawProduct, error) {
	if url == "" {
		return nil, ErrInvalidURL
	}

	config, err := s.storage.GetShopConfig(shopID)
	if err != nil {
		s.logger.Error("Failed to get shop config for scrape+save", map[string]interface{}{
			"error":   err,
			"shop_id": shopID,
		})
		return nil, ErrShopNotFound
	}

	if !config.Enabled {
		return nil, ErrShopDisabled
	}

	return s.ScrapeAndSave(ctx, url, config)
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

// ParseCatalog обходит каталог магазина и извлекает ссылки на товары
// catalogURL - начальный URL каталога (например, https://gigatron.rs/mobilni-telefoni)
// maxPages - максимальное количество страниц для обхода (0 = без ограничений)
func (s *Service) ParseCatalog(ctx context.Context, catalogURL string, shopConfig *ShopConfig, maxPages int) (*CatalogResult, error) {
	s.logger.Info("Starting catalog parsing", map[string]interface{}{
		"catalog_url": catalogURL,
		"shop_id":     shopConfig.ID,
		"max_pages":   maxPages,
	})

	result := &CatalogResult{
		ProductURLs: make([]string, 0),
	}

	// Селекторы для каталога
	productLinkSelector := shopConfig.Selectors["catalog_product_link"]
	if productLinkSelector == "" {
		// Универсальные дефолтные селекторы (работают для большинства магазинов)
		productLinkSelector = "a.product-box, .product-item a, .product-card a, .product-title a, .product a, article a, .item a"
	}

	nextPageSelector := shopConfig.Selectors["catalog_next_page"]
	if nextPageSelector == "" {
		// Универсальные селекторы для пагинации
		nextPageSelector = "a.next, .pagination .next, .pager .next, .pagination-next, a[rel=\"next\"], .pagination a:contains(\"Следующая\")"
	}

	// Инициализация Colly
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)

	c.SetRequestTimeout(60 * time.Second)
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	// Ограничение скорости (rate limiting)
	if shopConfig.RateLimit > 0 {
		_ = c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 1,
			Delay:       time.Duration(1000/shopConfig.RateLimit) * time.Millisecond,
		})
	}

	visitedPages := make(map[string]bool)
	pageCount := 0
	productURLsMap := make(map[string]bool) // Для дедупликации
	totalLinksFound := 0                    // Всего найдено ссылок
	filteredOutCount := 0                   // Отфильтровано isProductURL
	duplicateCount := 0                     // Дубликатов

	s.logger.Info("Using catalog selectors", map[string]interface{}{
		"product_link_selector": productLinkSelector,
		"next_page_selector":     nextPageSelector,
		"catalog_url":            catalogURL,
	})

	// Обработчик ссылок на товары
	c.OnHTML(productLinkSelector, func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if href == "" {
			return
		}

		totalLinksFound++

		// Преобразуем относительные URL в абсолютные
		productURL := href
		if strings.HasPrefix(href, "/") {
			productURL = shopConfig.BaseURL + href
		} else if !strings.HasPrefix(href, "http") {
			productURL = shopConfig.BaseURL + "/" + href
		}

		// Проверяем на дубликаты
		if productURLsMap[productURL] {
			duplicateCount++
			return
		}

		// Проверяем, что это URL товара (не категории или другой страницы)
		if !s.isProductURL(productURL, shopConfig) {
			filteredOutCount++
			return
		}

		productURLsMap[productURL] = true
		result.ProductURLs = append(result.ProductURLs, productURL)
	})

	// Обработчик ссылки на следующую страницу
	var nextPageURL string
	c.OnHTML(nextPageSelector, func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if href == "" {
			return
		}

		nextURL := href
		if strings.HasPrefix(href, "/") {
			nextURL = shopConfig.BaseURL + href
		} else if !strings.HasPrefix(href, "http") {
			nextURL = shopConfig.BaseURL + "/" + href
		}

		if !visitedPages[nextURL] && (maxPages == 0 || pageCount < maxPages) {
			nextPageURL = nextURL
		}
	})

	// Обработка ошибок
	c.OnError(func(r *colly.Response, err error) {
		s.logger.Error("Catalog parsing error", map[string]interface{}{
			"url":   r.Request.URL.String(),
			"error": err.Error(),
		})
	})

	// Обход страниц
	currentURL := catalogURL
	for {
		if maxPages > 0 && pageCount >= maxPages {
			break
		}

		if visitedPages[currentURL] {
			break
		}

		visitedPages[currentURL] = true
		pageCount++
		nextPageURL = "" // Сбрасываем перед каждой страницей

		// Сохраняем количество найденных ссылок до обработки страницы
		linksBeforePage := len(result.ProductURLs)
		
		s.logger.Info("Parsing catalog page", map[string]interface{}{
			"url":          currentURL,
			"page":         pageCount,
			"found_so_far": linksBeforePage,
		})

		err := c.Visit(currentURL)
		if err != nil {
			s.logger.Error("Failed to visit catalog page", map[string]interface{}{
				"url":   currentURL,
				"error": err.Error(),
			})
			break
		}

		// Логируем результаты обработки страницы
		linksAfterPage := len(result.ProductURLs)
		linksOnThisPage := linksAfterPage - linksBeforePage
		s.logger.Info("Page processed", map[string]interface{}{
			"page":            pageCount,
			"links_on_page":   linksOnThisPage,
			"total_links":     linksAfterPage,
			"next_page_found": nextPageURL != "",
		})

		// Если есть следующая страница, переходим к ней
		if nextPageURL != "" && !visitedPages[nextPageURL] {
			currentURL = nextPageURL
		} else {
			break
		}

		// Пауза между страницами
		time.Sleep(2 * time.Second)
	}

	result.TotalFound = len(result.ProductURLs)
	
	// Детальная статистика для диагностики
	s.logger.Info("Catalog parsing completed", map[string]interface{}{
		"total_urls":        result.TotalFound,
		"pages":            pageCount,
		"total_links_found": totalLinksFound,
		"filtered_out":     filteredOutCount,
		"duplicates":       duplicateCount,
		"shop_id":          shopConfig.ID,
		"shop_name":        shopConfig.Name,
	})

	// Логируем примеры найденных URL (первые 3) для диагностики
	if len(result.ProductURLs) > 0 {
		examples := result.ProductURLs
		if len(examples) > 3 {
			examples = examples[:3]
		}
		s.logger.Info("Example product URLs found", map[string]interface{}{
			"examples": examples,
		})
	} else {
		s.logger.Warn("No product URLs found - possible issues", map[string]interface{}{
			"total_links_found": totalLinksFound,
			"filtered_out":     filteredOutCount,
			"product_selector": productLinkSelector,
			"catalog_url":      catalogURL,
		})
	}

	return result, nil
}

// isProductURL проверяет, является ли URL ссылкой на товар (а не на категорию)
func (s *Service) isProductURL(url string, shopConfig *ShopConfig) bool {
	urlLower := strings.ToLower(url)

	// Проверка на главную страницу
	if shopConfig != nil {
		baseURL := strings.TrimSuffix(strings.ToLower(shopConfig.BaseURL), "/")
		if urlLower == baseURL || urlLower == baseURL+"/" {
			return false
		}
	}

	// Специальная обработка для Gigatron
	if shopConfig != nil && strings.Contains(strings.ToLower(shopConfig.BaseURL), "gigatron.rs") {
		baseURL := strings.TrimSuffix(strings.ToLower(shopConfig.BaseURL), "/")
		if strings.Contains(urlLower, "/kategorija/") {
			path := strings.TrimPrefix(urlLower, baseURL)
			parts := filterEmpty(strings.Split(path, "/"))
			// Для Gigatron товары обычно имеют минимум 3 части пути после /kategorija/
			return len(parts) >= 3
		}
	}

	// Расширенный список паттернов категорий (из AutoConfig)
	categoryPatterns := []string{
		"/kategorija/", "/kategorije/", "/category/", "/categories/",
		"/product-category/", "/product_category/",
		"/kategorija-proizvoda/", "/oznaka-proizvoda/",
		"/tag/", "/tags/", "/brend/", "/brand/",
		"/proizvodjac/", "/proizvodaci/", "/manufacturer/",
		"/shop/", "/store/", "/online-prodavnica/", "/prodavnica/",
		"/collections/", "/collection/",
	}

	for _, pattern := range categoryPatterns {
		if strings.Contains(urlLower, pattern) {
			// Дополнительная проверка: если после паттерна категории идет длинное название (товар)
			// Например: /kategorija/mobilni-telefoni/samsung-galaxy-s23-ultra-256gb
			idx := strings.Index(urlLower, pattern)
			if idx != -1 {
				afterPattern := urlLower[idx+len(pattern):]
				// Убираем параметры и якоря
				if paramIdx := strings.Index(afterPattern, "?"); paramIdx != -1 {
					afterPattern = afterPattern[:paramIdx]
				}
				if anchorIdx := strings.Index(afterPattern, "#"); anchorIdx != -1 {
					afterPattern = afterPattern[:anchorIdx]
				}
				// Убираем trailing slash
				afterPattern = strings.TrimSuffix(afterPattern, "/")
				
				// Если после паттерна категории идет длинное название (больше 20 символов) - это может быть товар
				if len(afterPattern) > 20 {
					// Проверяем количество слов (товары обычно имеют больше слов)
					words := strings.Split(afterPattern, "-")
					if len(words) > 3 {
						return true // Вероятно товар
					}
				}
			}
			return false
		}
	}

	// Исключаем служебные страницы
	excludePatterns := []string{
		"/pretraga", "/search", "/kontakt", "/contact",
		"/o-nama", "/about", "/stranica/", "/page/",
		"/login", "/cart", "/korpa", "/checkout", "/kosarica",
		"/account", "/nalog", "/profile", "/profil",
		"/register", "/registracija", "/signup",
		"/help", "/pomoc", "/support", "/podrska",
		"/faq", "/cesto-postavljana-pitanja",
		"/servis", "/service", "/reklamacije", "/warranty",
		"/delivery", "/dostava", "/shipping", "/isporuka",
		"/payment", "/placanje", "/naplata",
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(urlLower, pattern) {
			return false
		}
	}

	// Проверка на явные паттерны товаров (даем приоритет)
	productPatterns := []string{
		"/proizvod/", "/product/", "/p/", "/artikal/",
		"/item/", "/goods/", "/roba/",
	}

	for _, pattern := range productPatterns {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}

	// Если URL достаточно длинный (вероятно товар с полным названием)
	// и не содержит явных паттернов категорий - считаем товаром
	if len(url) > 50 {
		return true
	}

	return true
}

// filterEmpty удаляет пустые строки из слайса
func filterEmpty(s []string) []string {
	var result []string
	for _, v := range s {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}
