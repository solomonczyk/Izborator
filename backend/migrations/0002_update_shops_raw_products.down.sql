-- Откат миграции: Возврат к старой структуре shops и raw_products

-- Удаляем индекс для необработанных товаров
DROP INDEX IF EXISTS idx_raw_products_unprocessed;

-- Восстанавливаем старый индекс
CREATE INDEX IF NOT EXISTS idx_raw_products_external_id ON raw_products(shop_id, external_id);

-- Откат изменений в raw_products (если нужно)
-- Восстанавливаем scraped_at
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'parsed_at'
    ) THEN
        ALTER TABLE raw_products RENAME COLUMN parsed_at TO scraped_at;
    END IF;
END $$;

-- Восстанавливаем specs
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'specs_json'
    ) THEN
        ALTER TABLE raw_products RENAME COLUMN specs_json TO specs;
    END IF;
END $$;

-- Удаляем добавленные поля
ALTER TABLE raw_products DROP COLUMN IF EXISTS raw_payload;
ALTER TABLE raw_products DROP COLUMN IF EXISTS processed_at;

-- Восстанавливаем enabled в shops
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shops' AND column_name = 'is_active'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shops' AND column_name = 'enabled'
    ) THEN
        ALTER TABLE shops RENAME COLUMN is_active TO enabled;
    END IF;
END $$;

-- Удаляем code из shops
ALTER TABLE shops DROP CONSTRAINT IF EXISTS shops_code_unique;
ALTER TABLE shops DROP COLUMN IF EXISTS code;

