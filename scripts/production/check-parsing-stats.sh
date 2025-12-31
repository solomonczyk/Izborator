#!/bin/bash
# Проверка статистики парсинга

cd ~/Izborator

MACOLA_ID="ef60e5e1-8624-4dc0-9de7-656ba3efa482"
ALATNIK_ID="102179a0-d568-4f71-9351-da2474620fe9"

echo "=== Статистика по магазинам ==="
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    s.name,
    s.base_url,
    COUNT(DISTINCT rp.url) as total_urls_parsed,
    COUNT(DISTINCT CASE WHEN rp.name IS NOT NULL AND rp.name != '' AND rp.price > 0 THEN rp.url END) as successful,
    COUNT(DISTINCT CASE WHEN rp.name IS NULL OR rp.name = '' OR rp.price IS NULL OR rp.price = 0 THEN rp.url END) as failed
FROM shops s
LEFT JOIN raw_products rp ON rp.shop_id = s.id
WHERE s.id IN ('$MACOLA_ID', '$ALATNIK_ID')
GROUP BY s.name, s.base_url;
SQL

echo ""
echo "=== Примеры успешных парсингов ==="
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    url,
    name,
    price,
    currency
FROM raw_products
WHERE shop_id IN ('$MACOLA_ID', '$ALATNIK_ID')
AND name IS NOT NULL 
AND name != ''
AND price > 0
ORDER BY parsed_at DESC
LIMIT 5;
SQL

echo ""
echo "=== Примеры неудачных парсингов (пустые name или price) ==="
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    url,
    COALESCE(name, 'NULL') as name,
    COALESCE(price::text, 'NULL') as price
FROM raw_products
WHERE shop_id IN ('$MACOLA_ID', '$ALATNIK_ID')
AND (name IS NULL OR name = '' OR price IS NULL OR price = 0)
ORDER BY parsed_at DESC
LIMIT 5;
SQL

echo ""
echo "=== Всего URL в каталогах (из scraping_stats) ==="
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    shop_name,
    COUNT(*) as total_attempts,
    SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
    SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END) as errors
FROM scraping_stats
WHERE shop_id IN ('$MACOLA_ID', '$ALATNIK_ID')
GROUP BY shop_name;
SQL

