#!/bin/sh
# Скрипт для тестирования /browse API endpoints на сервере (совместим с Alpine sh)

set -e

# Используем внутренний адрес Docker network
API_BASE="${API_BASE:-http://backend:8080}"

echo "🔍 Тестирование /browse API endpoints на сервере"
echo "API Base: $API_BASE"
echo ""

# Функция для тестирования endpoint
test_endpoint() {
    local name="$1"
    local url="$2"
    local expected_status="${3:-200}"
    
    printf "Тест: %s... " "$name"
    
    # Используем curl (должен быть в Alpine)
    response=$(curl -s -w "\n%{http_code}" "$url" 2>&1 || echo -e "\n000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_status" ]; then
        printf "✅ OK (HTTP %s)\n" "$http_code"
        
        # Проверяем структуру JSON (если jq доступен)
        if command -v jq >/dev/null 2>&1; then
            items_count=$(echo "$body" | jq '.items | length' 2>/dev/null || echo "0")
            total=$(echo "$body" | jq '.total // 0' 2>/dev/null || echo "0")
            page=$(echo "$body" | jq '.page // 0' 2>/dev/null || echo "0")
            per_page=$(echo "$body" | jq '.per_page // 0' 2>/dev/null || echo "0")
            
            echo "   📊 Результаты: items=$items_count, total=$total, page=$page, per_page=$per_page"
            
            if [ "$items_count" -gt 0 ]; then
                first_item=$(echo "$body" | jq '.items[0] | {id, name, category_id, shops_count}' 2>/dev/null)
                echo "   📦 Первый товар: $first_item"
            fi
        else
            # Если jq нет, просто показываем начало ответа
            echo "   📄 Ответ (первые 200 символов): ${body#*?}"
        fi
    else
        printf "❌ FAILED (HTTP %s, ожидался %s)\n" "$http_code" "$expected_status"
        echo "   Ответ: $(echo "$body" | head -c 200)..."
        return 1
    fi
    echo ""
}

# Тест 1: Health check
echo "🔍 Проверка здоровья API..."
health_response=$(curl -s "$API_BASE/api/health" 2>&1 || echo "ERROR")
if echo "$health_response" | grep -q "ok\|status"; then
    echo "✅ API работает"
    echo ""
else
    echo "❌ API не отвечает"
    echo "   Ответ: $health_response"
    echo ""
    exit 1
fi

# Тест 2: Browse без фильтров
test_endpoint \
    "GET /api/v1/products/browse (без фильтра)" \
    "$API_BASE/api/v1/products/browse?page=1&per_page=5"

# Тест 3: Browse с категорией mobilni-telefoni
test_endpoint \
    "GET /api/v1/products/browse?category=mobilni-telefoni" \
    "$API_BASE/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=5"

# Тест 4: Browse с категорией laptopovi
test_endpoint \
    "GET /api/v1/products/browse?category=laptopovi" \
    "$API_BASE/api/v1/products/browse?category=laptopovi&page=1&per_page=5"

# Тест 5: Browse с несуществующей категорией (fallback)
test_endpoint \
    "GET /api/v1/products/browse?category=neexistujuca-kategorija (fallback)" \
    "$API_BASE/api/v1/products/browse?category=neexistujuca-kategorija&page=1&per_page=5" \
    "200"

# Тест 6: Проверка структуры BrowseResult
echo "🔍 Проверка структуры BrowseResult..."
response=$(curl -s "$API_BASE/api/v1/products/browse?page=1&per_page=1")
if echo "$response" | grep -q "\"items\"" && echo "$response" | grep -q "\"total\"" && echo "$response" | grep -q "\"page\"" && echo "$response" | grep -q "\"per_page\""; then
    echo "✅ Структура BrowseResult корректна"
    echo "   Поля: items, total, page, per_page"
else
    echo "❌ Структура BrowseResult некорректна"
    echo "   Ответ: $(echo "$response" | head -c 200)..."
fi
echo ""

echo "✅ Все тесты завершены!"

