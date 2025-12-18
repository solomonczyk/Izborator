#!/bin/bash
# Скрипт для проверки результатов полного конвейера Project Horizon

echo "📊 Проверка результатов Project Horizon Pipeline"
echo "================================================"
echo ""

cd ~/Izborator

echo "🔍 1. Статус миграций:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT version, dirty FROM schema_migrations;
" 2>/dev/null || echo "⚠️  Не удалось получить статус"
echo ""

echo "✅ 2. Проверка таблицы shop_config_attempts:"
if docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts" > /dev/null 2>&1; then
  echo "✅ Таблица shop_config_attempts существует"
else
  echo "❌ Таблица shop_config_attempts не существует"
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
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT 
      name, 
      base_url, 
      is_active,
      is_auto_configured,
      ai_config_model,
      name_selector,
      price_selector,
      created_at
  FROM shops 
  WHERE is_auto_configured = true 
  ORDER BY created_at DESC
  LIMIT 10;
" 2>/dev/null || echo "⚠️  Не удалось получить результаты"
echo ""

echo "🤖 5. Попытки конфигурации (последние 5):"
if docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts" > /dev/null 2>&1; then
  docker-compose exec -T postgres psql -U postgres -d izborator -c "
    SELECT 
        id,
        shop_id,
        status,
        error_message,
        created_at
    FROM shop_config_attempts
    ORDER BY created_at DESC
    LIMIT 5;
  " 2>/dev/null || echo "⚠️  Не удалось получить попытки"
else
  echo "⚠️  Таблица shop_config_attempts не существует"
fi
echo ""

echo "✅ Проверка завершена!"

