package handlers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/products"
	"github.com/solomonczyk/izborator/internal/storage"
)

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// TestProductsHandler_GetByID_Integration тестирует GET /api/v1/products/{id}
func TestProductsHandler_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, pg, cleanup := setupTestServer(t)
	defer cleanup()
	defer server.Close()

	// Создаём тестовый товар с ценами
	productsStorage := storage.NewProductsAdapter(pg, nil, nil)

	product := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Test Product",
		Description: "Test Description",
		Brand:       "Test Brand",
		Category:    "Test Category",
		ImageURL:    "https://example.com/image.jpg",
		Specs: map[string]string{
			"color": "black",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := productsStorage.SaveProduct(product); err != nil {
		t.Fatalf("Failed to save product: %v", err)
	}

	price := &products.ProductPrice{
		ProductID: product.ID,
		ShopID:    uuid.New().String(),
		ShopName:  "Test Shop",
		Price:     100.0,
		Currency:  "RSD",
		URL:       "https://shop.com/product",
		InStock:   true,
		UpdatedAt: time.Now(),
	}

	storage.EnsureTestShop(t, pg, price.ShopID, price.ShopName)
	if err := productsStorage.SaveProductPrice(price); err != nil {
		t.Fatalf("Failed to save price: %v", err)
	}

	// Тест 1: Получение существующего товара
	t.Run("ExistingProduct", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/"+product.ID)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result ProductResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.ID != product.ID {
			t.Errorf("Expected ID %s, got %s", product.ID, result.ID)
		}
		if result.Name != product.Name {
			t.Errorf("Expected Name %s, got %s", product.Name, result.Name)
		}
		if len(result.Prices) != 1 {
			t.Errorf("Expected 1 price, got %d", len(result.Prices))
		}
	})

	// Тест 2: Получение несуществующего товара
	t.Run("NonExistentProduct", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		resp := makeRequest(t, server, "GET", "/api/v1/products/"+nonExistentID)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}

		errResp := decodeErrorResponse(t, resp)
		if errResp.Error.Code != "NOT_FOUND" {
			t.Errorf("Expected error code NOT_FOUND, got %s", errResp.Error.Code)
		}
		if errResp.Error.Message == "" {
			t.Error("Expected non-empty error message")
		}
	})

	// Тест 3: Невалидный UUID
	t.Run("InvalidUUID", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/invalid-uuid")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		errResp := decodeErrorResponse(t, resp)
		if errResp.Error.Code != "VALIDATION_FAILED" {
			t.Errorf("Expected error code VALIDATION_FAILED, got %s", errResp.Error.Code)
		}
		if errResp.Error.Message == "" {
			t.Error("Expected non-empty error message")
		}
	})
}

// TestProductsHandler_Search_Integration тестирует GET /api/v1/products/search
func TestProductsHandler_Search_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, pg, cleanup := setupTestServer(t)
	defer cleanup()
	defer server.Close()

	// Создаём тестовые товары
	productsStorage := storage.NewProductsAdapter(pg, nil, nil)

	products := []*products.Product{
		{
			ID:          uuid.New().String(),
			Name:        "iPhone 15 Pro",
			Brand:       "Apple",
			Category:    "Smartphones",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Samsung Galaxy S24",
			Brand:       "Samsung",
			Category:    "Smartphones",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, p := range products {
		if err := productsStorage.SaveProduct(p); err != nil {
			t.Fatalf("Failed to save product: %v", err)
		}
	}

	// Тест 1: Поиск по названию
	t.Run("SearchByName", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/search?q=iPhone")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Search возвращает []*products.Product, но JSON декодирует в []products.Product
		var result []struct {
			ID          string            `json:"id"`
			Name        string            `json:"name"`
			Description string            `json:"description"`
			Brand       string            `json:"brand"`
			Category    string            `json:"category"`
			ImageURL    string            `json:"image_url"`
			Specs       map[string]string `json:"specs"`
			CreatedAt   string            `json:"created_at"`
			UpdatedAt   string            `json:"updated_at"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(result) == 0 {
			t.Error("Expected at least 1 result")
		}

		found := false
		for i := range result {
			if result[i].Name == "iPhone 15 Pro" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find 'iPhone 15 Pro' in results")
		}
	})

	// Тест 2: Пустой запрос
	t.Run("EmptyQuery", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/search?q=")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		errResp := decodeErrorResponse(t, resp)
		if errResp.Error.Code != "VALIDATION_FAILED" {
			t.Errorf("Expected error code VALIDATION_FAILED, got %s", errResp.Error.Code)
		}
		if errResp.Error.Message == "" {
			t.Error("Expected non-empty error message")
		}
	})
}

// TestProductsHandler_Browse_Integration тестирует GET /api/v1/products/browse
func TestProductsHandler_Browse_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, pg, cleanup := setupTestServer(t)
	defer cleanup()
	defer server.Close()

	// Создаём тестовые товары с ценами
	productsStorage := storage.NewProductsAdapter(pg, nil, nil)

	product1 := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Product 1",
		Brand:       "Brand A",
		Category:    "Category 1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	product2 := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Product 2",
		Brand:       "Brand B",
		Category:    "Category 1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, p := range []*products.Product{product1, product2} {
		if err := productsStorage.SaveProduct(p); err != nil {
			t.Fatalf("Failed to save product: %v", err)
		}

		price := &products.ProductPrice{
			ProductID: p.ID,
			ShopID:    uuid.New().String(),
			ShopName:  "Test Shop",
			Price:     100.0,
			Currency:  "RSD",
			InStock:   true,
			UpdatedAt: time.Now(),
		}

		storage.EnsureTestShop(t, pg, price.ShopID, price.ShopName)
		if err := productsStorage.SaveProductPrice(price); err != nil {
			t.Fatalf("Failed to save price: %v", err)
		}
	}

	// Тест 1: Browse без фильтров
	t.Run("BrowseWithoutFilters", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/browse?page=1&per_page=10")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result products.BrowseResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.Total < 2 {
			t.Errorf("Expected at least 2 products, got %d", result.Total)
		}
		if len(result.Items) < 2 {
			t.Errorf("Expected at least 2 items, got %d", len(result.Items))
		}
	})

	// Тест 2: Browse с фильтром по цене
	t.Run("BrowseWithPriceFilter", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/browse?min_price=50&max_price=150&page=1&per_page=10")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result products.BrowseResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Проверяем, что все товары в диапазоне цен
		for _, item := range result.Items {
			if item.MinPrice < 50 || item.MaxPrice > 150 {
				t.Errorf("Product %s price out of range: min=%f, max=%f", item.ID, item.MinPrice, item.MaxPrice)
			}
		}
	})

	// Тест 3: Невалидная пагинация
	t.Run("InvalidPagination", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/browse?page=0&per_page=10")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		errResp := decodeErrorResponse(t, resp)
		if errResp.Error.Code != "VALIDATION_FAILED" {
			t.Errorf("Expected error code VALIDATION_FAILED, got %s", errResp.Error.Code)
		}
		if errResp.Error.Message == "" {
			t.Error("Expected non-empty error message")
		}
	})
}

func decodeErrorResponse(t *testing.T, resp *http.Response) errorResponse {
	t.Helper()

	var payload errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}
	return payload
}

// TestProductsHandler_GetPrices_Integration тестирует GET /api/v1/products/{id}/prices
func TestProductsHandler_GetPrices_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, pg, cleanup := setupTestServer(t)
	defer cleanup()
	defer server.Close()

	// Создаём товар с ценами
	productsStorage := storage.NewProductsAdapter(pg, nil, nil)

	product := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Test Product",
		Brand:       "Test Brand",
		Category:    "Test Category",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := productsStorage.SaveProduct(product); err != nil {
		t.Fatalf("Failed to save product: %v", err)
	}

	prices := []*products.ProductPrice{
		{
			ProductID: product.ID,
			ShopID:    uuid.New().String(),
			ShopName:  "Shop 1",
			Price:     100.0,
			Currency:  "RSD",
			InStock:   true,
			UpdatedAt: time.Now(),
		},
		{
			ProductID: product.ID,
			ShopID:    uuid.New().String(),
			ShopName:  "Shop 2",
			Price:     120.0,
			Currency:  "RSD",
			InStock:   true,
			UpdatedAt: time.Now(),
		},
	}

	for _, price := range prices {
		storage.EnsureTestShop(t, pg, price.ShopID, price.ShopName)
		if err := productsStorage.SaveProductPrice(price); err != nil {
			t.Fatalf("Failed to save price: %v", err)
		}
	}

	// Тест: Получение цен
	resp := makeRequest(t, server, "GET", "/api/v1/products/"+product.ID+"/prices")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	pricesArray, ok := result["prices"].([]interface{})
	if !ok {
		t.Fatal("Expected 'prices' array in response")
	}

	if len(pricesArray) != 2 {
		t.Errorf("Expected 2 prices, got %d", len(pricesArray))
	}
}

