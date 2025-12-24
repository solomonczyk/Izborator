package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/logger"
	"github.com/solomonczyk/izborator/internal/products"
)

// TestProductsAdapter_GetProduct тестирует получение товара по ID
func TestProductsAdapter_GetProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	log := logger.New("error")
	adapter := NewProductsAdapter(pg, nil, log)

	// Тест 1: Получение несуществующего товара
	t.Run("NonExistentProduct", func(t *testing.T) {
		testID := uuid.New().String()
		product, err := adapter.GetProduct(testID)
		if err == nil {
			t.Errorf("Expected error for non-existent product, got nil")
		}
		if product != nil {
			t.Errorf("Expected nil product, got %v", product)
		}
		if err != products.ErrProductNotFound {
			t.Errorf("Expected ErrProductNotFound, got %v", err)
		}
	})

	// Тест 2: Создание и получение товара
	t.Run("CreateAndGet", func(t *testing.T) {
		product := &products.Product{
			ID:          uuid.New().String(),
			Name:        "Test Product",
			Description: "Test Description",
			Brand:       "Test Brand",
			Category:    "Test Category",
			ImageURL:    "https://example.com/image.jpg",
			Specs: map[string]string{
				"color": "black",
				"size":  "large",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Сохраняем товар
		if err := adapter.SaveProduct(product); err != nil {
			t.Fatalf("Failed to save product: %v", err)
		}

		// Получаем товар
		retrieved, err := adapter.GetProduct(product.ID)
		if err != nil {
			t.Fatalf("Failed to get product: %v", err)
		}

		// Проверяем данные
		if retrieved.ID != product.ID {
			t.Errorf("Expected ID %s, got %s", product.ID, retrieved.ID)
		}
		if retrieved.Name != product.Name {
			t.Errorf("Expected Name %s, got %s", product.Name, retrieved.Name)
		}
		if retrieved.Brand != product.Brand {
			t.Errorf("Expected Brand %s, got %s", product.Brand, retrieved.Brand)
		}
		if len(retrieved.Specs) != 2 {
			t.Errorf("Expected 2 specs, got %d", len(retrieved.Specs))
		}
		if retrieved.Specs["color"] != "black" {
			t.Errorf("Expected color 'black', got '%s'", retrieved.Specs["color"])
		}
	})

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_prices"})
}

// TestProductsAdapter_SearchProducts тестирует поиск товаров
func TestProductsAdapter_SearchProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	log := logger.New("error")
	adapter := NewProductsAdapter(pg, nil, log)

	// Создаём тестовые товары
	products := []*products.Product{
		{
			ID:          uuid.New().String(),
			Name:        "iPhone 15 Pro",
			Description: "Apple smartphone",
			Brand:       "Apple",
			Category:    "Smartphones",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Samsung Galaxy S24",
			Description: "Samsung smartphone",
			Brand:       "Samsung",
			Category:    "Smartphones",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "MacBook Pro",
			Description: "Apple laptop",
			Brand:       "Apple",
			Category:    "Laptops",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, p := range products {
		if err := adapter.SaveProduct(p); err != nil {
			t.Fatalf("Failed to save product: %v", err)
		}
	}

	// Тест 1: Поиск по названию
	t.Run("SearchByName", func(t *testing.T) {
		searchResults, total, err := adapter.SearchProducts("iPhone", 10, 0)
		if err != nil {
			t.Fatalf("SearchProducts failed: %v", err)
		}
		if total < 1 {
			t.Errorf("Expected at least 1 result, got %d", total)
		}
		if len(searchResults) < 1 {
			t.Errorf("Expected at least 1 item, got %d", len(searchResults))
		}
		found := false
		for _, p := range searchResults {
			if p.Name == "iPhone 15 Pro" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find 'iPhone 15 Pro' in results")
		}
	})

	// Тест 2: Поиск по бренду
	t.Run("SearchByBrand", func(t *testing.T) {
		searchResults, total, err := adapter.SearchProducts("Apple", 10, 0)
		if err != nil {
			t.Fatalf("SearchProducts failed: %v", err)
		}
		if total < 2 {
			t.Errorf("Expected at least 2 results for 'Apple', got %d", total)
		}
		if len(searchResults) < 2 {
			t.Errorf("Expected at least 2 items for 'Apple', got %d", len(searchResults))
		}
	})

	// Тест 3: Пагинация
	t.Run("Pagination", func(t *testing.T) {
		searchResults, total, err := adapter.SearchProducts("", 2, 0)
		if err != nil {
			t.Fatalf("SearchProducts failed: %v", err)
		}
		if len(searchResults) > 2 {
			t.Errorf("Expected max 2 items, got %d", len(searchResults))
		}
		if total < 3 {
			t.Errorf("Expected total >= 3, got %d", total)
		}
	})

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_prices"})
}

// TestProductsAdapter_GetProductPrices тестирует получение цен товара
func TestProductsAdapter_GetProductPrices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	log := logger.New("error")
	adapter := NewProductsAdapter(pg, nil, log)

	// Создаём товар
	product := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Test Product",
		Brand:       "Test Brand",
		Category:    "Test Category",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := adapter.SaveProduct(product); err != nil {
		t.Fatalf("Failed to save product: %v", err)
	}

	// Создаём цены
	prices := []*products.ProductPrice{
		{
			ProductID: product.ID,
			ShopID:    uuid.New().String(),
			ShopName:  "Shop 1",
			Price:     100.0,
			Currency:  "RSD",
			URL:       "https://shop1.com/product",
			InStock:   true,
			UpdatedAt: time.Now(),
		},
		{
			ProductID: product.ID,
			ShopID:    uuid.New().String(),
			ShopName:  "Shop 2",
			Price:     120.0,
			Currency:  "RSD",
			URL:       "https://shop2.com/product",
			InStock:   true,
			UpdatedAt: time.Now(),
		},
	}

	for _, price := range prices {
		if err := adapter.SaveProductPrice(price); err != nil {
			t.Fatalf("Failed to save price: %v", err)
		}
	}

	// Тест: Получение цен
	retrievedPrices, err := adapter.GetProductPrices(product.ID)
	if err != nil {
		t.Fatalf("GetProductPrices failed: %v", err)
	}

	if len(retrievedPrices) != 2 {
		t.Errorf("Expected 2 prices, got %d", len(retrievedPrices))
	}

	// Проверяем сортировку (по возрастанию цены)
	if retrievedPrices[0].Price != 100.0 {
		t.Errorf("Expected first price to be 100.0, got %f", retrievedPrices[0].Price)
	}
	if retrievedPrices[1].Price != 120.0 {
		t.Errorf("Expected second price to be 120.0, got %f", retrievedPrices[1].Price)
	}

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_prices"})
}

// TestProductsAdapter_Browse тестирует каталог товаров
func TestProductsAdapter_Browse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	log := logger.New("error")
	adapter := NewProductsAdapter(pg, nil, log)

	// Создаём товары с ценами
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

	if err := adapter.SaveProduct(product1); err != nil {
		t.Fatalf("Failed to save product1: %v", err)
	}
	if err := adapter.SaveProduct(product2); err != nil {
		t.Fatalf("Failed to save product2: %v", err)
	}

	// Добавляем цены
	price1 := &products.ProductPrice{
		ProductID: product1.ID,
		ShopID:    uuid.New().String(),
		ShopName:  "Shop 1",
		Price:     100.0,
		Currency:  "RSD",
		InStock:   true,
		UpdatedAt: time.Now(),
	}

	if err := adapter.SaveProductPrice(price1); err != nil {
		t.Fatalf("Failed to save price1: %v", err)
	}

	// Тест: Browse без фильтров
	ctx := context.Background()
	params := products.BrowseParams{
		Page:    1,
		PerPage: 10,
		Sort:    "name_asc",
	}

	result, err := adapter.Browse(ctx, params)
	if err != nil {
		t.Fatalf("Browse failed: %v", err)
	}

	if result.Total < 2 {
		t.Errorf("Expected at least 2 products, got %d", result.Total)
	}

	// Тест: Browse с фильтром по категории
	params.Category = "Category 1"
	result2, err := adapter.Browse(ctx, params)
	if err != nil {
		t.Fatalf("Browse with category filter failed: %v", err)
	}

	if result2.Total < 2 {
		t.Errorf("Expected at least 2 products in category, got %d", result2.Total)
	}

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_prices"})
}

