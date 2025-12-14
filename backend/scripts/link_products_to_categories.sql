-- Привязка товаров к категориям и городам

-- 1. Motorola -> Mobilni telefoni (mobilni-telefoni)
UPDATE products 
SET category_id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'
WHERE name ILIKE '%MOTOROLA%' OR name ILIKE '%motorola%';

-- 2. HP ZBook -> Laptopovi (laptopovi)
UPDATE products 
SET category_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc'
WHERE name ILIKE '%HP%ZBook%' OR name ILIKE '%laptop%';

-- 3. TCL -> Televizori (пока привяжем к Elektronika, т.к. категории Televizori нет)
UPDATE products 
SET category_id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'
WHERE name ILIKE '%TCL%' OR name ILIKE '%televizor%';

-- 4. Привязываем цены к городу Beograd (для всех товаров без city_id)
UPDATE product_prices
SET city_id = '11111111-1111-1111-1111-111111111111'
WHERE city_id IS NULL;

-- Проверка результата
SELECT 
    p.id,
    p.name,
    c.name_sr as category_name,
    pp.price,
    ci.name_sr as city_name
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN product_prices pp ON p.id = pp.product_id
LEFT JOIN cities ci ON pp.city_id = ci.id
ORDER BY p.name;

