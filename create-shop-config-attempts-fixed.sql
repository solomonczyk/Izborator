-- Исправленная версия создания таблицы shop_config_attempts
-- shop_id должен быть VARCHAR(255), так как shops.id имеет тип VARCHAR(255)

CREATE TABLE IF NOT EXISTS shop_config_attempts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    potential_shop_id UUID REFERENCES potential_shops(id) ON DELETE SET NULL,
    shop_id         VARCHAR(255) REFERENCES shops(id) ON DELETE SET NULL,
    html_sample     TEXT,
    ai_response     JSONB,
    validation_result JSONB,
    status          VARCHAR(20),
    error_message   TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_potential_shop ON shop_config_attempts(potential_shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_shop ON shop_config_attempts(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_status ON shop_config_attempts(status);

