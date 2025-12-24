#!/bin/bash
# Скрипт для проверки состояния базы данных перед запуском AutoConfig

echo "📊 Проверка состояния базы данных"
echo "=================================="
echo ""

echo "1️⃣ Статистика по potential_shops (статусы):"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count
FROM potential_shops
GROUP BY status
ORDER BY status;
" 2>/dev/null || echo "⚠️  Не удалось получить статистику"
echo ""

echo "2️⃣ Классифицированные кандидаты (для AutoConfig):"
CLASSIFIED_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -A -c "
SELECT COUNT(*) 
FROM potential_shops 
WHERE status = 'classified';
" 2>/dev/null | tr -d ' \n')

if [ -n "$CLASSIFIED_COUNT" ] && [ "$CLASSIFIED_COUNT" -gt 0 ]; then
  echo "✅ Найдено $CLASSIFIED_COUNT классифицированных кандидатов"
else
  echo "⚠️  Классифицированных кандидатов не найдено (0)"
  echo "💡 Нужно запустить Classifier: docker-compose run --rm backend ./classifier -classify-all"
fi
echo ""

echo "3️⃣ Автоматически созданные магазины:"
AUTO_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -A -c "
SELECT COUNT(*) 
FROM shops 
WHERE is_auto_configured = true;
" 2>/dev/null | tr -d ' \n')

if [ -n "$AUTO_COUNT" ]; then
  echo "📦 Всего автоматически создано магазинов: $AUTO_COUNT"
else
  echo "⚠️  Не удалось получить количество"
fi
echo ""

echo "✅ Проверка завершена!"

