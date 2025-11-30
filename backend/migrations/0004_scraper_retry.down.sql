ALTER TABLE shops
DROP COLUMN IF EXISTS retry_limit,
DROP COLUMN IF EXISTS retry_backoff_ms;

