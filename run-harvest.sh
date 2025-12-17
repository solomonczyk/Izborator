#!/bin/bash
# 🏭 Скрипт для запуска "Фабрики" на продакшене
# Запускает полную цепочку: Discovery → Classifier → AutoConfig

set -e

echo "🏭 Project Horizon - Запуск 'Фабрики'"
echo "======================================"
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Шаг 1: Discovery
echo -e "${YELLOW}🔍 Шаг 1: Discovery (поиск кандидатов)${NC}"
echo "Запускаем поиск новых доменов..."
docker-compose run --rm backend ./discovery
echo -e "${GREEN}✅ Discovery завершен${NC}"
echo ""

# Небольшая пауза
sleep 2

# Шаг 2: Classifier
echo -e "${YELLOW}🔍 Шаг 2: Classifier (классификация)${NC}"
echo "Запускаем классификацию найденных доменов..."
docker-compose run --rm backend ./classifier -classify-all
echo -e "${GREEN}✅ Classifier завершен${NC}"
echo ""

# Проверяем количество классифицированных
CLASSIFIED_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT COUNT(*) FROM potential_shops WHERE status = 'classified';" | tr -d ' ')
echo "Классифицировано магазинов: $CLASSIFIED_COUNT"

if [ "$CLASSIFIED_COUNT" -eq "0" ]; then
    echo -e "${RED}❌ Нет классифицированных магазинов для AutoConfig!${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}🤖 Шаг 3: AutoConfig (AI генерация селекторов)${NC}"
echo "Запускаем генерацию конфигов для 5 магазинов..."
docker-compose run --rm backend ./autoconfig -limit 5
echo -e "${GREEN}✅ AutoConfig завершен${NC}"
echo ""

# Шаг 4: Проверка результатов
echo -e "${YELLOW}📊 Шаг 4: Проверка результатов${NC}"
echo ""
echo "Созданные магазины (AutoConfig):"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    name, 
    base_url, 
    is_active,
    is_auto_configured,
    ai_config_model,
    created_at
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC
LIMIT 10;
"

echo ""
echo "Статистика по статусам:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count
FROM potential_shops
GROUP BY status
ORDER BY status;
"

echo ""
echo -e "${GREEN}✅ 'Фабрика' завершила работу!${NC}"

