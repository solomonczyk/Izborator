-- Скрипт для объединения дубликатов товаров
-- Объединяет товары с одинаковым названием и брендом, перенося цены в один товар

-- Шаг 1: Находим дубликаты (товары с одинаковым названием и брендом)
-- Оставляем товар с минимальным ID как основной, остальные - дубликаты

-- Шаг 2: Переносим цены из дубликатов в основной товар
UPDATE product_prices pp
SET product_id = (
    SELECT MIN(p2.id)
    FROM products p1
    JOIN products p2 ON LOWER(TRIM(p1.name)) = LOWER(TRIM(p2.name))
        AND (p1.brand = '' OR p2.brand = '' OR LOWER(TRIM(p1.brand)) = LOWER(TRIM(p2.brand)))
    WHERE pp.product_id = p1.id
    GROUP BY LOWER(TRIM(p1.name)), 
             CASE WHEN p1.brand = '' OR p2.brand = '' THEN '' ELSE LOWER(TRIM(p1.brand)) END
)
WHERE EXISTS (
    SELECT 1
    FROM products p1
    JOIN products p2 ON LOWER(TRIM(p1.name)) = LOWER(TRIM(p2.name))
        AND (p1.brand = '' OR p2.brand = '' OR LOWER(TRIM(p1.brand)) = LOWER(TRIM(p2.brand)))
        AND p1.id != p2.id
        AND p1.id = pp.product_id
        AND p2.id < p1.id
);

-- Шаг 3: Удаляем дубликаты товаров (оставляем только основной)
DELETE FROM products
WHERE id IN (
    SELECT p1.id
    FROM products p1
    JOIN products p2 ON LOWER(TRIM(p1.name)) = LOWER(TRIM(p2.name))
        AND (p1.brand = '' OR p2.brand = '' OR LOWER(TRIM(p1.brand)) = LOWER(TRIM(p2.brand)))
        AND p1.id > p2.id
);

-- Проверка: сколько дубликатов осталось
SELECT 
    LOWER(TRIM(name)) as normalized_name,
    COUNT(*) as count
FROM products
GROUP BY LOWER(TRIM(name))
HAVING COUNT(*) > 1;

