package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/config"
	"github.com/solomonczyk/izborator/internal/logger"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run cmd/test-selectors/main.go <URL> [selector]")
		fmt.Println("Пример: go run cmd/test-selectors/main.go https://gigatron.rs/mobilni-telefoni-tableti-i-oprema/mobilni-telefoni")
		os.Exit(1)
	}

	url := os.Args[1]
	selector := ".product-box a, .product-item a, .product-card a, .product-title a, article a, .item a"
	if len(os.Args) > 2 {
		selector = os.Args[2]
	}

	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)

	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Printf("🔍 ТЕСТИРОВАНИЕ СЕЛЕКТОРОВ\n")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Selector: %s\n", selector)
	fmt.Println()

	var foundLinks []string
	var foundTexts []string
	var statusCode int
	var errorMsg string

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
	)
	c.SetRequestTimeout(30 * time.Second)
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.OnResponse(func(r *colly.Response) {
		statusCode = r.StatusCode
		log.Info("Response received", map[string]interface{}{
			"status_code": statusCode,
			"content_length": len(r.Body),
		})
	})

	c.OnHTML(selector, func(e *colly.HTMLElement) {
		href := e.Attr("href")
		text := strings.TrimSpace(e.Text)
		
		if href != "" {
			// Преобразуем относительные URL в абсолютные
			if strings.HasPrefix(href, "/") {
				baseURL := strings.Split(url, "/")[0] + "//" + strings.Split(url, "/")[2]
				href = baseURL + href
			} else if !strings.HasPrefix(href, "http") {
				baseURL := strings.TrimSuffix(url, "/")
				href = baseURL + "/" + href
			}
			foundLinks = append(foundLinks, href)
		}
		
		if text != "" {
			foundTexts = append(foundTexts, text)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		errorMsg = err.Error()
		log.Error("Request failed", map[string]interface{}{
			"url": r.Request.URL.String(),
			"error": err.Error(),
			"status_code": r.StatusCode,
		})
	})

	// Собираем все ссылки для анализа структуры
	var allLinks []string
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if href != "" && !strings.HasPrefix(href, "#") && !strings.HasPrefix(href, "javascript:") {
			if strings.HasPrefix(href, "/") {
				baseURL := strings.Split(url, "/")[0] + "//" + strings.Split(url, "/")[2]
				href = baseURL + href
			} else if !strings.HasPrefix(href, "http") {
				baseURL := strings.TrimSuffix(url, "/")
				href = baseURL + "/" + href
			}
			allLinks = append(allLinks, href)
		}
	})

	err = c.Visit(url)
	if err != nil {
		fmt.Printf("❌ Ошибка при загрузке страницы: %v\n", err)
		if errorMsg != "" {
			fmt.Printf("   Детали: %s\n", errorMsg)
		}
		os.Exit(1)
	}

	fmt.Println("📊 РЕЗУЛЬТАТЫ:")
	fmt.Printf("   HTTP Status: %d\n", statusCode)
	if statusCode == 403 {
		fmt.Println("   ⚠️  Forbidden - сайт блокирует запросы")
		fmt.Println("   💡 Возможные решения:")
		fmt.Println("      - Добавить больше заголовков (Accept, Accept-Language)")
		fmt.Println("      - Использовать прокси")
		fmt.Println("      - Увеличить задержку между запросами")
	}
	fmt.Printf("   Найдено ссылок по селектору: %d\n", len(foundLinks))
	fmt.Printf("   Найдено текстов: %d\n", len(foundTexts))
	fmt.Println()

	if len(foundLinks) > 0 {
		fmt.Println("✅ НАЙДЕННЫЕ ССЫЛКИ (первые 10):")
		max := 10
		if len(foundLinks) < max {
			max = len(foundLinks)
		}
		for i, link := range foundLinks[:max] {
			fmt.Printf("   %d. %s\n", i+1, link)
		}
		if len(foundLinks) > max {
			fmt.Printf("   ... и еще %d ссылок\n", len(foundLinks)-max)
		}
		fmt.Println()
	} else {
		fmt.Println("❌ Ссылки не найдены!")
		fmt.Println()
		fmt.Println("💡 ВОЗМОЖНЫЕ ПРИЧИНЫ:")
		fmt.Println("   1. Селектор не соответствует структуре HTML")
		fmt.Println("   2. Страница загружается динамически (JavaScript)")
		fmt.Println("   3. Сайт блокирует запросы (403 Forbidden)")
		fmt.Println("   4. Неправильный URL")
		fmt.Println()
		fmt.Println("🔧 РЕКОМЕНДАЦИИ:")
		fmt.Println("   1. Проверьте HTML страницы в браузере (F12)")
		fmt.Println("   2. Попробуйте другие селекторы:")
		fmt.Println("      - 'a[href*=\"/product/\"]'")
		fmt.Println("      - '.product a'")
		fmt.Println("      - 'article a'")
		fmt.Println("      - '.item a'")
		fmt.Println("   3. Используйте более общие селекторы для начала")
	}

	if len(foundTexts) > 0 {
		fmt.Println("📝 НАЙДЕННЫЕ ТЕКСТЫ (первые 5):")
		max := 5
		if len(foundTexts) < max {
			max = len(foundTexts)
		}
		for i, text := range foundTexts[:max] {
			if len(text) > 60 {
				text = text[:60] + "..."
			}
			fmt.Printf("   %d. %s\n", i+1, text)
		}
		fmt.Println()
	}

	// Показываем примеры всех ссылок для анализа структуры
	if len(foundLinks) == 0 && len(allLinks) > 0 {
		fmt.Println("🔍 АНАЛИЗ: Селектор не нашел ссылки, но на странице есть ссылки:")
		fmt.Printf("   Всего ссылок на странице: %d\n", len(allLinks))
		fmt.Println("   Примеры ссылок (первые 10):")
		max := 10
		if len(allLinks) < max {
			max = len(allLinks)
		}
		for i, link := range allLinks[:max] {
			// Показываем только ссылки, которые могут быть товарами
			if strings.Contains(link, "mobilni-telefoni") || strings.Contains(link, "product") || strings.Contains(link, "proizvod") {
				fmt.Printf("   %d. %s\n", i+1, link)
			}
		}
		fmt.Println()
	}
}
