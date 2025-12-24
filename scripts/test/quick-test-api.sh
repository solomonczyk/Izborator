#!/bin/sh
# Быстрое тестирование API - одна команда для выполнения на сервере

API="http://backend:8080"

echo "=== Тест 1: Health ==="
curl -s "$API/api/health" && echo ""

echo ""
echo "=== Тест 2: Browse (без фильтра) ==="
curl -s "$API/api/v1/products/browse?page=1&per_page=2" | head -c 300 && echo ""

echo ""
echo "=== Тест 3: Browse (category=mobilni-telefoni) ==="
curl -s "$API/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=2" | head -c 300 && echo ""

echo ""
echo "=== Тест 4: Browse (category=laptopovi) ==="
curl -s "$API/api/v1/products/browse?category=laptopovi&page=1&per_page=2" | head -c 300 && echo ""

echo ""
echo "=== Тест 5: Browse (несуществующая категория) ==="
curl -s "$API/api/v1/products/browse?category=neexistujuca&page=1&per_page=2" | head -c 300 && echo ""

echo ""
echo "✅ Готово!"

