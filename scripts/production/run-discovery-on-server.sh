#!/bin/bash
# Скрипт для запуска Discovery на продакшен сервере через SSH
# Использование: ./run-discovery-on-server.sh [max-results] [delay]

set -e

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Параметры
MAX_RESULTS=${1:-200}
DELAY=${2:-1}
SERVER="root@152.53.227.37"

echo -e "${BLUE}🔍 Запуск Discovery Worker на продакшене${NC}"
echo "=============================================="
echo ""
echo "Сервер: $SERVER"
echo "Параметры:"
echo "  - Максимум результатов на запрос: $MAX_RESULTS"
echo "  - Задержка между запросами: ${DELAY}s"
echo ""

# Запуск через SSH
echo -e "${BLUE}🚀 Подключение к серверу и запуск Discovery...${NC}"
echo ""

ssh $SERVER << EOF
cd ~/Izborator

# Проверка переменных окружения
if [ -z "\$GOOGLE_API_KEY" ] || [ -z "\$GOOGLE_CX" ]; then
    echo -e "${YELLOW}⚠️  GOOGLE_API_KEY или GOOGLE_CX не установлены${NC}"
    echo "   Проверьте .env файл"
    exit 1
fi

# Определяем версию docker-compose
DOCKER_COMPOSE="docker compose"
if ! docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
fi

# Запуск Discovery
echo -e "${BLUE}🚀 Запуск Discovery...${NC}"
$DOCKER_COMPOSE run --rm backend ./discovery -max-results $MAX_RESULTS -delay ${DELAY}s

echo ""
echo -e "${GREEN}✅ Discovery завершен!${NC}"
EOF

echo ""
echo -e "${GREEN}✅ Discovery выполнен на сервере!${NC}"
echo ""
echo "Следующие шаги:"
echo "  1. Проверить найденные кандидаты:"
echo "     ./scripts/production/check-autoconfig-on-server.sh"
echo ""
echo "  2. Классифицировать найденные сайты:"
echo "     ssh $SERVER 'cd ~/Izborator && docker compose run --rm backend ./classifier -classify-all -limit 100'"
echo ""
