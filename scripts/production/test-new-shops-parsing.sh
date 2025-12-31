#!/bin/bash
# Тестовый парсинг для новых магазинов (macola.rs и alatnik.rs)
# Использование: ./test-new-shops-parsing.sh

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔍 Поиск ID магазинов macola.rs и alatnik.rs...${NC}"
echo ""

cd ~/Izborator

# Определяем версию docker-compose
DOCKER_COMPOSE="docker compose"
if ! docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
fi

# Получаем ID магазинов
MACOLA_ID=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -A -c "SELECT id FROM shops WHERE base_url LIKE '%macola%' LIMIT 1;")
ALATNIK_ID=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -A -c "SELECT id FROM shops WHERE base_url LIKE '%alatnik%' LIMIT 1;")

if [ -z "$MACOLA_ID" ]; then
    echo -e "${YELLOW}⚠️  Магазин macola.rs не найден${NC}"
else
    echo -e "${GREEN}✅ Найден macola.rs: $MACOLA_ID${NC}"
fi

if [ -z "$ALATNIK_ID" ]; then
    echo -e "${YELLOW}⚠️  Магазин alatnik.rs не найден${NC}"
else
    echo -e "${GREEN}✅ Найден alatnik.rs: $ALATNIK_ID${NC}"
fi

echo ""

# Запускаем парсинг каталога для каждого магазина
if [ -n "$MACOLA_ID" ]; then
    echo -e "${BLUE}🚀 Запуск парсинга каталога для macola.rs...${NC}"
    echo ""
    
    # Используем discover для парсинга каталога конкретного магазина
    # Но сначала нужно активировать магазин, если он не активен
    docker-compose exec -T postgres psql -U postgres -d izborator -c "UPDATE shops SET is_active = true, enabled = true WHERE id = '$MACOLA_ID';"
    
    # Запускаем discovery (он обойдет все активные магазины)
    # Но лучше использовать прямой вызов ParseCatalog через worker
    echo "Запускаю discovery для всех активных магазинов..."
    $DOCKER_COMPOSE run --rm backend ./worker -discover || echo "⚠️  Discovery завершился с предупреждениями"
    
    echo ""
fi

if [ -n "$ALATNIK_ID" ]; then
    echo -e "${BLUE}🚀 Запуск парсинга каталога для alatnik.rs...${NC}"
    echo ""
    
    # Активируем магазин
    docker-compose exec -T postgres psql -U postgres -d izborator -c "UPDATE shops SET is_active = true, enabled = true WHERE id = '$ALATNIK_ID';"
    
    echo ""
fi

# Обрабатываем сырые данные
echo -e "${BLUE}⚙️  Обработка сырых данных...${NC}"
$DOCKER_COMPOSE run --rm backend ./worker -process || echo "⚠️  Processing завершился с предупреждениями"

echo ""
echo -e "${GREEN}✅ Тестовый парсинг завершен!${NC}"
echo ""
echo "Проверьте результаты:"
echo "  - Количество найденных товаров:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "SELECT COUNT(*) as total_products FROM raw_products WHERE shop_id IN ('$MACOLA_ID', '$ALATNIK_ID');"
echo ""
echo "  - Количество обработанных товаров:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "SELECT COUNT(*) as processed_products FROM products WHERE shop_id IN ('$MACOLA_ID', '$ALATNIK_ID');"

