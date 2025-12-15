-- Проверка спарсенного товара Dell
SELECT 
    name, 
    price, 
    shop_name, 
    processed, 
    created_at 
FROM raw_products 
WHERE name LIKE '%Dell%XPS%' 
ORDER BY created_at DESC 
LIMIT 5;

-- Проверка обработанных товаров
SELECT 
    p.name, 
    pp.price, 
    pp.shop_name,
    pp.url
FROM products p
JOIN product_prices pp ON pp.product_id = p.id
WHERE p.name LIKE '%Dell%XPS%'
ORDER BY pp.updated_at DESC
LIMIT 5;


