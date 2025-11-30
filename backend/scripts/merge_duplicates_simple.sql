-- Простой скрипт для объединения дубликатов Motorola
-- Оставляем товар с ID: e34d9eb2-b9fc-46c5-a319-07fe52452888 (первый по алфавиту)
-- Удаляем: 9ed6a2da-75b5-4581-8c80-953f60db5e9a

-- Шаг 1: Переносим цены из дубликата в основной товар
UPDATE product_prices
SET product_id = 'e34d9eb2-b9fc-46c5-a319-07fe52452888'
WHERE product_id = '9ed6a2da-75b5-4581-8c80-953f60db5e9a'
  AND NOT EXISTS (
    SELECT 1 FROM product_prices pp2
    WHERE pp2.product_id = 'e34d9eb2-b9fc-46c5-a319-07fe52452888'
      AND pp2.shop_id = product_prices.shop_id
  );

-- Шаг 2: Удаляем дублирующиеся цены (если они остались)
DELETE FROM product_prices
WHERE product_id = '9ed6a2da-75b5-4581-8c80-953f60db5e9a';

-- Шаг 3: Удаляем дубликат товара
DELETE FROM products
WHERE id = '9ed6a2da-75b5-4581-8c80-953f60db5e9a';

-- Проверка: должно остаться только 1 товар
SELECT id, name, brand FROM products WHERE LOWER(TRIM(name)) LIKE '%motorola%';

