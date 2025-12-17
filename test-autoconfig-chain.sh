#!/bin/bash
# Скрипт для тестирования полной цепочки Project Horizon

set -e

echo "🚀 Project Horizon - Финальный Тест-Драйв"
echo "=========================================="
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Шаг 1: Discovery (если нужно)
echo -e "${YELLOW}Шаг 1: Discovery (поиск кандидатов)${NC}"
echo "Проверяем, есть ли уже кандидаты в БД..."
CANDIDATES_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT COUNT(*) FROM potential_shops WHERE status = 'new';" | tr -d ' ')

if [ "$CANDIDATES_COUNT" -eq "0" ]; then
    echo "Кандидатов нет. Запускаем Discovery..."
    docker-compose run --rm backend ./discovery
    echo -e "${GREEN}✅ Discovery завершен${NC}"
else
    echo -e "${GREEN}✅ Найдено кандидатов: $CANDIDATES_COUNT (пропускаем Discovery)${NC}"
fi

echo ""
echo -e "${YELLOW}Шаг 2: Classifier (классификация)${NC}"
echo "Запускаем классификатор на всех найденных доменах..."
docker-compose run --rm backend ./classifier -classify-all -limit 10

CLASSIFIED_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT COUNT(*) FROM potential_shops WHERE status = 'classified';" | tr -d ' ')
echo -e "${GREEN}✅ Классифицировано магазинов: $CLASSIFIED_COUNT${NC}"

if [ "$CLASSIFIED_COUNT" -eq "0" ]; then
    echo -e "${RED}❌ Нет классифицированных магазинов для AutoConfig!${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Шаг 3: AutoConfig (AI генерация селекторов) 🧠${NC}"
echo "Запускаем AutoConfig на 1 кандидате (для теста)..."
docker-compose run --rm backend ./autoconfig -limit 1

echo ""
echo -e "${YELLOW}Шаг 4: Проверка результата${NC}"
echo "Проверяем созданные магазины..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    name, 
    base_url, 
    is_auto_configured,
    ai_config_model,
    selectors->>'name' as name_selector,
    selectors->>'price' as price_selector,
    created_at
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC 
LIMIT 1;
"

echo ""
echo -e "${GREEN}✅ Тест-драйв завершен!${NC}"

