#!/bin/sh
# Скрипт для проверки результатов AutoConfig

echo "🔍 Проверка результатов AutoConfig..."
echo ""

# Проверка автоматически созданных магазинов
echo "📊 Магазины с is_auto_configured = true:"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    id,
    name, 
    base_url, 
    is_active, 
    is_auto_configured,
    created_at 
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC;
"

echo ""
echo "📈 Статистика по магазинам:"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    COUNT(*) FILTER (WHERE is_auto_configured = true) as auto_configured_count,
    COUNT(*) FILTER (WHERE is_auto_configured = false) as manual_count,
    COUNT(*) as total_shops
FROM shops;
"

echo ""
echo "🔧 Проверка shop_config_attempts:"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count
FROM shop_config_attempts
GROUP BY status
ORDER BY status;
"

echo ""
echo "✅ Проверка завершена!"

