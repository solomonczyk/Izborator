package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/solomonczyk/izborator/internal/categories"
	"github.com/solomonczyk/izborator/internal/cities"
	"github.com/solomonczyk/izborator/internal/i18n"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/pricehistory"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/storage"
)

// setupTestServer создаёт тестовый HTTP сервер с реальными сервисами
func setupTestServer(t *testing.T) (*httptest.Server, *storage.Postgres, func()) {
	t.Helper()

	// Настраиваем тестовую БД
	pg := storage.SetupTestDB(t)
	deferFunc := func() {
		storage.CleanupTestData(t, pg, []string{"product_prices", "products", "raw_products", "product_matches", "categories", "cities"})
		pg.Close()
	}

	// Создаём сервисы
	log := logger.New("error")
	// Создаём translator для тестов (игнорируем ошибку, если файлы не найдены)
	translator, _ := i18n.NewTranslator("../../i18n/locales")
	if translator == nil {
		// Fallback: создаём пустой translator
		translator = &i18n.Translator{}
	}

	// Storage адаптеры
	productsStorage := storage.NewProductsAdapter(pg, nil, log)
	priceHistoryStorage := storage.NewPriceHistoryAdapter(pg)
	categoriesStorage := storage.NewCategoriesAdapter(pg)
	citiesStorage := storage.NewCitiesAdapter(pg)

	// Сервисы
	productsService := products.New(productsStorage, log)
	priceHistoryService := pricehistory.New(priceHistoryStorage, log)
	categoriesService := categories.New(categoriesStorage, log)
	citiesService := cities.New(citiesStorage, log)

	// Handler
	handler := NewProductsHandler(
		productsService,
		priceHistoryService,
		categoriesService,
		citiesService,
		log,
		translator,
	)

	// Создаём тестовый роутер с chi
	r := chi.NewRouter()
	r.Route("/api/v1/products", func(r chi.Router) {
		r.Get("/search", handler.Search)
		r.Get("/browse", handler.Browse)
		r.Get("/{id}", handler.GetByID)
		r.Get("/{id}/prices", handler.GetPrices)
		r.Get("/{id}/price-history", handler.GetPriceHistory)
	})

	server := httptest.NewServer(r)

	return server, pg, deferFunc
}

// makeRequest выполняет HTTP запрос к тестовому серверу
func makeRequest(t *testing.T, server *httptest.Server, method, path string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, server.URL+path, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

