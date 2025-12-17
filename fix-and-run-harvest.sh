#!/bin/bash
# Скрипт для исправления проблем и запуска Harvest

set -e

echo "🔧 Исправление проблем и запуск Harvest"
echo "========================================"
echo ""

# Шаг 1: Применение миграций
echo "📦 Шаг 1: Применение миграций..."
docker-compose run --rm backend ./migrate || echo "⚠️  Миграции уже применены или ошибка"
echo "✅ Миграции применены"
echo ""

# Шаг 2: Загрузка переменных
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "✅ Переменные окружения загружены"
else
    echo "⚠️  Файл .env не найден"
fi
echo ""

# Шаг 3: Classifier
echo "🔍 Шаг 2: Classifier (классификация 85 кандидатов)..."
docker-compose run --rm backend ./classifier -classify-all || echo "⚠️  Classifier завершился с предупреждениями"

CLASSIFIED_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT COUNT(*) FROM potential_shops WHERE status = 'classified';" | tr -d ' ')
echo "✅ Классифицировано магазинов: $CLASSIFIED_COUNT"
echo ""

if [ "$CLASSIFIED_COUNT" -eq "0" ]; then
    echo "❌ Нет классифицированных магазинов для AutoConfig!"
    echo "Проверь логи Classifier выше"
    exit 1
fi

# Шаг 4: AutoConfig
echo "🤖 Шаг 3: AutoConfig (AI генерация селекторов)..."
echo "Обрабатываем 5 магазинов..."
docker-compose run --rm \
  -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
  -e OPENAI_MODEL="${OPENAI_MODEL:-gpt-4o-mini}" \
  backend ./autoconfig -limit 5 || echo "⚠️  AutoConfig завершился с предупреждениями"
echo "✅ AutoConfig завершен"
echo ""

# Шаг 5: Проверка результатов
echo "📊 Шаг 4: Проверка результатов..."
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
echo "✅ Готово!"

