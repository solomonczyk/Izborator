#!/bin/bash
# Скрипт для проверки результатов harvest на сервере

echo "📊 Проверка результатов Harvest (Project Horizon)"
echo "=================================================="
echo ""

echo "🛍️  Созданные магазины (AutoConfig):"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    name, 
    base_url, 
    is_active,
    is_auto_configured,
    ai_config_model,
    selectors->>'name' as name_selector,
    selectors->>'price' as price_selector,
    created_at
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC
LIMIT 10;
"

echo ""
echo "📈 Статистика по статусам potential_shops:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count
FROM potential_shops
GROUP BY status
ORDER BY status;
"

echo ""
echo "🤖 Попытки конфигурации:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count,
    MAX(created_at) as last_attempt
FROM shop_config_attempts
GROUP BY status
ORDER BY status;
" 2>/dev/null || echo "⚠️  Таблица shop_config_attempts не существует (нужно применить миграции)"

