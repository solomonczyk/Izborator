-- 0006_discovery_tables.up.sql
-- Таблицы для автоматического обнаружения и классификации магазинов (Project Horizon)

------------------------------------------------------------
-- 1. Таблица potential_shops — кандидаты на магазины
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS potential_shops (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain          VARCHAR(255) UNIQUE NOT NULL,        -- например, "tehnomanija.rs"
    source          VARCHAR(50),                          -- "google_search", "manual", "registry"
    status          VARCHAR(20) DEFAULT 'new',           -- new, classified, configured, rejected, active
    confidence_score FLOAT DEFAULT 0,                    -- 0.0 - 1.0 (насколько уверены, что это магазин)
    discovered_at   TIMESTAMPTZ DEFAULT NOW(),
    classified_at   TIMESTAMPTZ NULL,
    metadata        JSONB,                                -- Заголовки, meta tags, контакты
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_potential_shops_status ON potential_shops(status);
CREATE INDEX IF NOT EXISTS idx_potential_shops_domain ON potential_shops(domain);
CREATE INDEX IF NOT EXISTS idx_potential_shops_confidence ON potential_shops(confidence_score DESC);

------------------------------------------------------------
-- 2. Расширение основной таблицы shops
------------------------------------------------------------
ALTER TABLE shops ADD COLUMN IF NOT EXISTS is_auto_configured BOOLEAN DEFAULT FALSE;
ALTER TABLE shops ADD COLUMN IF NOT EXISTS ai_config_model VARCHAR(50); -- версия модели, которая создала конфиг
ALTER TABLE shops ADD COLUMN IF NOT EXISTS discovery_source VARCHAR(50); -- откуда найден магазин

------------------------------------------------------------
-- 3. Таблица для логов AI-конфигурации
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_config_attempts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    potential_shop_id UUID REFERENCES potential_shops(id) ON DELETE SET NULL,
    shop_id         VARCHAR(255) REFERENCES shops(id) ON DELETE SET NULL,
    html_sample     TEXT,                                -- Очищенный HTML для анализа
    ai_response     JSONB,                               -- Ответ LLM
    validation_result JSONB,                            -- Результат проверки селекторов
    status          VARCHAR(20),                         -- success, failed, pending
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_potential_shop ON shop_config_attempts(potential_shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_shop ON shop_config_attempts(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_status ON shop_config_attempts(status);

------------------------------------------------------------
-- 4. Триггер обновления updated_at для potential_shops
------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'set_potential_shops_updated_at'
    ) THEN
        CREATE TRIGGER set_potential_shops_updated_at
        BEFORE UPDATE ON potential_shops
        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
    END IF;
END;
$$;

