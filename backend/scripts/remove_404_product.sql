-- Удаление ошибочного товара "Greška 404"
-- Этот товар был спарсен со страницы 404 и не должен быть в каталоге

-- Сначала удаляем связанные цены
DELETE FROM product_prices 
WHERE product_id IN (
    SELECT id FROM products 
    WHERE name LIKE '%404%' OR name LIKE '%Greška%'
);

-- Затем удаляем сам товар
DELETE FROM products 
WHERE name LIKE '%404%' OR name LIKE '%Greška%';

-- Проверяем результат
SELECT COUNT(*) as deleted_count 
FROM products 
WHERE name LIKE '%404%' OR name LIKE '%Greška%';

