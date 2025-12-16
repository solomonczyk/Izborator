-- 0006_discovery_tables.down.sql
-- Откат миграции для таблиц обнаружения магазинов

DROP TABLE IF EXISTS shop_config_attempts;
DROP TABLE IF EXISTS potential_shops;

ALTER TABLE shops DROP COLUMN IF EXISTS is_auto_configured;
ALTER TABLE shops DROP COLUMN IF EXISTS ai_config_model;
ALTER TABLE shops DROP COLUMN IF EXISTS discovery_source;

