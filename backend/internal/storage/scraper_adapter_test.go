package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/scraper"
)

// TestScraperAdapter_SaveRawProduct тестирует сохранение сырого товара
func TestScraperAdapter_SaveRawProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := SetupTestDB(t)
	defer pg.Close()

	adapter := NewScraperAdapter(pg)

	// Тест 1: Сохранение нового товара
	t.Run("SaveNewProduct", func(t *testing.T) {
		rawProduct := &scraper.RawProduct{
			ShopID:     uuid.New().String(),
			ShopName:   "Test Shop",
			ExternalID: "ext-123",
			URL:        "https://example.com/product",
			Name:       "Test Product",
			Description: "Test Description",
			Brand:      "Test Brand",
			Category:   "Test Category",
			Price:      100.0,
			Currency:   "RSD",
			ImageURLs:  []string{"https://example.com/image1.jpg"},
			Specs: map[string]string{
				"color": "black",
				"size":  "large",
			},
			InStock:  true,
			ParsedAt: time.Now(),
		}

		if err := adapter.SaveRawProduct(rawProduct); err != nil {
			t.Fatalf("Failed to save raw product: %v", err)
		}
	})

	// Тест 2: Обновление существующего товара (ON CONFLICT)
	t.Run("UpdateExistingProduct", func(t *testing.T) {
		shopID := uuid.New().String()
		externalID := "ext-456"

		// Первое сохранение
		rawProduct1 := &scraper.RawProduct{
			ShopID:     shopID,
			ShopName:   "Test Shop",
			ExternalID: externalID,
			Name:       "Old Name",
			Price:      100.0,
			Currency:   "RSD",
			InStock:    true,
			ParsedAt:   time.Now(),
		}

		if err := adapter.SaveRawProduct(rawProduct1); err != nil {
			t.Fatalf("Failed to save raw product (first time): %v", err)
		}

		// Обновление с новыми данными
		rawProduct2 := &scraper.RawProduct{
			ShopID:     shopID,
			ShopName:   "Test Shop Updated",
			ExternalID: externalID,
			Name:       "New Name",
			Price:      120.0,
			Currency:   "RSD",
			InStock:    false,
			ParsedAt:   time.Now(),
		}

		if err := adapter.SaveRawProduct(rawProduct2); err != nil {
			t.Fatalf("Failed to update raw product: %v", err)
		}

		// Проверяем, что товар обновился
		products, err := adapter.GetUnprocessedRawProducts(10)
		if err != nil {
			t.Fatalf("Failed to get unprocessed products: %v", err)
		}

		found := false
		for _, p := range products {
			if p.ShopID == shopID && p.ExternalID == externalID {
				found = true
				if p.Name != "New Name" {
					t.Errorf("Expected name 'New Name', got '%s'", p.Name)
				}
				if p.Price != 120.0 {
					t.Errorf("Expected price 120.0, got %f", p.Price)
				}
				if p.InStock {
					t.Error("Expected InStock to be false after update")
				}
				break
			}
		}

		if !found {
			t.Error("Failed to find updated product")
		}
	})

	// Очистка
	CleanupTestData(t, pg, []string{"raw_products"})
}

// TestScraperAdapter_GetUnprocessedRawProducts тестирует получение необработанных товаров
func TestScraperAdapter_GetUnprocessedRawProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := SetupTestDB(t)
	defer pg.Close()

	adapter := NewScraperAdapter(pg)

	// Создаём несколько необработанных товаров
	products := []*scraper.RawProduct{
		{
			ShopID:     uuid.New().String(),
			ShopName:   "Shop 1",
			ExternalID: "ext-1",
			Name:       "Product 1",
			Price:      100.0,
			Currency:   "RSD",
			InStock:    true,
			ParsedAt:   time.Now(),
		},
		{
			ShopID:     uuid.New().String(),
			ShopName:   "Shop 2",
			ExternalID: "ext-2",
			Name:       "Product 2",
			Price:      200.0,
			Currency:   "RSD",
			InStock:    true,
			ParsedAt:   time.Now(),
		},
		{
			ShopID:     uuid.New().String(),
			ShopName:   "Shop 3",
			ExternalID: "ext-3",
			Name:       "Product 3",
			Price:      300.0,
			Currency:   "RSD",
			InStock:    true,
			ParsedAt:   time.Now(),
		},
	}

	for _, p := range products {
		if err := adapter.SaveRawProduct(p); err != nil {
			t.Fatalf("Failed to save raw product: %v", err)
		}
	}

	// Тест: Получение необработанных товаров
	retrieved, err := adapter.GetUnprocessedRawProducts(10)
	if err != nil {
		t.Fatalf("GetUnprocessedRawProducts failed: %v", err)
	}

	if len(retrieved) < 3 {
		t.Errorf("Expected at least 3 products, got %d", len(retrieved))
	}

	// Тест: Ограничение лимита
	retrievedLimited, err := adapter.GetUnprocessedRawProducts(2)
	if err != nil {
		t.Fatalf("GetUnprocessedRawProducts with limit failed: %v", err)
	}

	if len(retrievedLimited) > 2 {
		t.Errorf("Expected max 2 products, got %d", len(retrievedLimited))
	}

	// Очистка
	CleanupTestData(t, pg, []string{"raw_products"})
}

// TestScraperAdapter_MarkRawProductAsProcessed тестирует пометку товара как обработанного
func TestScraperAdapter_MarkRawProductAsProcessed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := SetupTestDB(t)
	defer pg.Close()

	adapter := NewScraperAdapter(pg)

	shopID := uuid.New().String()
	externalID := "ext-processed"

	// Создаём необработанный товар
	rawProduct := &scraper.RawProduct{
		ShopID:     shopID,
		ShopName:   "Test Shop",
		ExternalID: externalID,
		Name:       "Product to Process",
		Price:      100.0,
		Currency:   "RSD",
		InStock:    true,
		ParsedAt:   time.Now(),
	}

	if err := adapter.SaveRawProduct(rawProduct); err != nil {
		t.Fatalf("Failed to save raw product: %v", err)
	}

	// Проверяем, что товар необработан
	products, err := adapter.GetUnprocessedRawProducts(10)
	if err != nil {
		t.Fatalf("GetUnprocessedRawProducts failed: %v", err)
	}

	found := false
	for _, p := range products {
		if p.ShopID == shopID && p.ExternalID == externalID {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Product should be in unprocessed list")
	}

	// Помечаем как обработанный
	if err := adapter.MarkRawProductAsProcessed(shopID, externalID); err != nil {
		t.Fatalf("MarkRawProductAsProcessed failed: %v", err)
	}

	// Проверяем, что товар больше не в списке необработанных
	productsAfter, err := adapter.GetUnprocessedRawProducts(10)
	if err != nil {
		t.Fatalf("GetUnprocessedRawProducts after marking failed: %v", err)
	}

	for _, p := range productsAfter {
		if p.ShopID == shopID && p.ExternalID == externalID {
			t.Error("Product should not be in unprocessed list after marking")
		}
	}

	// Очистка
	CleanupTestData(t, pg, []string{"raw_products"})
}

