#!/bin/bash
# Скрипт для проверки качества данных в автоматически созданных магазинах

set -e

cd ~/Izborator

echo "🔍 Анализ качества данных в автоматически созданных магазинах"
echo "=========================================="

# 1. Список автоматически созданных магазинов
echo ""
echo "📊 Автоматически созданные магазины:"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    id,
    name,
    code,
    base_url,
    is_active,
    is_auto_configured,
    ai_config_model,
    discovery_source,
    created_at
FROM shops
WHERE is_auto_configured = true
ORDER BY created_at DESC;
"

# 2. Проверка селекторов
echo ""
echo "📋 Селекторы для каждого магазина:"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    s.name as shop_name,
    s.selectors->>'name' as name_selector,
    s.selectors->>'price' as price_selector,
    s.selectors->>'image' as image_selector,
    s.selectors->>'description' as description_selector
FROM shops s
WHERE s.is_auto_configured = true
ORDER BY s.created_at DESC;
"

# 3. Статистика попыток конфигурации
echo ""
echo "📈 Статистика попыток конфигурации:"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '24 hours') as last_24h
FROM shop_config_attempts
GROUP BY status
ORDER BY status;
"

# 4. Проверка, есть ли уже спарсенные товары для этих магазинов
echo ""
echo "🛍️  Количество спарсенных товаров:"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    s.name as shop_name,
    COUNT(rp.id) as raw_products_count,
    COUNT(rp.id) FILTER (WHERE rp.processed = true) as processed_count,
    COUNT(rp.id) FILTER (WHERE rp.processed = false) as unprocessed_count
FROM shops s
LEFT JOIN raw_products rp ON rp.shop_id = s.id
WHERE s.is_auto_configured = true
GROUP BY s.id, s.name
ORDER BY s.created_at DESC;
"

# 5. Проверка последних спарсенных товаров (если есть)
echo ""
echo "📦 Последние спарсенные товары (если есть):"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    s.name as shop_name,
    rp.name as product_name,
    rp.price,
    rp.currency,
    rp.parsed_at
FROM shops s
JOIN raw_products rp ON rp.shop_id = s.id
WHERE s.is_auto_configured = true
ORDER BY rp.parsed_at DESC
LIMIT 10;
"

echo ""
echo "✅ Анализ завершен!"
