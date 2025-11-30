-- 0005_catalog_core.up.sql
-- Ядро каталога для любых типов товаров (Serbia-first)

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

------------------------------------------------------------
-- 1. Таблица categories — иерархия категорий
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id   UUID NULL REFERENCES categories(id) ON DELETE SET NULL,
    slug        TEXT NOT NULL UNIQUE,
    code        TEXT NOT NULL UNIQUE,
    name_sr     TEXT NOT NULL,
    name_sr_lc  TEXT NOT NULL,
    level       SMALLINT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  INTEGER NOT NULL DEFAULT 100,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_categories_level ON categories(level);
CREATE INDEX IF NOT EXISTS idx_categories_name_sr_lc ON categories(name_sr_lc);

------------------------------------------------------------
-- 2. Таблица product_types — типы товаров
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product_types (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT NOT NULL UNIQUE,
    name_sr     TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

------------------------------------------------------------
-- 3. Связка category ↔ product_types (многие-ко-многим)
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS category_product_types (
    category_id     UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    product_type_id UUID NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    PRIMARY KEY (category_id, product_type_id)
);

------------------------------------------------------------
-- 4. Таблица attributes — справочник атрибутов
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attributes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            TEXT NOT NULL UNIQUE,
    name_sr         TEXT NOT NULL,
    data_type       TEXT NOT NULL,
    unit_sr         TEXT NULL,
    is_filterable   BOOLEAN NOT NULL DEFAULT TRUE,
    is_sortable     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

------------------------------------------------------------
-- 5. Связка product_type ↔ attributes
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS product_type_attributes (
    product_type_id UUID NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    attribute_id    UUID NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    is_required     BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      INTEGER NOT NULL DEFAULT 100,
    PRIMARY KEY (product_type_id, attribute_id)
);

------------------------------------------------------------
-- 6. Таблица cities — города Сербии
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT NOT NULL UNIQUE,
    name_sr     TEXT NOT NULL,
    region_sr   TEXT NULL,
    sort_order  INTEGER NOT NULL DEFAULT 100,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

------------------------------------------------------------
-- 7. Изменения в существующих таблицах
------------------------------------------------------------
-- products: добавляем category_id и product_type_id
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS category_id UUID NULL REFERENCES categories(id);

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS product_type_id UUID NULL REFERENCES product_types(id);

CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_product_type_id ON products(product_type_id);

-- shops: дефолтный город
ALTER TABLE shops
    ADD COLUMN IF NOT EXISTS default_city_id UUID NULL REFERENCES cities(id);

CREATE INDEX IF NOT EXISTS idx_shops_default_city_id ON shops(default_city_id);

-- product_prices: привязка к городу
ALTER TABLE product_prices
    ADD COLUMN IF NOT EXISTS city_id UUID NULL REFERENCES cities(id);

CREATE INDEX IF NOT EXISTS idx_product_prices_city_id ON product_prices(city_id);
CREATE INDEX IF NOT EXISTS idx_product_prices_product_city ON product_prices(product_id, city_id);

------------------------------------------------------------
-- 8. Триггеры обновления updated_at
------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_categories_updated_at'
    ) THEN
        CREATE TRIGGER set_categories_updated_at
        BEFORE UPDATE ON categories
        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_product_types_updated_at'
    ) THEN
        CREATE TRIGGER set_product_types_updated_at
        BEFORE UPDATE ON product_types
        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_attributes_updated_at'
    ) THEN
        CREATE TRIGGER set_attributes_updated_at
        BEFORE UPDATE ON attributes
        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_cities_updated_at'
    ) THEN
        CREATE TRIGGER set_cities_updated_at
        BEFORE UPDATE ON cities
        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
    END IF;
END;
$$;


