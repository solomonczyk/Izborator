-- Добавление конфигурации магазина Gigatron для тестирования парсинга
-- Примечание: id в таблице shops это VARCHAR(255), не UUID
INSERT INTO shops (id, name, base_url, is_active, code, selectors, rate_limit)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- Фиксированный UUID для теста (хранится как строка)
    'Gigatron',
    'https://gigatron.rs',
    true,
    'gigatron',
    '{
        "name": "h1", 
        "price": ".pp-price-new, .product-price-new", 
        "image": ".pp-img-wrap img", 
        "description": ".pp-description",
        "brand": ".pp-brand",
        "catalog_url": "https://gigatron.rs/mobilni-telefoni",
        "catalog_product_link": ".product-box a, .product-item a, .product-title a",
        "catalog_next_page": ".pagination .next, .pagination-next"
    }'::jsonb,
    2 -- 2 запросов в секунду
)
ON CONFLICT (id) DO NOTHING;

