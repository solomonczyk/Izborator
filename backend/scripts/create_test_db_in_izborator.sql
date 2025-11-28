-- Создание схемы в базе данных izborator
-- Выполните: docker exec -i izborator_postgres psql -U postgres -d izborator < backend/scripts/create_test_db_in_izborator.sql

-- Создаём таблицу shops
CREATE TABLE IF NOT EXISTS shops (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT        NOT NULL,
    code            TEXT        NOT NULL UNIQUE,
    base_url        TEXT        NOT NULL,
    selectors       JSONB,
    rate_limit      INTEGER     DEFAULT 10,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создаём таблицу raw_products
CREATE TABLE IF NOT EXISTS raw_products (
    shop_id         UUID        NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    external_id     TEXT        NOT NULL,
    url             TEXT        NOT NULL,
    name            TEXT        NOT NULL,
    description     TEXT,
    brand           TEXT,
    category        TEXT,
    price           NUMERIC(18, 2),
    currency        TEXT        DEFAULT 'RSD',
    image_urls      JSONB,
    specs_json      JSONB,
    raw_payload     JSONB,
    parsed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed       BOOLEAN     NOT NULL DEFAULT FALSE,
    processed_at    TIMESTAMPTZ,
    in_stock        BOOLEAN     DEFAULT TRUE,
    shop_name       TEXT,
    PRIMARY KEY (shop_id, external_id)
);

-- Создаём таблицу products
CREATE TABLE IF NOT EXISTS products (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT        NOT NULL,
    description     TEXT,
    brand           TEXT,
    category        TEXT,
    image_url       TEXT,
    specs           JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создаём таблицу product_prices
CREATE TABLE IF NOT EXISTS product_prices (
    product_id      UUID        NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    shop_id         UUID        NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    shop_name       TEXT        NOT NULL,
    price           NUMERIC(18, 2) NOT NULL,
    currency        TEXT        NOT NULL DEFAULT 'RSD',
    url             TEXT        NOT NULL,
    in_stock        BOOLEAN     DEFAULT TRUE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (product_id, shop_id)
);

-- Создаём индексы
CREATE INDEX IF NOT EXISTS idx_raw_products_unprocessed
    ON raw_products (processed, shop_id, parsed_at DESC)
    WHERE processed = FALSE;

CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_product_prices_product_id ON product_prices(product_id);
CREATE INDEX IF NOT EXISTS idx_product_prices_shop_id ON product_prices(shop_id);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_shops_updated_at ON shops;
CREATE TRIGGER update_shops_updated_at BEFORE UPDATE ON shops
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Вставляем тестовый магазин Gigatron
INSERT INTO shops (id, name, code, base_url, selectors, rate_limit, is_active)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000'::UUID,
    'Gigatron',
    'gigatron',
    'https://gigatron.rs',
    '{
        "name": "h1.product-title, .product-title",
        "price": ".price, .product-price, [data-price]",
        "image": "img.product-image, .product-image img",
        "description": ".product-description, .description",
        "category": ".breadcrumb a:last-child, .category",
        "brand": ".brand, [data-brand]"
    }'::JSONB,
    5,
    TRUE
)
ON CONFLICT (code) DO UPDATE SET
    selectors = EXCLUDED.selectors,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Выводим результат
SELECT 'База данных izborator создана и настроена!' AS status;
SELECT COUNT(*) AS shops_count FROM shops;
SELECT COUNT(*) AS products_count FROM products;

