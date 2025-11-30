-- Откат миграции scraping_stats
DROP INDEX IF EXISTS idx_scraping_stats_status;
DROP INDEX IF EXISTS idx_scraping_stats_scraped_at;
DROP INDEX IF EXISTS idx_scraping_stats_shop_id;
DROP TABLE IF EXISTS scraping_stats;

ALTER TABLE shops DROP COLUMN IF EXISTS last_scraped_at;
ALTER TABLE shops DROP COLUMN IF EXISTS scraping_enabled;

