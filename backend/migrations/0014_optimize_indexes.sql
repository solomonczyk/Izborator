-- Миграция для оптимизации индексов и производительности запросов

-- Индексы для поиска товаров (оптимизация searchViaPostgres)
CREATE INDEX IF NOT EXISTS idx_products_name_search ON products USING gin(to_tsvector('serbian', name));
CREATE INDEX IF NOT EXISTS idx_products_brand_search ON products USING gin(to_tsvector('serbian', brand));
CREATE INDEX IF NOT EXISTS idx_products_description_search ON products USING gin(to_tsvector('serbian', COALESCE(description, '')));

-- Композитный индекс для поиска по нескольким полям
CREATE INDEX IF NOT EXISTS idx_products_search_fields ON products (name, brand, category) WHERE name IS NOT NULL;

-- Индексы для фильтрации по категориям
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products (category_id) WHERE category_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category) WHERE category IS NOT NULL;

-- Индексы для цен (оптимизация GetProductPrices)
CREATE INDEX IF NOT EXISTS idx_product_prices_product_id_shop_id ON product_prices (product_id, shop_id);
CREATE INDEX IF NOT EXISTS idx_product_prices_price ON product_prices (price) WHERE price IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_product_prices_updated_at ON product_prices (updated_at DESC);

-- Индекс для GetURLsForRescrape (оптимизация запроса с DISTINCT ON)
CREATE INDEX IF NOT EXISTS idx_product_prices_rescrape ON product_prices (url, shop_id, updated_at) 
WHERE url IS NOT NULL AND updated_at IS NOT NULL;

-- Индексы для browse запросов
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products (updated_at DESC);

-- Индексы для matching (поиск похожих товаров)
-- Функциональные индексы для LOWER(TRIM()) операций
CREATE INDEX IF NOT EXISTS idx_products_name_lower_trim ON products (LOWER(TRIM(name)));
CREATE INDEX IF NOT EXISTS idx_products_brand_lower_trim ON products (LOWER(TRIM(brand))) WHERE brand IS NOT NULL AND brand != '';

-- Индекс для product_matches
CREATE INDEX IF NOT EXISTS idx_product_matches_product_id ON product_matches (product_id);
CREATE INDEX IF NOT EXISTS idx_product_matches_similarity ON product_matches (product_id, similarity DESC);

-- Индексы для categories
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories (parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_categories_slug_active ON categories (slug) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_categories_active_sort ON categories (is_active, sort_order, name_sr) WHERE is_active = true;

-- Индексы для cities
CREATE INDEX IF NOT EXISTS idx_cities_slug_active ON cities (slug) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_cities_active_sort ON cities (is_active, sort_order, name_sr) WHERE is_active = true;

-- Комментарии к индексам
COMMENT ON INDEX idx_products_name_search IS 'Full-text search index for product names';
COMMENT ON INDEX idx_products_brand_search IS 'Full-text search index for product brands';
COMMENT ON INDEX idx_product_prices_rescrape IS 'Optimized index for rescraping outdated prices';
COMMENT ON INDEX idx_products_name_lower_trim IS 'Functional index for matching queries with LOWER(TRIM(name))';
COMMENT ON INDEX idx_product_matches_similarity IS 'Optimized index for GetMatches query';

