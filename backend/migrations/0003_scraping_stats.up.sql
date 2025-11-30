-- Таблица для статистики парсинга
CREATE TABLE IF NOT EXISTS scraping_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id VARCHAR(255) NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    shop_name VARCHAR(255) NOT NULL,
    scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'error', 'partial')),
    products_found INTEGER DEFAULT 0,
    products_saved INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    error_message TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для scraping_stats
CREATE INDEX IF NOT EXISTS idx_scraping_stats_shop_id ON scraping_stats(shop_id);
CREATE INDEX IF NOT EXISTS idx_scraping_stats_scraped_at ON scraping_stats(scraped_at);
CREATE INDEX IF NOT EXISTS idx_scraping_stats_status ON scraping_stats(status);

-- Добавляем поле last_scraped_at в shops для отслеживания последнего парсинга
ALTER TABLE shops ADD COLUMN IF NOT EXISTS last_scraped_at TIMESTAMP;
ALTER TABLE shops ADD COLUMN IF NOT EXISTS scraping_enabled BOOLEAN DEFAULT true;

