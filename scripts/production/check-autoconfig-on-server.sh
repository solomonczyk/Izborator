#!/bin/bash
# Скрипт для проверки данных AutoConfig на продакшен сервере
# Запускать на сервере: ssh root@152.53.227.37

set -e

echo "🔍 Проверка данных для тестирования AutoConfig"
echo "=============================================="
echo ""

# Переходим в директорию проекта
cd ~/Izborator 2>/dev/null || cd /root/Izborator 2>/dev/null || {
    echo "❌ Не удалось найти директорию проекта"
    echo "   Убедитесь, что вы находитесь в директории Izborator"
    exit 1
}

# Проверяем, что docker-compose доступен
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose не найден"
    exit 1
fi

# Используем docker compose (новый синтаксис) или docker-compose (старый)
DOCKER_COMPOSE="docker compose"
if ! docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
fi

echo "1️⃣ Статистика по potential_shops:"
$DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count,
    COUNT(*) FILTER (WHERE metadata->>'site_type' = 'service_provider') as service_providers,
    COUNT(*) FILTER (WHERE metadata->>'site_type' = 'ecommerce') as ecommerce,
    ROUND(MAX(confidence_score)::numeric, 2) as max_score,
    ROUND(AVG(confidence_score)::numeric, 2) as avg_score
FROM potential_shops
GROUP BY status
ORDER BY status;
" 2>/dev/null || echo "⚠️  Не удалось получить статистику"

echo ""
echo "2️⃣ Классифицированные кандидаты (готовы для AutoConfig):"
CLASSIFIED_COUNT=$($DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -t -A -c "
SELECT COUNT(*) 
FROM potential_shops 
WHERE status = 'classified';
" 2>/dev/null | tr -d ' \n')

if [ -n "$CLASSIFIED_COUNT" ] && [ "$CLASSIFIED_COUNT" -gt "0" ]; then
    echo "✅ Найдено $CLASSIFIED_COUNT классифицированных кандидатов"
    echo ""
    
    # Service providers
    SERVICE_PROVIDER_COUNT=$($DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -t -A -c "
    SELECT COUNT(*) 
    FROM potential_shops 
    WHERE status = 'classified' 
    AND metadata->>'site_type' = 'service_provider';
    " 2>/dev/null | tr -d ' \n')
    
    if [ -n "$SERVICE_PROVIDER_COUNT" ] && [ "$SERVICE_PROVIDER_COUNT" -gt "0" ]; then
        echo "✅ Из них service_provider: $SERVICE_PROVIDER_COUNT (отлично для тестирования таблиц!)"
    else
        echo "⚠️  Service providers не найдены. Нужно запустить Discovery для поиска услуг."
    fi
    echo ""
    
    # Примеры кандидатов
    echo "   Примеры кандидатов:"
    $DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -c "
    SELECT 
        domain,
        status,
        ROUND(confidence_score::numeric, 2) as score,
        COALESCE(metadata->>'site_type', 'N/A') as site_type,
        discovered_at::date as discovered
    FROM potential_shops 
    WHERE status = 'classified'
    ORDER BY confidence_score DESC, discovered_at DESC
    LIMIT 5;
    " 2>/dev/null || echo "⚠️  Не удалось получить примеры"
else
    echo "❌ Классифицированных кандидатов нет (0)"
    echo "   Нужно запустить:"
    echo "   1. Discovery (поиск кандидатов)"
    echo "   2. Classifier (классификация)"
fi

echo ""
echo "3️⃣ Магазины, созданные через AutoConfig:"
AUTOCONFIG_COUNT=$($DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -t -A -c "
SELECT COUNT(*) 
FROM shops 
WHERE is_auto_configured = true;
" 2>/dev/null | tr -d ' \n')

if [ -n "$AUTOCONFIG_COUNT" ] && [ "$AUTOCONFIG_COUNT" -gt "0" ]; then
    echo "✅ Найдено $AUTOCONFIG_COUNT автоматически созданных магазинов"
    echo ""
    echo "   Последние созданные:"
    $DOCKER_COMPOSE exec -T postgres psql -U postgres -d izborator -c "
    SELECT 
        name,
        base_url,
        CASE WHEN is_active THEN 'Yes' ELSE 'No' END as active,
        COALESCE(ai_config_model, 'N/A') as model,
        CASE WHEN selectors->>'name' IS NOT NULL THEN '✅' ELSE '❌' END as has_name,
        CASE WHEN selectors->>'price' IS NOT NULL THEN '✅' ELSE '❌' END as has_price,
        created_at::date as created
    FROM shops 
    WHERE is_auto_configured = true 
    ORDER BY created_at DESC 
    LIMIT 5;
    " 2>/dev/null || echo "⚠️  Не удалось получить данные"
else
    echo "⚠️  Автоматически созданных магазинов нет"
fi

echo ""
echo "4️⃣ Рекомендации для тестирования:"
echo ""

if [ -z "$CLASSIFIED_COUNT" ] || [ "$CLASSIFIED_COUNT" -eq "0" ]; then
    echo "📋 Для получения данных выполните:"
    echo ""
    echo "   1. Запустить Discovery для поиска сайтов услуг:"
    echo "      $DOCKER_COMPOSE run --rm backend ./discovery -max-results 200"
    echo ""
    echo "   2. Запустить Classifier для классификации:"
    echo "      $DOCKER_COMPOSE run --rm backend ./classifier -classify-all -limit 50"
    echo ""
    echo "   3. Запустить AutoConfig для тестирования:"
    echo "      $DOCKER_COMPOSE run --rm backend ./autoconfig -limit 5"
elif [ -n "$SERVICE_PROVIDER_COUNT" ] && [ "$SERVICE_PROVIDER_COUNT" -gt "0" ]; then
    echo "✅ Отлично! Есть данные для тестирования табличных данных"
    echo ""
    echo "   Запустите AutoConfig:"
    echo "   $DOCKER_COMPOSE run --rm backend ./autoconfig -limit 3"
else
    echo "⚠️  Есть классифицированные кандидаты, но нет service_provider"
    echo ""
    echo "   Для получения service_provider запустите Discovery:"
    echo "   $DOCKER_COMPOSE run --rm backend ./discovery -max-results 200"
fi

echo ""
echo "✅ Проверка завершена!"

