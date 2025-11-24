-- Обновление селектора цены для Gigatron
-- Пробуем разные варианты селекторов для цены

UPDATE shops 
SET selectors = jsonb_set(
    selectors, 
    '{price}', 
    '"span:contains(\"RSD\"), .price, [class*=\"price\"], .product-price, .pp-price-new, .product-price-new, div:contains(\"RSD\")"'
)
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

