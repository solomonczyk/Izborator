-- Обновление конфигурации Gigatron для автоматического парсинга каталога
-- Выполни этот скрипт на сервере, чтобы включить автоматическое обнаружение товаров

UPDATE shops
SET selectors = selectors || '{
    "catalog_url": "https://gigatron.rs/mobilni-telefoni",
    "catalog_product_link": ".product-box a, .product-item a, .product-title a, a[href*=\"/mobilni-telefoni/\"]",
    "catalog_next_page": ".pagination .next, .pagination-next, a[rel=\"next\"]"
}'::jsonb
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

-- Проверка
SELECT id, name, selectors->>'catalog_url' as catalog_url 
FROM shops 
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

