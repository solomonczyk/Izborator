#!/bin/bash
# Скрипт для создания таблицы shop_config_attempts вручную

echo "🔧 Создание таблицы shop_config_attempts"
echo "========================================"
echo ""

cd ~/Izborator

echo "📊 Шаг 1: Проверка текущего состояния..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts" > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "✅ Таблица shop_config_attempts уже существует"
  exit 0
fi

echo "❌ Таблица shop_config_attempts не существует"
echo ""

echo "🔧 Шаг 2: Создание таблицы shop_config_attempts..."
docker-compose exec -T postgres psql -U postgres -d izborator << 'SQL'
-- Создание таблицы shop_config_attempts
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

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_potential_shop ON shop_config_attempts(potential_shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_shop ON shop_config_attempts(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_config_attempts_status ON shop_config_attempts(status);

-- Проверка
SELECT 'Таблица shop_config_attempts создана успешно!' as result;
SQL

if [ $? -eq 0 ]; then
  echo ""
  echo "✅ Таблица shop_config_attempts создана!"
  echo ""
  echo "📊 Шаг 3: Проверка структуры таблицы..."
  docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts"
  echo ""
  echo "✅ Готово! Теперь можно запускать Classifier и AutoConfig"
else
  echo ""
  echo "❌ Ошибка при создании таблицы"
  exit 1
fi

