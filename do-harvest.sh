#!/bin/bash
# Простой и надежный скрипт для выполнения Harvest

set -e

cd ~/Izborator

echo "🏭 Project Horizon - Harvest"
echo "============================"
echo ""

# 1. Миграции
echo "📦 Шаг 1: Миграции..."
docker-compose run --rm backend ./migrate
echo ""

# 2. Переменные
echo "📝 Шаг 2: Переменные..."
export $(cat .env | grep -v '^#' | xargs)
echo "✅ Загружены"
echo ""

# 3. Classifier
echo "🔍 Шаг 3: Classifier..."
docker-compose run --rm backend ./classifier -classify-all
echo ""

# 4. Проверка
echo "📊 Проверка после Classifier:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "SELECT status, COUNT(*) FROM potential_shops GROUP BY status;"
echo ""

# 5. AutoConfig
CLASSIFIED=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT COUNT(*) FROM potential_shops WHERE status = 'classified';" | tr -d ' ')

if [ "$CLASSIFIED" -gt "0" ]; then
    echo "🤖 Шаг 4: AutoConfig ($CLASSIFIED магазинов)..."
    docker-compose run --rm \
      -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
      -e OPENAI_MODEL="gpt-4o-mini" \
      backend ./autoconfig -limit 5
    echo ""
fi

# 6. Результаты
echo "📊 Финальные результаты:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT name, base_url, is_auto_configured, created_at 
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC 
LIMIT 10;
"

