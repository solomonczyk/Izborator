#!/bin/bash
# Простой скрипт для исправления и запуска Harvest

set -e

echo "🔧 Исправление и запуск Harvest"
echo "================================"
echo ""

# Шаг 1: Миграции
echo "📦 Шаг 1: Применение миграций..."
docker-compose run --rm backend ./migrate 2>&1 | tail -5
echo ""

# Шаг 2: Переменные
echo "📝 Шаг 2: Загрузка переменных окружения..."
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "✅ Переменные загружены"
else
    echo "❌ Файл .env не найден!"
    exit 1
fi
echo ""

# Шаг 3: Classifier
echo "🔍 Шаг 3: Запуск Classifier на всех кандидатах..."
docker-compose run --rm backend ./classifier -classify-all
echo ""

# Шаг 4: Проверка
echo "📊 Проверка результатов Classifier..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT status, COUNT(*) as count 
FROM potential_shops 
GROUP BY status 
ORDER BY status;
"
echo ""

# Шаг 5: AutoConfig (если есть classified)
CLASSIFIED=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT COUNT(*) FROM potential_shops WHERE status = 'classified';" | tr -d ' ')

if [ "$CLASSIFIED" -gt "0" ]; then
    echo "🤖 Шаг 4: Запуск AutoConfig ($CLASSIFIED классифицированных магазинов)..."
    docker-compose run --rm \
      -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
      -e OPENAI_MODEL="${OPENAI_MODEL:-gpt-4o-mini}" \
      backend ./autoconfig -limit 5
    echo ""
    
    echo "📊 Финальные результаты:"
    bash check-harvest-results.sh
else
    echo "⚠️  Нет классифицированных магазинов для AutoConfig"
    echo "Проверь логи Classifier выше"
fi

