#!/bin/sh
# Прямое тестирование API через curl (для выполнения на сервере)

API_BASE="http://backend:8080"

echo "🔍 Тестирование /browse API endpoints"
echo "API Base: $API_BASE"
echo ""

# Health check
echo "1. Health check..."
curl -s "$API_BASE/api/health"
echo ""
echo ""

# Browse без фильтра
echo "2. Browse без фильтра..."
curl -s "$API_BASE/api/v1/products/browse?page=1&per_page=5" | head -c 500
echo ""
echo ""

# Browse с категорией mobilni-telefoni
echo "3. Browse с категорией mobilni-telefoni..."
curl -s "$API_BASE/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=5" | head -c 500
echo ""
echo ""

# Browse с категорией laptopovi
echo "4. Browse с категорией laptopovi..."
curl -s "$API_BASE/api/v1/products/browse?category=laptopovi&page=1&per_page=5" | head -c 500
echo ""
echo ""

# Browse с несуществующей категорией
echo "5. Browse с несуществующей категорией (fallback)..."
curl -s "$API_BASE/api/v1/products/browse?category=neexistujuca-kategorija&page=1&per_page=5" | head -c 500
echo ""
echo ""

echo "✅ Тесты завершены!"

