-- Добавление конфигурации магазина Gigatron для тестирования парсинга
-- Примечание: id в таблице shops это VARCHAR(255), не UUID
INSERT INTO shops (id, name, base_url, enabled, selectors, rate_limit)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', -- Фиксированный UUID для теста (хранится как строка)
    'Gigatron',
    'https://gigatron.rs',
    true,
    '{
        "name": "h1", 
        "price": ".pp-price-new, .product-price-new", 
        "image": ".pp-img-wrap img", 
        "description": ".pp-description",
        "brand": ".pp-brand" 
    }'::jsonb,
    2 -- 2 запроса в секунду
)
ON CONFLICT (id) DO NOTHING;

