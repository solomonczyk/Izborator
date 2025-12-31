#!/bin/bash
# Тестовый парсинг ("сбор урожая") для новых магазинов
# macola.rs и alatnik.rs

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

cd ~/Izborator

MACOLA_ID="ef60e5e1-8624-4dc0-9de7-656ba3efa482"
ALATNIK_ID="102179a0-d568-4f71-9351-da2474620fe9"

# Определяем версию docker-compose
DOCKER_COMPOSE="docker compose"
if ! docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
fi

echo -e "${BLUE}[STEP 1] Активация магазинов...${NC}"
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
UPDATE shops 
SET is_active = true, enabled = true 
WHERE id IN ('$MACOLA_ID', '$ALATNIK_ID');
SQL

echo ""
echo -e "${BLUE}[STEP 2] Запуск discovery для парсинга каталогов...${NC}"
$DOCKER_COMPOSE run --rm backend ./worker -discover || echo -e "${YELLOW}⚠️  Discovery завершился с предупреждениями${NC}"

echo ""
echo -e "${BLUE}[STEP 3] Обработка сырых данных...${NC}"
$DOCKER_COMPOSE run --rm backend ./worker -process || echo -e "${YELLOW}⚠️  Processing завершился с предупреждениями${NC}"

echo ""
echo -e "${BLUE}[STEP 4] Индексация в Meilisearch...${NC}"
$DOCKER_COMPOSE run --rm backend ./worker -reindex || echo -e "${YELLOW}⚠️  Reindex завершился с предупреждениями${NC}"

echo ""
echo -e "${GREEN}[STEP 5] Проверка результатов...${NC}"
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    s.name as shop_name,
    s.base_url,
    COUNT(DISTINCT rp.id) as raw_products,
    COUNT(DISTINCT p.id) as processed_products
FROM shops s
LEFT JOIN raw_products rp ON rp.shop_id = s.id
LEFT JOIN products p ON p.shop_id = s.id
WHERE s.id IN ('$MACOLA_ID', '$ALATNIK_ID')
GROUP BY s.name, s.base_url;
SQL

echo ""
echo -e "${GREEN}✅ Тестовый парсинг завершен!${NC}"
echo ""
echo "Следующие шаги:"
echo "  1. Проверить товары на фронтенде"
echo "  2. Убедиться, что цены парсятся корректно"
echo "  3. Проверить категории товаров"

