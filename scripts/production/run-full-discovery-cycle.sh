#!/bin/bash
# Полный цикл Discovery на продакшене: Discovery -> Classifier -> AutoConfig
# Использование: ./run-full-discovery-cycle.sh [limit-autoconfig]

set -e

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Параметры
LIMIT_AUTOCONFIG=${1:-5}

echo -e "${BLUE}🚀 Запуск полного цикла Discovery на продакшене${NC}"
echo "=============================================="
echo ""
echo "Параметры:"
echo "  - Лимит для AutoConfig: $LIMIT_AUTOCONFIG"
echo ""

# Определяем версию docker-compose
DOCKER_COMPOSE="docker compose"
if ! docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
fi

# Шаг 1: Discovery
echo -e "${YELLOW}[STEP 1] Запуск Discovery (поиск кандидатов)...${NC}"
echo ""

$DOCKER_COMPOSE run --rm backend ./discovery || {
    echo -e "${YELLOW}⚠️  Discovery завершился с предупреждениями${NC}"
}

echo ""
echo -e "${GREEN}✅ Discovery завершен${NC}"
echo ""

# Шаг 2: Classifier
echo -e "${YELLOW}[STEP 2] Запуск Classifier (классификация)...${NC}"
echo ""

$DOCKER_COMPOSE run --rm backend ./classifier -classify-all || {
    echo -e "${YELLOW}⚠️  Classifier завершился с предупреждениями${NC}"
}

echo ""
echo -e "${GREEN}✅ Classifier завершен${NC}"
echo ""

# Шаг 3: AutoConfig
echo -e "${YELLOW}[STEP 3] Запуск AutoConfig (AI генерация селекторов)...${NC}"
echo "Обрабатываем $LIMIT_AUTOCONFIG кандидатов..."
echo ""

$DOCKER_COMPOSE run --rm \
    -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
    -e OPENAI_MODEL="${OPENAI_MODEL:-gpt-4o-mini}" \
    backend ./autoconfig -limit $LIMIT_AUTOCONFIG || {
    echo -e "${YELLOW}⚠️  AutoConfig завершился с предупреждениями${NC}"
}

echo ""
echo -e "${GREEN}✅ AutoConfig завершен${NC}"
echo ""

echo -e "${GREEN}✅ Полный цикл Discovery завершен!${NC}"
echo ""

