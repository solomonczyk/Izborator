#!/bin/sh
# Одна команда для тестирования всех endpoints

API="http://backend:8080"

echo "=== 1. Health ===" && curl -s "$API/api/health" && echo -e "\n"
echo "=== 2. Browse (без фильтра) ===" && curl -s "$API/api/v1/products/browse?page=1&per_page=2" && echo -e "\n"
echo "=== 3. Browse (category=mobilni-telefoni) ===" && curl -s "$API/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=2" && echo -e "\n"
echo "=== 4. Browse (category=laptopovi) ===" && curl -s "$API/api/v1/products/browse?category=laptopovi&page=1&per_page=2" && echo -e "\n"
echo "=== 5. Browse (несуществующая категория) ===" && curl -s "$API/api/v1/products/browse?category=neexistujuca&page=1&per_page=2" && echo -e "\n"
echo "✅ Готово!"

