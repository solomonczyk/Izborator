-- Откат миграции оптимизации запросов

DROP INDEX IF EXISTS idx_raw_products_processed_parsed_at;
DROP INDEX IF EXISTS idx_product_prices_city;
DROP INDEX IF EXISTS idx_product_prices_no_city;
DROP INDEX IF EXISTS idx_products_category_brand;
DROP INDEX IF EXISTS idx_product_prices_shop;

