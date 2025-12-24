package storage

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/solomonczyk/izborator/internal/matching"
	"github.com/solomonczyk/izborator/internal/products"
)

// TestMatchingAdapter_FindSimilarProducts тестирует поиск похожих товаров
func TestMatchingAdapter_FindSimilarProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	adapter := NewMatchingAdapter(pg)
	productsAdapter := NewProductsAdapter(pg, nil, nil)

	// Создаём тестовые товары
	testProducts := []*products.Product{
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
			Name:        "iPhone 15",
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

	for _, p := range testProducts {
		if err := productsAdapter.SaveProduct(p); err != nil {
			t.Fatalf("Failed to save product: %v", err)
		}
	}

	// Тест 1: Поиск по точному названию и бренду
	t.Run("ExactMatch", func(t *testing.T) {
		results, err := adapter.FindSimilarProducts("iPhone 15 Pro", "Apple", 10)
		if err != nil {
			t.Fatalf("FindSimilarProducts failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least 1 result for exact match")
		}

		found := false
		for _, p := range results {
			if p.Name == "iPhone 15 Pro" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find 'iPhone 15 Pro' in results")
		}
	})

	// Тест 2: Поиск по частичному совпадению
	t.Run("PartialMatch", func(t *testing.T) {
		results, err := adapter.FindSimilarProducts("iPhone", "Apple", 10)
		if err != nil {
			t.Fatalf("FindSimilarProducts failed: %v", err)
		}

		if len(results) < 2 {
			t.Errorf("Expected at least 2 results for 'iPhone', got %d", len(results))
		}
	})

	// Тест 3: Поиск без бренда
	t.Run("NoBrand", func(t *testing.T) {
		results, err := adapter.FindSimilarProducts("iPhone 15", "", 10)
		if err != nil {
			t.Fatalf("FindSimilarProducts failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least 1 result even without brand")
		}
	})

	// Тест 4: Ограничение лимита
	t.Run("Limit", func(t *testing.T) {
		results, err := adapter.FindSimilarProducts("iPhone", "Apple", 1)
		if err != nil {
			t.Fatalf("FindSimilarProducts failed: %v", err)
		}

		if len(results) > 1 {
			t.Errorf("Expected max 1 result, got %d", len(results))
		}
	})

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_matches"})
}

// TestMatchingAdapter_SaveMatch тестирует сохранение сопоставления
func TestMatchingAdapter_SaveMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	adapter := NewMatchingAdapter(pg)
	productsAdapter := NewProductsAdapter(pg, nil, nil)

	// Создаём два товара
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

	if err := productsAdapter.SaveProduct(product1); err != nil {
		t.Fatalf("Failed to save product1: %v", err)
	}
	if err := productsAdapter.SaveProduct(product2); err != nil {
		t.Fatalf("Failed to save product2: %v", err)
	}

	// Тест 1: Сохранение нового сопоставления
	t.Run("SaveNewMatch", func(t *testing.T) {
		match := &matching.ProductMatch{
			ProductID:  product1.ID,
			MatchedID:  product2.ID,
			Similarity: 0.85,
			Confidence: "high",
			MatchedAt:  time.Now(),
		}

		if err := adapter.SaveMatch(match); err != nil {
			t.Fatalf("SaveMatch failed: %v", err)
		}
	})

	// Тест 2: Обновление существующего сопоставления (ON CONFLICT)
	t.Run("UpdateExistingMatch", func(t *testing.T) {
		// Первое сохранение
		match1 := &matching.ProductMatch{
			ProductID:  product1.ID,
			MatchedID:  product2.ID,
			Similarity: 0.75,
			Confidence: "medium",
			MatchedAt:  time.Now(),
		}

		if err := adapter.SaveMatch(match1); err != nil {
			t.Fatalf("SaveMatch (first time) failed: %v", err)
		}

		// Обновление
		match2 := &matching.ProductMatch{
			ProductID:  product1.ID,
			MatchedID:  product2.ID,
			Similarity: 0.90,
			Confidence: "high",
			MatchedAt:  time.Now(),
		}

		if err := adapter.SaveMatch(match2); err != nil {
			t.Fatalf("SaveMatch (update) failed: %v", err)
		}
	})

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_matches"})
}

// TestMatchingAdapter_GetMatches тестирует получение сопоставлений
func TestMatchingAdapter_GetMatches(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pg := setupTestDB(t)
	defer pg.Close()

	adapter := NewMatchingAdapter(pg)
	productsAdapter := NewProductsAdapter(pg, nil, nil)

	// Создаём товары
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

	product3 := &products.Product{
		ID:          uuid.New().String(),
		Name:        "Product 3",
		Brand:       "Brand C",
		Category:    "Category 1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, p := range []*products.Product{product1, product2, product3} {
		if err := productsAdapter.SaveProduct(p); err != nil {
			t.Fatalf("Failed to save product: %v", err)
		}
	}

	// Создаём сопоставления
	matches := []*matching.ProductMatch{
		{
			ProductID:  product1.ID,
			MatchedID:  product2.ID,
			Similarity: 0.90,
			Confidence: "high",
			MatchedAt:  time.Now(),
		},
		{
			ProductID:  product1.ID,
			MatchedID:  product3.ID,
			Similarity: 0.75,
			Confidence: "medium",
			MatchedAt:  time.Now(),
		},
	}

	for _, m := range matches {
		if err := adapter.SaveMatch(m); err != nil {
			t.Fatalf("Failed to save match: %v", err)
		}
	}

	// Тест: Получение всех сопоставлений для товара
	retrieved, err := adapter.GetMatches(product1.ID)
	if err != nil {
		t.Fatalf("GetMatches failed: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(retrieved))
	}

	// Проверяем сортировку (по убыванию similarity)
	if len(retrieved) >= 2 {
		if retrieved[0].Similarity < retrieved[1].Similarity {
			t.Error("Matches should be sorted by similarity DESC")
		}
	}

	// Очистка
	cleanupTestData(t, pg, []string{"products", "product_matches"})
}

