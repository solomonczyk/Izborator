-- 0005_catalog_core.down.sql
-- Откат ядра каталога

-- Удаляем внешние ключи / колонки из существующих таблиц
ALTER TABLE product_prices
    DROP COLUMN IF EXISTS city_id;

ALTER TABLE shops
    DROP COLUMN IF EXISTS default_city_id;

ALTER TABLE products
    DROP COLUMN IF EXISTS category_id;

ALTER TABLE products
    DROP COLUMN IF EXISTS product_type_id;

-- Удаляем таблицы в правильном порядке (снизу вверх по зависимостям)
DROP TABLE IF EXISTS product_type_attributes;
DROP TABLE IF EXISTS attributes;
DROP TABLE IF EXISTS category_product_types;
DROP TABLE IF EXISTS product_types;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS cities;


