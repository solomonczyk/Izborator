package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/products"
)

// TestProcessorAdapter_SaveProduct тестирует сохранение товара
func TestProcessorAdapter_SaveProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	adapter := NewProcessorAdapter(pg, nil)

	// Тест 1: Сохранение нового товара
	t.Run("SaveNewProduct", func(t *testing.T) {
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

		if err := adapter.SaveProduct(product); err != nil {
			t.Fatalf("Failed to save product: %v", err)
		}
	})

	// Тест 2: Обновление существующего товара (ON CONFLICT)
	t.Run("UpdateExistingProduct", func(t *testing.T) {
		productID := uuid.New().String()

		// Первое сохранение
		product1 := &products.Product{
			ID:          productID,
			Name:        "Old Name",
			Description: "Old Description",
			Brand:       "Old Brand",
			Category:    "Old Category",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := adapter.SaveProduct(product1); err != nil {
			t.Fatalf("Failed to save product (first time): %v", err)
		}

		// Обновление
		product2 := &products.Product{
			ID:          productID,
			Name:        "New Name",
			Description: "New Description",
			Brand:       "New Brand",
			Category:    "New Category",
			CreatedAt:   product1.CreatedAt, // Сохраняем оригинальную дату создания
			UpdatedAt:   time.Now(),
		}

		if err := adapter.SaveProduct(product2); err != nil {
			t.Fatalf("Failed to update product: %v", err)
		}
	})

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_prices"})
}

// TestProcessorAdapter_SavePrice тестирует сохранение цены
func TestProcessorAdapter_SavePrice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	adapter := NewProcessorAdapter(pg, nil)

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

	// Тест 1: Сохранение новой цены
	t.Run("SaveNewPrice", func(t *testing.T) {
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

		if err := adapter.SavePrice(price); err != nil {
			t.Fatalf("Failed to save price: %v", err)
		}
	})

	// Тест 2: Обновление существующей цены (ON CONFLICT)
	t.Run("UpdateExistingPrice", func(t *testing.T) {
		shopID := uuid.New().String()

		// Первое сохранение
		price1 := &products.ProductPrice{
			ProductID: product.ID,
			ShopID:    shopID,
			ShopName:  "Test Shop",
			Price:     100.0,
			Currency:  "RSD",
			InStock:   true,
			UpdatedAt: time.Now(),
		}

		if err := adapter.SavePrice(price1); err != nil {
			t.Fatalf("Failed to save price (first time): %v", err)
		}

		// Обновление цены
		price2 := &products.ProductPrice{
			ProductID: product.ID,
			ShopID:    shopID,
			ShopName:  "Test Shop Updated",
			Price:     120.0,
			Currency:  "RSD",
			InStock:   false,
			UpdatedAt: time.Now(),
		}

		if err := adapter.SavePrice(price2); err != nil {
			t.Fatalf("Failed to update price: %v", err)
		}
	})

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_prices"})
}

