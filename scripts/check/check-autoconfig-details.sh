#!/bin/bash
# Комплексная проверка результатов AutoConfig

echo "🔍 Детальная проверка результатов AutoConfig"
echo "=============================================="
echo ""

cd ~/Izborator 2>/dev/null || { echo "❌ Не удалось перейти в ~/Izborator"; exit 1; }

echo "1️⃣ Статистика по potential_shops (после AutoConfig):"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count
FROM potential_shops
GROUP BY status
ORDER BY status;
" 2>/dev/null || echo "⚠️  Не удалось получить статистику"
echo ""

echo "2️⃣ Автоматически созданные магазины (детали):"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    id,
    name, 
    base_url, 
    is_active,
    is_auto_configured,
    ai_config_model,
    CASE 
        WHEN selectors->>'name' IS NOT NULL THEN '✅'
        ELSE '❌'
    END as has_name_selector,
    CASE 
        WHEN selectors->>'price' IS NOT NULL THEN '✅'
        ELSE '❌'
    END as has_price_selector,
    CASE 
        WHEN selectors->>'image' IS NOT NULL THEN '✅'
        ELSE '❌'
    END as has_image_selector,
    created_at
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC;
" 2>/dev/null || echo "⚠️  Не удалось получить данные"
echo ""

echo "3️⃣ Статистика по shop_config_attempts (попытки конфигурации):"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count,
    MAX(created_at) as last_attempt
FROM shop_config_attempts
GROUP BY status
ORDER BY status;
" 2>/dev/null || echo "⚠️  Не удалось получить статистику"
echo ""

echo "4️⃣ Последние 5 попыток конфигурации:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    potential_shop_id,
    status,
    error_message,
    created_at
FROM shop_config_attempts
ORDER BY created_at DESC
LIMIT 5;
" 2>/dev/null || echo "⚠️  Не удалось получить данные"
echo ""

echo "5️⃣ Классифицированные кандидаты (осталось для обработки):"
CLASSIFIED_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -A -c "
SELECT COUNT(*) 
FROM potential_shops 
WHERE status = 'classified';
" 2>/dev/null | tr -d ' \n')

if [ -n "$CLASSIFIED_COUNT" ] && [ "$CLASSIFIED_COUNT" -gt 0 ]; then
  echo "✅ Найдено $CLASSIFIED_COUNT классифицированных кандидатов для AutoConfig"
  echo ""
  echo "   Примеры кандидатов:"
  docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT 
      id,
      domain,
      status,
      confidence_score,
      created_at
  FROM potential_shops 
  WHERE status = 'classified'
  ORDER BY confidence_score DESC, created_at DESC
  LIMIT 5;
  " 2>/dev/null || echo "⚠️  Не удалось получить данные"
else
  echo "⚠️  Классифицированных кандидатов не найдено (0)"
fi
echo ""

echo "6️⃣ Качество конфигурации магазинов (проверка селекторов):"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    COUNT(*) FILTER (WHERE selectors->>'name' IS NOT NULL) as with_name_selector,
    COUNT(*) FILTER (WHERE selectors->>'price' IS NOT NULL) as with_price_selector,
    COUNT(*) FILTER (WHERE selectors->>'image' IS NOT NULL) as with_image_selector,
    COUNT(*) FILTER (
        WHERE selectors->>'name' IS NOT NULL 
        AND selectors->>'price' IS NOT NULL
    ) as with_both_essential,
    COUNT(*) as total_auto_configured
FROM shops 
WHERE is_auto_configured = true;
" 2>/dev/null || echo "⚠️  Не удалось получить статистику"
echo ""

echo "✅ Проверка завершена!"

