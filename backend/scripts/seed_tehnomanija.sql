-- 1. Добавляем магазин Tehnomanija
INSERT INTO shops (id, name, code, base_url, is_active, retry_limit, retry_backoff_ms, selectors)
VALUES (
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22', -- Фиксированный UUID для удобства
    'Tehnomanija',
    'tehnomanija',
    'https://www.tehnomanija.rs',
    true,
    3,
    2000,
    '{
        "name": "h1.product-name", 
        "price": ".product-price-new",
        "image": ".product-image-gallery img",
        "description": ".product-description",
        "brand": ".product-brand" 
    }'::jsonb
)
ON CONFLICT (code) DO NOTHING;

-- Примечание: Селекторы могут потребовать уточнения, так как верстка меняется.
-- Но мы начнем с этих.

