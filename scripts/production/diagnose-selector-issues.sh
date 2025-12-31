#!/bin/bash
# Диагностика проблем с селекторами

set -e

cd ~/Izborator

MACOLA_ID="ef60e5e1-8624-4dc0-9de7-656ba3efa482"
ALATNIK_ID="102179a0-d568-4f71-9351-da2474620fe9"

echo "=== Селекторы для магазинов ==="
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    name,
    base_url,
    selectors->>'name' as name_selector,
    selectors->>'price' as price_selector
FROM shops
WHERE id IN ('$MACOLA_ID', '$ALATNIK_ID');
SQL

echo ""
echo "=== Статистика ошибок ==="
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    COUNT(*) as total_errors,
    COUNT(DISTINCT url) as unique_urls_with_errors
FROM raw_products
WHERE error_message LIKE '%failed to extract%';
SQL

echo ""
echo "=== Примеры URL с ошибками ==="
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    url,
    error_message
FROM raw_products
WHERE error_message LIKE '%failed to extract%'
ORDER BY created_at DESC
LIMIT 5;
SQL

echo ""
echo "=== Статистика по магазинам ==="
docker-compose exec -T postgres psql -U postgres -d izborator <<SQL
SELECT 
    s.name,
    COUNT(rp.id) as total_attempts,
    COUNT(CASE WHEN rp.error_message IS NULL THEN 1 END) as success,
    COUNT(CASE WHEN rp.error_message LIKE '%failed to extract%' THEN 1 END) as extract_errors
FROM shops s
LEFT JOIN raw_products rp ON rp.shop_id = s.id
WHERE s.id IN ('$MACOLA_ID', '$ALATNIK_ID')
GROUP BY s.name;
SQL

