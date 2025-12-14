-- Обновление селекторов для магазина shop-001
UPDATE shops 
SET selectors = '{
    "name": "h1",
    "price": ".pp-price-new, .product-price-new, [itemprop=\"price\"]",
    "image": ".pp-img-wrap img, img[itemprop=\"image\"], .product-image img",
    "description": ".pp-description, .product-description",
    "brand": ".pp-brand, [itemprop=\"brand\"]"
}'::jsonb
WHERE id = 'shop-001';

-- Проверка
SELECT id, name, selectors FROM shops WHERE id = 'shop-001';

