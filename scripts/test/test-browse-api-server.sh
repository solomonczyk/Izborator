#!/bin/bash
# Скрипт для тестирования /browse API endpoints на сервере

set -e

# Определяем URL API
# Если скрипт запускается на сервере через docker-compose exec, используем внутренний адрес
# Если запускается локально, используем внешний адрес сервера
API_BASE="${API_BASE:-http://backend:8080}"

# Если переменная SERVER_IP задана, используем внешний адрес
if [ -n "$SERVER_IP" ]; then
    API_BASE="https://${SERVER_IP}"
fi

echo "🔍 Тестирование /browse API endpoints на сервере"
echo "API Base: $API_BASE"
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для тестирования endpoint
test_endpoint() {
    local name="$1"
    local url="$2"
    local expected_status="${3:-200}"
    
    echo -n "Тест: $name... "
    
    # Используем curl с опцией -k для игнорирования SSL ошибок (если самоподписанный сертификат)
    response=$(curl -s -k -w "\n%{http_code}" "$url" 2>&1 || echo -e "\n000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ OK${NC} (HTTP $http_code)"
        
        # Проверяем структуру JSON
        if echo "$body" | jq . >/dev/null 2>&1; then
            items_count=$(echo "$body" | jq '.items | length' 2>/dev/null || echo "0")
            total=$(echo "$body" | jq '.total // 0' 2>/dev/null || echo "0")
            page=$(echo "$body" | jq '.page // 0' 2>/dev/null || echo "0")
            per_page=$(echo "$body" | jq '.per_page // 0' 2>/dev/null || echo "0")
            
            echo "   📊 Результаты: items=$items_count, total=$total, page=$page, per_page=$per_page"
            
            # Показываем первый товар, если есть
            if [ "$items_count" -gt 0 ]; then
                first_item=$(echo "$body" | jq '.items[0] | {id, name, category_id, shops_count}' 2>/dev/null)
                echo "   📦 Первый товар: $first_item"
            fi
        else
            echo -e "   ${YELLOW}⚠️  Ответ не является валидным JSON${NC}"
            echo "   Ответ: ${body:0:200}..."
        fi
    else
        echo -e "${RED}❌ FAILED${NC} (HTTP $http_code, ожидался $expected_status)"
        echo "   Ответ: ${body:0:200}..."
        return 1
    fi
    echo ""
}

# Тест 1: Health check
echo "🔍 Проверка здоровья API..."
health_response=$(curl -s -k "$API_BASE/api/health" 2>&1 || echo "ERROR")
if echo "$health_response" | grep -q "ok\|status"; then
    echo -e "${GREEN}✅ API работает${NC}"
    echo ""
else
    echo -e "${RED}❌ API не отвечает${NC}"
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
response=$(curl -s -k "$API_BASE/api/v1/products/browse?page=1&per_page=1")
if echo "$response" | jq 'has("items") and has("total") and has("page") and has("per_page")' | grep -q true; then
    echo -e "${GREEN}✅ Структура BrowseResult корректна${NC}"
    echo "   Поля: items, total, page, per_page"
else
    echo -e "${RED}❌ Структура BrowseResult некорректна${NC}"
    echo "   Ответ: ${response:0:200}..."
fi
echo ""

echo -e "${GREEN}✅ Все тесты завершены!${NC}"

