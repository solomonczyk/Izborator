package storage

import (
	"testing"
)

// TestMatchingAdapter_FindSimilarProducts тестирует поиск похожих товаров
func TestMatchingAdapter_FindSimilarProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// TODO: Настроить тестовую БД
	// pg := setupTestDB(t)
	// adapter := NewMatchingAdapter(pg)
	
	// products, err := adapter.FindSimilarProducts("iPhone 15", "Apple", 10)
	// if err != nil {
	// 	t.Fatalf("FindSimilarProducts failed: %v", err)
	// }
	
	// if len(products) == 0 {
	// 	t.Log("No similar products found (expected in empty test DB)")
	// }
	
	t.Log("Integration test placeholder - requires test DB setup")
}

// TestMatchingAdapter_SaveMatch тестирует сохранение сопоставления
func TestMatchingAdapter_SaveMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	t.Log("Integration test placeholder - requires test DB setup")
}

