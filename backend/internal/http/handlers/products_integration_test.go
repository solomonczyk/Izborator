package handlers

import (
	"context"
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

// TestProductsHandler_Browse_WithCategoriesAndCities_Integration - E2E тест фильтрации по категориям и городам
func TestProductsHandler_Browse_WithCategoriesAndCities_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server, pg, cleanup := setupTestServer(t)
	defer cleanup()
	defer server.Close()

	ctx := context.Background()

	// 1. Создаём тестовые категории (родительскую и дочернюю)
	parentCategoryID := uuid.New().String()
	childCategoryID := uuid.New().String()
	parentSlug := "test-electronics"
	childSlug := "test-phones"

	_, err := pg.DB().Exec(ctx, `
		INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, is_active, sort_order)
		VALUES 
			($1, NULL, $2, 'TEST_ELECTRONICS', 'Test Electronics', 'test electronics', 1, true, 10),
			($3, $1, $4, 'TEST_PHONES', 'Test Phones', 'test phones', 2, true, 20)
		ON CONFLICT (slug) DO UPDATE SET
			parent_id = EXCLUDED.parent_id,
			code = EXCLUDED.code,
			name_sr = EXCLUDED.name_sr,
			name_sr_lc = EXCLUDED.name_sr_lc,
			level = EXCLUDED.level,
			is_active = EXCLUDED.is_active
	`, parentCategoryID, parentSlug, childCategoryID, childSlug)
	if err != nil {
		t.Fatalf("Failed to create test categories: %v", err)
	}

	// 2. Создаём тестовые города
	city1ID := uuid.New().String()
	city2ID := uuid.New().String()
	city1Slug := "test-beograd"
	city2Slug := "test-novi-sad"

	_, err = pg.DB().Exec(ctx, `
		INSERT INTO cities (id, slug, name_sr, sort_order, is_active)
		VALUES 
			($1, $2, 'Test Beograd', 10, true),
			($3, $4, 'Test Novi Sad', 20, true)
		ON CONFLICT (slug) DO UPDATE SET
			name_sr = EXCLUDED.name_sr,
			sort_order = EXCLUDED.sort_order,
			is_active = EXCLUDED.is_active
	`, city1ID, city1Slug, city2ID, city2Slug)
	if err != nil {
		t.Fatalf("Failed to create test cities: %v", err)
	}

	// 3. Создаём тестовые товары с привязкой к категориям
	productsStorage := storage.NewProductsAdapter(pg, nil, nil)

	// Товар 1: в родительской категории, доступен в городе 1
	product1 := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Product in Parent Category",
		Brand:       "Brand A",
		Category:    "Test Electronics",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	// Сохраняем товар
	if err := productsStorage.SaveProduct(product1); err != nil {
		t.Fatalf("Failed to save product1: %v", err)
	}
	// Привязываем к родительской категории
	_, err = pg.DB().Exec(ctx, `UPDATE products SET category_id = $1 WHERE id = $2`, parentCategoryID, product1.ID)
	if err != nil {
		t.Fatalf("Failed to link product1 to category: %v", err)
	}

	// Товар 2: в дочерней категории, доступен в городе 1
	product2 := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Product in Child Category",
		Brand:       "Brand B",
		Category:    "Test Phones",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := productsStorage.SaveProduct(product2); err != nil {
		t.Fatalf("Failed to save product2: %v", err)
	}
	_, err = pg.DB().Exec(ctx, `UPDATE products SET category_id = $1 WHERE id = $2`, childCategoryID, product2.ID)
	if err != nil {
		t.Fatalf("Failed to link product2 to category: %v", err)
	}

	// Товар 3: в дочерней категории, доступен в городе 2
	product3 := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Product in Child Category City 2",
		Brand:       "Brand C",
		Category:    "Test Phones",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := productsStorage.SaveProduct(product3); err != nil {
		t.Fatalf("Failed to save product3: %v", err)
	}
	_, err = pg.DB().Exec(ctx, `UPDATE products SET category_id = $1 WHERE id = $2`, childCategoryID, product3.ID)
	if err != nil {
		t.Fatalf("Failed to link product3 to category: %v", err)
	}

	// 4. Создаём цены с привязкой к городам
	shop1ID := uuid.New().String()
	shop2ID := uuid.New().String()
	storage.EnsureTestShop(t, pg, shop1ID, "Shop 1")
	storage.EnsureTestShop(t, pg, shop2ID, "Shop 2")

	// Цены для product1 (город 1)
	price1 := &products.ProductPrice{
		ProductID: product1.ID,
		ShopID:    shop1ID,
		ShopName:  "Shop 1",
		Price:     100.0,
		Currency:  "RSD",
		InStock:   true,
		UpdatedAt: time.Now(),
	}
	if err := productsStorage.SaveProductPrice(price1); err != nil {
		t.Fatalf("Failed to save price1: %v", err)
	}
	_, err = pg.DB().Exec(ctx, `UPDATE product_prices SET city_id = $1 WHERE product_id = $2 AND shop_id = $3`, city1ID, product1.ID, shop1ID)
	if err != nil {
		t.Fatalf("Failed to link price1 to city: %v", err)
	}

	// Цены для product2 (город 1)
	price2 := &products.ProductPrice{
		ProductID: product2.ID,
		ShopID:    shop1ID,
		ShopName:  "Shop 1",
		Price:     200.0,
		Currency:  "RSD",
		InStock:   true,
		UpdatedAt: time.Now(),
	}
	if err := productsStorage.SaveProductPrice(price2); err != nil {
		t.Fatalf("Failed to save price2: %v", err)
	}
	_, err = pg.DB().Exec(ctx, `UPDATE product_prices SET city_id = $1 WHERE product_id = $2 AND shop_id = $3`, city1ID, product2.ID, shop1ID)
	if err != nil {
		t.Fatalf("Failed to link price2 to city: %v", err)
	}

	// Цены для product3 (город 2)
	price3 := &products.ProductPrice{
		ProductID: product3.ID,
		ShopID:    shop2ID,
		ShopName:  "Shop 2",
		Price:     300.0,
		Currency:  "RSD",
		InStock:   true,
		UpdatedAt: time.Now(),
	}
	if err := productsStorage.SaveProductPrice(price3); err != nil {
		t.Fatalf("Failed to save price3: %v", err)
	}
	_, err = pg.DB().Exec(ctx, `UPDATE product_prices SET city_id = $1 WHERE product_id = $2 AND shop_id = $3`, city2ID, product3.ID, shop2ID)
	if err != nil {
		t.Fatalf("Failed to link price3 to city: %v", err)
	}

	// Тест 1: Фильтрация по родительской категории (должна включать дочерние)
	t.Run("FilterByParentCategory", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/browse?category="+parentSlug+"&page=1&per_page=10")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result products.BrowseResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Должны найтись product1 (родительская) и product2, product3 (дочерние)
		if result.Total < 3 {
			t.Errorf("Expected at least 3 products (parent + children), got %d", result.Total)
		}

		// Проверяем, что все товары из нужной категории
		foundProduct1 := false
		foundProduct2 := false
		foundProduct3 := false
		for _, item := range result.Items {
			if item.ID == product1.ID {
				foundProduct1 = true
			}
			if item.ID == product2.ID {
				foundProduct2 = true
			}
			if item.ID == product3.ID {
				foundProduct3 = true
			}
		}

		if !foundProduct1 {
			t.Error("Expected to find product1 in parent category results")
		}
		if !foundProduct2 {
			t.Error("Expected to find product2 in parent category results (child category)")
		}
		if !foundProduct3 {
			t.Error("Expected to find product3 in parent category results (child category)")
		}
	})

	// Тест 2: Фильтрация по дочерней категории
	t.Run("FilterByChildCategory", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/browse?category="+childSlug+"&page=1&per_page=10")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result products.BrowseResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Должны найтись только product2 и product3 (дочерняя категория)
		if result.Total < 2 {
			t.Errorf("Expected at least 2 products (children only), got %d", result.Total)
		}

		foundProduct2 := false
		foundProduct3 := false
		for _, item := range result.Items {
			if item.ID == product2.ID {
				foundProduct2 = true
			}
			if item.ID == product3.ID {
				foundProduct3 = true
			}
			// product1 не должен быть в результатах
			if item.ID == product1.ID {
				t.Error("Product1 should not be in child category results")
			}
		}

		if !foundProduct2 {
			t.Error("Expected to find product2 in child category results")
		}
		if !foundProduct3 {
			t.Error("Expected to find product3 in child category results")
		}
	})

	// Тест 3: Фильтрация по городу
	t.Run("FilterByCity", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/browse?city="+city1Slug+"&page=1&per_page=10")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result products.BrowseResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Должны найтись product1 и product2 (город 1)
		if result.Total < 2 {
			t.Errorf("Expected at least 2 products in city1, got %d", result.Total)
		}

		foundProduct1 := false
		foundProduct2 := false
		for _, item := range result.Items {
			if item.ID == product1.ID {
				foundProduct1 = true
			}
			if item.ID == product2.ID {
				foundProduct2 = true
			}
			// product3 не должен быть в результатах (он в городе 2)
			if item.ID == product3.ID {
				t.Error("Product3 should not be in city1 results")
			}
		}

		if !foundProduct1 {
			t.Error("Expected to find product1 in city1 results")
		}
		if !foundProduct2 {
			t.Error("Expected to find product2 in city1 results")
		}
	})

	// Тест 4: Комбинированная фильтрация (категория + город)
	t.Run("FilterByCategoryAndCity", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/browse?category="+childSlug+"&city="+city1Slug+"&page=1&per_page=10")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result products.BrowseResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Должен найтись только product2 (дочерняя категория + город 1)
		if result.Total < 1 {
			t.Errorf("Expected at least 1 product, got %d", result.Total)
		}

		foundProduct2 := false
		for _, item := range result.Items {
			if item.ID == product2.ID {
				foundProduct2 = true
			}
			// product1 не должен быть (не в дочерней категории)
			if item.ID == product1.ID {
				t.Error("Product1 should not be in child category results")
			}
			// product3 не должен быть (не в городе 1)
			if item.ID == product3.ID {
				t.Error("Product3 should not be in city1 results")
			}
		}

		if !foundProduct2 {
			t.Error("Expected to find product2 in child category + city1 results")
		}
	})

	// Тест 5: Комбинированная фильтрация (категория + город + цена)
	t.Run("FilterByCategoryCityAndPrice", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/products/browse?category="+childSlug+"&city="+city1Slug+"&min_price=150&max_price=250&page=1&per_page=10")
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result products.BrowseResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Должен найтись только product2 (цена 200 в диапазоне 150-250)
		if result.Total < 1 {
			t.Errorf("Expected at least 1 product, got %d", result.Total)
		}

		// Проверяем, что все товары в диапазоне цен
		for _, item := range result.Items {
			if item.MinPrice < 150 || item.MaxPrice > 250 {
				t.Errorf("Product %s price out of range: min=%f, max=%f", item.ID, item.MinPrice, item.MaxPrice)
			}
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

