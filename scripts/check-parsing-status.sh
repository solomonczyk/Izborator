#!/bin/bash
# Скрипт для проверки статуса парсинга магазинов

echo "🔍 Проверка статуса парсинга..."
echo ""

# Подключаемся к базе данных через Docker
docker-compose exec -T postgres psql -U postgres -d izborator <<EOF

-- Статистика по магазинам
SELECT 
    '📊 Статистика магазинов' as info;
    
SELECT 
    COUNT(*) as total_shops,
    COUNT(*) FILTER (WHERE is_active = true) as enabled_shops,
    COUNT(*) FILTER (WHERE is_active = false) as disabled_shops,
    COUNT(*) FILTER (WHERE selectors IS NOT NULL AND selectors != '{}'::jsonb) as configured_shops,
    COUNT(*) FILTER (WHERE is_active = true AND (selectors IS NULL OR selectors = '{}'::jsonb)) as enabled_but_not_configured
FROM shops;

-- Список активных магазинов с селекторами
SELECT 
    '✅ Активные магазины с селекторами:' as info;
    
SELECT 
    name,
    base_url,
    CASE 
        WHEN selectors IS NULL OR selectors = '{}'::jsonb THEN '❌ Нет селекторов'
        ELSE '✅ Настроен'
    END as config_status,
    (SELECT COUNT(*) FROM raw_products WHERE shop_id = shops.id) as raw_products_count,
    (SELECT COUNT(*) FROM product_prices WHERE shop_id = shops.id) as prices_count
FROM shops
WHERE is_active = true
ORDER BY name;

-- Статистика по товарам
SELECT 
    '📦 Статистика по товарам:' as info;
    
SELECT 
    COUNT(*) as total_products,
    COUNT(*) FILTER (WHERE type = 'good') as goods_count,
    COUNT(*) FILTER (WHERE type = 'service') as services_count,
    COUNT(*) FILTER (WHERE type IS NULL OR type = '') as untyped_count
FROM products;

-- Статистика по raw_products
SELECT 
    '📋 Статистика по raw_products:' as info;
    
SELECT 
    COUNT(*) as total_raw_products,
    COUNT(*) FILTER (WHERE processed = true) as processed_count,
    COUNT(*) FILTER (WHERE processed = false) as unprocessed_count
FROM raw_products;

-- Статистика по ценам
SELECT 
    '💰 Статистика по ценам:' as info;
    
SELECT 
    COUNT(*) as total_prices,
    COUNT(DISTINCT product_id) as unique_products,
    COUNT(DISTINCT shop_id) as unique_shops,
    AVG(price) as avg_price,
    MIN(price) as min_price,
    MAX(price) as max_price
FROM product_prices;

EOF

echo ""
echo "✅ Проверка завершена!"

