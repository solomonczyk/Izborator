-- Откат миграции: Удаление начальной схемы базы данных

DROP TRIGGER IF EXISTS update_shops_updated_at ON shops;
DROP TRIGGER IF EXISTS update_product_prices_updated_at ON product_prices;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS product_matches;
DROP TABLE IF EXISTS raw_products;
DROP TABLE IF EXISTS product_prices;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS shops;

