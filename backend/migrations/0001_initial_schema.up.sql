-- Миграция: Создание начальной схемы базы данных
-- Дата: 2025-01-XX

-- Таблица магазинов
CREATE TABLE IF NOT EXISTS shops (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    base_url VARCHAR(500) NOT NULL,
    selectors JSONB,
    rate_limit INTEGER DEFAULT 1,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица канонических товаров
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(500) NOT NULL,
    description TEXT,
    brand VARCHAR(255),
    category VARCHAR(255),
    image_url VARCHAR(500),
    specs JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для products
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);

-- Таблица цен товаров в магазинах
CREATE TABLE IF NOT EXISTS product_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    shop_id VARCHAR(255) NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    shop_name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'RSD',
    url VARCHAR(500),
    in_stock BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, shop_id)
);

-- Индексы для product_prices
CREATE INDEX IF NOT EXISTS idx_product_prices_product_id ON product_prices(product_id);
CREATE INDEX IF NOT EXISTS idx_product_prices_shop_id ON product_prices(shop_id);
CREATE INDEX IF NOT EXISTS idx_product_prices_updated_at ON product_prices(updated_at);

-- Таблица сырых данных парсинга
CREATE TABLE IF NOT EXISTS raw_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id VARCHAR(255) NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    shop_name VARCHAR(255) NOT NULL,
    external_id VARCHAR(255),
    name VARCHAR(500) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2),
    currency VARCHAR(10) DEFAULT 'RSD',
    url VARCHAR(500),
    image_urls JSONB,
    category VARCHAR(255),
    brand VARCHAR(255),
    specs JSONB,
    in_stock BOOLEAN DEFAULT true,
    scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для raw_products
CREATE INDEX IF NOT EXISTS idx_raw_products_shop_id ON raw_products(shop_id);
CREATE INDEX IF NOT EXISTS idx_raw_products_processed ON raw_products(processed);
CREATE INDEX IF NOT EXISTS idx_raw_products_scraped_at ON raw_products(scraped_at);
CREATE INDEX IF NOT EXISTS idx_raw_products_external_id ON raw_products(shop_id, external_id);

-- Таблица сопоставлений товаров
CREATE TABLE IF NOT EXISTS product_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    matched_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    similarity DECIMAL(3, 2) NOT NULL CHECK (similarity >= 0 AND similarity <= 1),
    confidence VARCHAR(20) DEFAULT 'medium' CHECK (confidence IN ('high', 'medium', 'low')),
    matched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, matched_id)
);

-- Индексы для product_matches
CREATE INDEX IF NOT EXISTS idx_product_matches_product_id ON product_matches(product_id);
CREATE INDEX IF NOT EXISTS idx_product_matches_matched_id ON product_matches(matched_id);
CREATE INDEX IF NOT EXISTS idx_product_matches_similarity ON product_matches(similarity);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_product_prices_updated_at BEFORE UPDATE ON product_prices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shops_updated_at BEFORE UPDATE ON shops
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

