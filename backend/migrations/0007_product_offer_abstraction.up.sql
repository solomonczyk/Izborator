-- 0007_product_offer_abstraction.up.sql
-- Pivot: Расширение модели Product для поддержки товаров и услуг
-- Дата: 2025-01-XX
-- Описание: Добавляем поля для универсальной модели Offer (товары + услуги)

------------------------------------------------------------
-- 1. Добавление типа продукта (good | service)
------------------------------------------------------------
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS type VARCHAR(20) DEFAULT 'good' 
    CHECK (type IN ('good', 'service'));

CREATE INDEX IF NOT EXISTS idx_products_type ON products(type);

------------------------------------------------------------
-- 2. Метаданные для услуг (JSONB)
-- Содержит: duration, master_name, service_area
------------------------------------------------------------
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS service_metadata JSONB;

-- Индекс для поиска по метаданным услуг
CREATE INDEX IF NOT EXISTS idx_products_service_metadata 
    ON products USING GIN (service_metadata)
    WHERE type = 'service';

------------------------------------------------------------
-- 3. Флаги логистики
------------------------------------------------------------
-- is_deliverable: товар можно доставить (для товаров)
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS is_deliverable BOOLEAN DEFAULT TRUE;

-- is_onsite: услуга с выездом мастера (для услуг)
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS is_onsite BOOLEAN DEFAULT FALSE;

-- Индексы для фильтрации по логистике
CREATE INDEX IF NOT EXISTS idx_products_is_deliverable 
    ON products(is_deliverable) 
    WHERE type = 'good';

CREATE INDEX IF NOT EXISTS idx_products_is_onsite 
    ON products(is_onsite) 
    WHERE type = 'service';

------------------------------------------------------------
-- 4. Обновление существующих записей
-- Все существующие продукты - это товары (good)
------------------------------------------------------------
UPDATE products 
SET type = 'good', 
    is_deliverable = TRUE,
    is_onsite = FALSE
WHERE type IS NULL OR type = '';

------------------------------------------------------------
-- 5. Комментарии к полям
------------------------------------------------------------
COMMENT ON COLUMN products.type IS 'Тип предложения: good (товар) или service (услуга)';
COMMENT ON COLUMN products.service_metadata IS 'Метаданные для услуг: duration (длительность), master_name (имя мастера), service_area (район обслуживания)';
COMMENT ON COLUMN products.is_deliverable IS 'Товар можно доставить (для товаров)';
COMMENT ON COLUMN products.is_onsite IS 'Услуга с выездом мастера (для услуг)';

