-- Миграция для оптимизации SQL запросов и добавления недостающих индексов

-- Индекс для GetUnprocessedRawProducts (оптимизация запроса с WHERE processed = FALSE)
CREATE INDEX IF NOT EXISTS idx_raw_products_processed_parsed_at ON raw_products (processed, parsed_at) 
WHERE processed = FALSE;

-- Композитный индекс для product_prices с city_id (оптимизация GetProductPricesByCity)
CREATE INDEX IF NOT EXISTS idx_product_prices_city ON product_prices (product_id, city_id, price, updated_at DESC) 
WHERE city_id IS NOT NULL;

-- Индекс для product_prices без city_id (для запросов с OR city_id IS NULL)
CREATE INDEX IF NOT EXISTS idx_product_prices_no_city ON product_prices (product_id, price, updated_at DESC) 
WHERE city_id IS NULL;

-- Индекс для browse запросов с фильтрами
CREATE INDEX IF NOT EXISTS idx_products_category_brand ON products (category_id, brand) 
WHERE category_id IS NOT NULL;

-- Индекс для поиска по shop_id в browse
CREATE INDEX IF NOT EXISTS idx_product_prices_shop ON product_prices (shop_id, product_id, price) 
WHERE price IS NOT NULL;

-- Частичный индекс для активных товаров (если есть поле is_active)
-- CREATE INDEX IF NOT EXISTS idx_products_active ON products (id, name, category_id) WHERE is_active = true;

-- Комментарии к индексам
COMMENT ON INDEX idx_raw_products_processed_parsed_at IS 'Optimized index for GetUnprocessedRawProducts query';
COMMENT ON INDEX idx_product_prices_city IS 'Optimized index for GetProductPricesByCity with city filter';
COMMENT ON INDEX idx_product_prices_no_city IS 'Optimized index for GetProductPricesByCity without city filter';
COMMENT ON INDEX idx_products_category_brand IS 'Optimized index for browse queries with category and brand filters';

