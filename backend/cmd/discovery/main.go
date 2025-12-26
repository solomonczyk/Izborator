package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/solomonczyk/izborator/internal/app"
	"github.com/solomonczyk/izborator/internal/classifier"
	"github.com/solomonczyk/izborator/internal/config"
)

// Dorking Queries — запросы для поиска магазинов и услуг в Сербии
var queries = []string{
	// E-commerce магазины (существующие запросы)
	"site:.rs \"dodaj u korpu\"",
	"site:.rs \"kupi odmah\"",
	"site:.rs \"cena rsd\"",
	"site:.rs inurl:proizvod",
	"site:.rs inurl:kategorija",
	"site:.rs \"besplatna dostava\" cena",
	"site:.rs \"online prodavnica\"",
	"site:.rs \"internet prodavnica\"",
	"site:.rs \"e-shop\"",
	"site:.rs \"webshop\"",
	
	// Услуги - прайс-листы и цены
	"site:.rs \"cenovnik usluga\"",
	"site:.rs \"cenovnik\" cena",
	"site:.rs \"cena usluge\"",
	"site:.rs \"cena rada\"",
	"site:.rs \"zakazivanje termina\"",
	"site:.rs \"rezervacija\" cena",
	
	// Медицинские услуги
	"site:.rs \"zubarska ordinacija\" cene",
	"site:.rs \"dermatolog\" cena",
	"site:.rs \"fizioterapija\" cena",
	"site:.rs \"masaza\" cena",
	
	// Красота и уход
	"site:.rs \"frizerski salon\" cena",
	"site:.rs \"manikir pedikir\" cena",
	"site:.rs \"kozmeticki salon\" cena",
	
	// Ремонт и обслуживание
	"site:.rs \"servis\" cena",
	"site:.rs \"popravka\" cena",
	"site:.rs \"montaza\" cena",
	
	// Образование и курсы
	"site:.rs \"kurs\" cena",
	"site:.rs \"obuka\" cena",
	"site:.rs \"skola\" cena",
	
	// Юридические и консультационные услуги
	"site:.rs \"advokat\" cena",
	"site:.rs \"notar\" cena",
	"site:.rs \"konsultacije\" cena",
	
	// Транспорт и доставка
	"site:.rs \"prevoz\" cena",
	"site:.rs \"dostava\" cena",
	"site:.rs \"kurirska sluzba\" cena",
	
	// Общие паттерны для услуг
	"site:.rs inurl:cenovnik",
	"site:.rs inurl:cene",
	"site:.rs inurl:usluge",
	"site:.rs \"tabela cena\"",
	"site:.rs \"cena po satu\"",
	"site:.rs \"cena po terminu\"",
}

// GoogleResult структура ответа Google Custom Search API
type GoogleResult struct {
	Items []struct {
		Link  string `json:"link"`
		Title string `json:"title"`
	} `json:"items"`
	SearchInformation struct {
		TotalResults string `json:"totalResults"`
	} `json:"searchInformation"`
}

func main() {
	// Загрузка .env файла
	_ = godotenv.Load()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Инициализация приложения
	application, err := app.NewWorkerApp(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to init app: %v", err))
	}
	defer application.Close()

	log := application.Logger()
	storage := application.GetClassifierStorage()

	// Получаем ключи из конфигурации
	apiKey := cfg.Google.APIKey
	cx := cfg.Google.CX

	// Флаги (опциональные, для переопределения)
	apiKeyFlag := flag.String("key", "", "Google API Key (optional, overrides env)")
	cxFlag := flag.String("cx", "", "Custom Search Engine ID (optional, overrides env)")
	maxResults := flag.Int("max-results", 100, "Maximum results per query (default: 100, max: 100)")
	delay := flag.Duration("delay", 1*time.Second, "Delay between requests (default: 1s)")
	flag.Parse()

	// Если переданы флаги, используем их (для обратной совместимости)
	if *apiKeyFlag != "" {
		apiKey = *apiKeyFlag
	}
	if *cxFlag != "" {
		cx = *cxFlag
	}

	if apiKey == "" || cx == "" {
		log.Fatal("API Key and CX are required. Set GOOGLE_API_KEY and GOOGLE_CX in .env or use -key and -cx flags", nil)
	}

	if *maxResults > 100 {
		*maxResults = 100 // Google ограничивает до 100 результатов на запрос
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	totalDiscovered := 0
	totalSkipped := 0

	log.Info("🔍 Starting Discovery Worker", map[string]interface{}{
		"queries":     len(queries),
		"max_results": *maxResults,
		"delay":       delay.String(),
	})

	for i, query := range queries {
		log.Info("🔎 Processing query", map[string]interface{}{
			"query":  query,
			"number": i + 1,
			"total":  len(queries),
		})

		// Google разрешает только 100 результатов на запрос (start=1, 11, 21, ... 91)
		// Максимум 10 страниц по 10 результатов
		pages := (*maxResults + 9) / 10 // Округление вверх
		if pages > 10 {
			pages = 10
		}

		for page := 0; page < pages; page++ {
			start := page*10 + 1

			googleURL := fmt.Sprintf(
				"https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&start=%d&gl=rs&num=10",
				apiKey, cx, url.QueryEscape(query), start,
			)

			resp, err := client.Get(googleURL)
			if err != nil {
				log.Error("Failed to request Google", map[string]interface{}{
					"error": err.Error(),
					"query": query,
					"page":  page + 1,
				})
				continue
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				log.Error("Google API returned error", map[string]interface{}{
					"status": resp.StatusCode,
					"query":  query,
					"page":   page + 1,
				})
				continue
			}

			var result GoogleResult
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				log.Error("Failed to decode Google response", map[string]interface{}{
					"error": err.Error(),
					"query": query,
				})
				continue
			}
			resp.Body.Close()

			if len(result.Items) == 0 {
				log.Info("No more results for query", map[string]interface{}{"query": query})
				break
			}

			for _, item := range result.Items {
				domain := extractDomain(item.Link)
				if domain == "" {
					continue
				}

				// Нормализация домена (убираем www.)
				domain = normalizeDomain(domain)

				// Проверяем, не существует ли уже этот домен
				existing, err := storage.GetPotentialShopByDomain(domain)
				if err == nil && existing != nil {
					totalSkipped++
					continue
				}

				// Создаем новый кандидат
				shop := &classifier.PotentialShop{
					ID:             uuid.New().String(),
					Domain:         domain,
					Source:         "google_search",
					Status:         "new",
					ConfidenceScore: 0.0,
					DiscoveredAt:   time.Now().Format(time.RFC3339),
					Metadata: map[string]interface{}{
						"title":      item.Title,
						"url":        item.Link,
						"query":      query,
						"page":       page + 1,
						"discovered": time.Now().Format(time.RFC3339),
					},
				}

				if err := storage.SavePotentialShop(shop); err != nil {
					log.Error("Failed to save potential shop", map[string]interface{}{
						"error":  err.Error(),
						"domain": domain,
					})
					continue
				}

				totalDiscovered++
				log.Info("🆕 Discovered candidate", map[string]interface{}{
					"domain": domain,
					"title":  item.Title,
				})
			}

			// Важно! Не спамим Google, иначе забанят ключ
			if page < pages-1 {
				time.Sleep(*delay)
			}
		}

		// Задержка между запросами
		if i < len(queries)-1 {
			time.Sleep(*delay * 2)
		}
	}

	log.Info("✅ Discovery completed", map[string]interface{}{
		"discovered": totalDiscovered,
		"skipped":    totalSkipped,
		"total":      totalDiscovered + totalSkipped,
	})
}

// extractDomain извлекает домен из URL
func extractDomain(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	return parsed.Host
}

// normalizeDomain нормализует домен (убирает www., приводит к нижнему регистру)
func normalizeDomain(domain string) string {
	domain = strings.ToLower(domain)
	domain = strings.TrimSpace(domain)
	
	// Убираем www.
		domain = strings.TrimPrefix(domain, "www.")
	
	// Убираем порт, если есть
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}
	
	return domain
}

