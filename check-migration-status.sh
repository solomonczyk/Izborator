#!/bin/bash
# Скрипт для проверки статуса миграций и таблицы shop_config_attempts

echo "🔍 Проверка статуса миграций и таблицы shop_config_attempts"
echo "============================================================"
echo ""

cd ~/Izborator 2>/dev/null || { echo "❌ Не удалось перейти в ~/Izborator"; exit 1; }

echo "📊 1. Статус миграций:"
MIGRATION_STATUS=$(docker-compose exec -T postgres psql -U postgres -d izborator -c "SELECT version, dirty FROM schema_migrations;" 2>/dev/null)
if [ $? -eq 0 ]; then
  echo "$MIGRATION_STATUS"
else
  echo "❌ Не удалось получить статус миграций"
fi
echo ""

echo "✅ 2. Проверка таблицы shop_config_attempts:"
if docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts" > /dev/null 2>&1; then
  echo "✅ Таблица shop_config_attempts СУЩЕСТВУЕТ"
  echo ""
  echo "   Структура таблицы:"
  docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts" 2>/dev/null | head -20
  TABLE_EXISTS=1
else
  echo "❌ Таблица shop_config_attempts НЕ СУЩЕСТВУЕТ"
  TABLE_EXISTS=0
fi
echo ""

echo "📊 3. Статистика potential_shops:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT 
      status,
      COUNT(*) as count
  FROM potential_shops
  GROUP BY status
  ORDER BY status;
" 2>/dev/null || echo "⚠️  Не удалось получить статистику"
echo ""

echo "🛍️  4. Созданные магазины (AutoConfig):"
AUTO_SHOP_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT COUNT(*) FROM shops WHERE is_auto_configured = true;" 2>/dev/null | tr -d ' ')
if [ -n "$AUTO_SHOP_COUNT" ]; then
  echo "   Количество: $AUTO_SHOP_COUNT"
  if [ "$AUTO_SHOP_COUNT" -gt "0" ]; then
    docker-compose exec -T postgres psql -U postgres -d izborator -c "
      SELECT 
          name, 
          base_url, 
          is_active,
          is_auto_configured,
          ai_config_model,
          created_at
      FROM shops 
      WHERE is_auto_configured = true 
      ORDER BY created_at DESC
      LIMIT 5;
    " 2>/dev/null
  fi
else
  echo "⚠️  Не удалось получить количество"
fi
echo ""

echo "🤖 5. Попытки конфигурации (последние 5):"
if [ $TABLE_EXISTS -eq 1 ]; then
  docker-compose exec -T postgres psql -U postgres -d izborator -c "
    SELECT 
        id,
        status,
        error_message,
        created_at
    FROM shop_config_attempts
    ORDER BY created_at DESC
    LIMIT 5;
  " 2>/dev/null || echo "⚠️  Не удалось получить попытки"
else
  echo "⚠️  Таблица не существует, попытки недоступны"
fi
echo ""

echo "============================================================"
if [ $TABLE_EXISTS -eq 1 ]; then
  echo "✅ Статус: Таблица shop_config_attempts существует"
else
  echo "❌ Статус: Таблица shop_config_attempts НЕ существует"
  echo ""
  echo "🔧 Что нужно сделать:"
  echo "   1. Запустить: docker-compose run --rm backend ./migrate"
  echo "   2. Или запустить workflow: Verify Migrations & Run Pipeline"
fi
echo ""

