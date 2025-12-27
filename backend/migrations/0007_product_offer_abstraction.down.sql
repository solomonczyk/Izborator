-- 0007_product_offer_abstraction.down.sql
-- Откат миграции расширения модели Product

-- Удаление индексов
DROP INDEX IF EXISTS idx_products_service_metadata;
DROP INDEX IF EXISTS idx_products_is_onsite;
DROP INDEX IF EXISTS idx_products_is_deliverable;
DROP INDEX IF EXISTS idx_products_type;

-- Удаление колонок
ALTER TABLE products DROP COLUMN IF EXISTS is_onsite;
ALTER TABLE products DROP COLUMN IF EXISTS is_deliverable;
ALTER TABLE products DROP COLUMN IF EXISTS service_metadata;
ALTER TABLE products DROP COLUMN IF EXISTS type;

