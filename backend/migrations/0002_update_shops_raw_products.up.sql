-- Миграция: Обновление структуры shops и raw_products
-- Дата: 2025-11-25
-- Описание: Приведение таблиц к финальной структуре для scraper + processor

-- Обновление таблицы shops
-- Добавляем поле code, если его нет
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shops' AND column_name = 'code'
    ) THEN
        ALTER TABLE shops ADD COLUMN code TEXT;
        -- Генерируем code из name для существующих записей
        UPDATE shops SET code = LOWER(REGEXP_REPLACE(name, '[^a-zA-Z0-9]', '', 'g'))
        WHERE code IS NULL;
        -- Делаем code обязательным и уникальным
        ALTER TABLE shops ALTER COLUMN code SET NOT NULL;
        ALTER TABLE shops ADD CONSTRAINT shops_code_unique UNIQUE (code);
    END IF;
END $$;

-- Переименовываем enabled в is_active, если нужно
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shops' AND column_name = 'enabled'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shops' AND column_name = 'is_active'
    ) THEN
        ALTER TABLE shops RENAME COLUMN enabled TO is_active;
    END IF;
END $$;

-- Обновление таблицы raw_products
-- Добавляем поле processed_at, если его нет
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'processed_at'
    ) THEN
        ALTER TABLE raw_products ADD COLUMN processed_at TIMESTAMPTZ;
    END IF;
END $$;

-- Добавляем поле raw_payload для полного сырого объекта
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'raw_payload'
    ) THEN
        ALTER TABLE raw_products ADD COLUMN raw_payload JSONB;
    END IF;
END $$;

-- Переименовываем specs в specs_json для ясности (если нужно)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'specs'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'specs_json'
    ) THEN
        ALTER TABLE raw_products RENAME COLUMN specs TO specs_json;
    END IF;
END $$;

-- Переименовываем scraped_at в parsed_at для единообразия
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'scraped_at'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'parsed_at'
    ) THEN
        ALTER TABLE raw_products RENAME COLUMN scraped_at TO parsed_at;
    END IF;
END $$;

-- Изменяем тип parsed_at на TIMESTAMPTZ, если нужно
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' 
        AND column_name = 'parsed_at' 
        AND data_type != 'timestamp with time zone'
    ) THEN
        ALTER TABLE raw_products ALTER COLUMN parsed_at TYPE TIMESTAMPTZ USING parsed_at::TIMESTAMPTZ;
    END IF;
END $$;

-- Удаляем старый первичный ключ, если он на id
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'raw_products' 
        AND constraint_type = 'PRIMARY KEY'
        AND constraint_name = 'raw_products_pkey'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'raw_products' AND column_name = 'id'
    ) THEN
        -- Удаляем старый PK
        ALTER TABLE raw_products DROP CONSTRAINT raw_products_pkey;
        -- Удаляем колонку id, если она есть
        ALTER TABLE raw_products DROP COLUMN IF EXISTS id;
    END IF;
END $$;

-- Создаём составной первичный ключ на (shop_id, external_id), если его нет
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'raw_products' 
        AND constraint_type = 'PRIMARY KEY'
        AND constraint_name LIKE '%shop_id%external_id%'
    ) THEN
        -- Убеждаемся, что external_id NOT NULL
        ALTER TABLE raw_products ALTER COLUMN external_id SET NOT NULL;
        -- Создаём составной PK
        ALTER TABLE raw_products ADD PRIMARY KEY (shop_id, external_id);
    END IF;
END $$;

-- Удаляем старый индекс, если он есть
DROP INDEX IF EXISTS idx_raw_products_external_id;

-- Создаём индекс для выборки необработанных сырых товаров
CREATE INDEX IF NOT EXISTS idx_raw_products_unprocessed
    ON raw_products (processed, shop_id, parsed_at DESC)
    WHERE processed = false;

-- Удаляем shop_name из raw_products, если он есть (избыточно, т.к. есть shop_id)
-- Но оставляем для обратной совместимости, если данные уже есть
-- DO $$
-- BEGIN
--     IF EXISTS (
--         SELECT 1 FROM information_schema.columns 
--         WHERE table_name = 'raw_products' AND column_name = 'shop_name'
--     ) THEN
--         ALTER TABLE raw_products DROP COLUMN shop_name;
--     END IF;
-- END $$;

