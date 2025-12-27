#!/bin/bash
# Скрипт для тестирования AutoConfig на продакшен сервере
# Запускать на сервере: ssh root@152.53.227.37

set -e

echo "🧪 Тестирование AutoConfig для табличных данных"
echo "================================================"
echo ""

# Переходим в директорию проекта
cd ~/Izborator 2>/dev/null || cd /root/Izborator 2>/dev/null || {
    echo "❌ Не удалось найти директорию проекта"
    exit 1
}

# Используем docker compose (новый синтаксис) или docker-compose (старый)
DOCKER_COMPOSE="docker compose"
if ! docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
fi

# Проверяем наличие классифицированных кандидатов
CLASSIFIED_COUNT=$($DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -t -A -c "
SELECT COUNT(*) 
FROM potential_shops 
WHERE status = 'classified';
" 2>/dev/null | tr -d ' \n')

if [ -z "$CLASSIFIED_COUNT" ] || [ "$CLASSIFIED_COUNT" -eq "0" ]; then
    echo "❌ Нет классифицированных кандидатов для тестирования"
    echo ""
    echo "   Сначала запустите:"
    echo "   1. Discovery: $DOCKER_COMPOSE run --rm backend ./discovery -max-results 200"
    echo "   2. Classifier: $DOCKER_COMPOSE run --rm backend ./classifier -classify-all -limit 50"
    exit 1
fi

echo "✅ Найдено $CLASSIFIED_COUNT классифицированных кандидатов"
echo ""

# Проверяем service_provider
SERVICE_PROVIDER_COUNT=$($DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -t -A -c "
SELECT COUNT(*) 
FROM potential_shops 
WHERE status = 'classified' 
AND metadata->>'site_type' = 'service_provider';
" 2>/dev/null | tr -d ' \n')

if [ -n "$SERVICE_PROVIDER_COUNT" ] && [ "$SERVICE_PROVIDER_COUNT" -gt "0" ]; then
    echo "✅ Найдено $SERVICE_PROVIDER_COUNT service_provider (отлично для тестирования таблиц!)"
else
    echo "⚠️  Service providers не найдены, но можно протестировать на ecommerce"
fi
echo ""

# Запускаем AutoConfig на 3 кандидатах
echo "🚀 Запускаем AutoConfig на 3 кандидатах..."
echo ""

$DOCKER_COMPOSE run --rm backend ./autoconfig -limit 3

echo ""
echo "📊 Проверяем результаты..."
echo ""

# Показываем последние созданные магазины
$DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    name,
    base_url,
    CASE WHEN is_active THEN 'Yes' ELSE 'No' END as active,
    COALESCE(ai_config_model, 'N/A') as model,
    CASE WHEN selectors->>'name' IS NOT NULL THEN '✅' ELSE '❌' END as has_name,
    CASE WHEN selectors->>'price' IS NOT NULL THEN '✅' ELSE '❌' END as has_price,
    CASE 
        WHEN selectors->>'name' LIKE '%table%' OR selectors->>'name' LIKE '%tr%' OR selectors->>'name' LIKE '%td%' 
        THEN '📋 Table'
        ELSE '📦 Card'
    END as selector_type,
    created_at::timestamp as created
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC 
LIMIT 5;
" 2>/dev/null || echo "⚠️  Не удалось получить результаты"

echo ""
echo "✅ Тестирование завершено!"
echo ""
echo "💡 Проверьте логи выше на наличие:"
echo "   - 'Validation successful' с names_count > 1 для service_provider"
echo "   - 'site_type: service_provider' в логах"
echo "   - Селекторы для таблиц (table, tr, td) в результатах"

