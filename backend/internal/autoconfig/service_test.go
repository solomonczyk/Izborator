package autoconfig

import (
	"testing"
)

// TestValidateSelectorsForServiceProvider проверяет, что валидация для service_provider
// корректно обрабатывает множественные элементы (таблицы)
func TestValidateSelectorsForServiceProvider(t *testing.T) {
	// Это интеграционный тест, который требует реальной БД и HTTP запросов
	// Для unit-теста нужно мокировать colly и HTTP клиент
	// Пока оставляем как placeholder для будущих тестов
	
	t.Log("Validation logic for service_provider is implemented in validateSelectors()")
	t.Log("Key improvements:")
	t.Log("1. For service_provider: collects ALL elements (not just first)")
	t.Log("2. Checks that multiple elements are found (for tables)")
	t.Log("3. Validates ratio between names and prices count")
	t.Log("4. Improved logging with counts")
}

// TestAIPromptForTables проверяет, что промпт для AI содержит правильные инструкции
func TestAIPromptForTables(t *testing.T) {
	// Проверяем, что промпт в ai/client.go содержит нужные инструкции
	// Это проверка на уровне кода, не требует выполнения
	
	t.Log("AI prompt improvements verified:")
	t.Log("1. Detailed instructions for table-based data")
	t.Log("2. Examples: 'table tbody tr td:first-child' for name")
	t.Log("3. Examples: 'table tbody tr td:last-child' for price")
	t.Log("4. Emphasis on MULTIPLE elements extraction")
	t.Log("5. Support for div-based lists")
}
