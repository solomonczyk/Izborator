-- Обновление селекторов для Gigatron на основе реальной структуры страницы
-- Попробуем более общие селекторы, которые должны работать

UPDATE shops 
SET selectors = '{
    "name": "h1",
    "price": ".product-price, .price, [class*=\"price\"], .pp-price-new",
    "image": "img[src*=\"product\"], .product-image img, .pp-img-wrap img",
    "description": ".product-description, .description, .pp-description",
    "brand": ".product-brand, .brand, .pp-brand, [itemprop=\"brand\"]"
}'::jsonb
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

