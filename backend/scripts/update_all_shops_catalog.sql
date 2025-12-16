-- Обновление конфигурации каталогов для ВСЕХ активных магазинов
-- Этот скрипт добавляет селекторы каталога для автоматического парсинга

-- 1. Gigatron - Мобильные телефоны
UPDATE shops
SET selectors = selectors || '{
    "catalog_url": "https://gigatron.rs/mobilni-telefoni",
    "catalog_product_link": ".product-box a, .product-item a, .product-title a, a[href*=\"/mobilni-telefoni/\"]",
    "catalog_next_page": ".pagination .next, .pagination-next, a[rel=\"next\"]"
}'::jsonb
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
  AND NOT (selectors ? 'catalog_url');

-- 2. Tehnomanija - Мобильные телефоны
UPDATE shops
SET selectors = selectors || '{
    "catalog_url": "https://www.tehnomanija.rs/telefoni-smart-satovi-i-tableti/mobilni-telefoni",
    "catalog_product_link": ".product-item a, .product-card a, .product-title a, a[href*=\"/mobilni-telefoni/\"]",
    "catalog_next_page": ".pagination .next, .pagination-next, a[rel=\"next\"]"
}'::jsonb
WHERE id = 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22'
  AND NOT (selectors ? 'catalog_url');

-- 3. Проверка результата
SELECT 
    id,
    name,
    base_url,
    selectors->>'catalog_url' as catalog_url,
    CASE 
        WHEN selectors ? 'catalog_url' THEN '✅ Настроен'
        ELSE '❌ Не настроен'
    END as status
FROM shops
WHERE is_active = true
ORDER BY name;

