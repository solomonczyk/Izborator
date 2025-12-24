package storage

import (
	"testing"
)

// TestCategoriesAdapter_GetTree тестирует получение дерева категорий
func TestCategoriesAdapter_GetTree(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// TODO: Настроить тестовую БД
	// pg := setupTestDB(t)
	// adapter := NewCategoriesAdapter(pg)
	
	// tree, err := adapter.GetTree()
	// if err != nil {
	// 	t.Fatalf("GetTree failed: %v", err)
	// }
	
	// if tree == nil {
	// 	t.Error("GetTree returned nil")
	// }
	
	t.Log("Integration test placeholder - requires test DB setup")
}

// TestCategoriesAdapter_GetBySlug тестирует получение категории по slug
func TestCategoriesAdapter_GetBySlug(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	t.Log("Integration test placeholder - requires test DB setup")
}

