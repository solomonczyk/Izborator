#!/bin/bash
# Тестовый скрипт для проверки обновления potential_shops вручную

echo "🔍 Тест обновления potential_shops"
echo "=================================="
echo ""

cd ~/Izborator

echo "1. Получаем первый ID со статусом 'new':"
FIRST_ID=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT id FROM potential_shops WHERE status = 'new' LIMIT 1;" | tr -d ' ')
echo "   ID: $FIRST_ID"
echo ""

if [ -z "$FIRST_ID" ]; then
  echo "❌ Нет записей со статусом 'new'"
  exit 1
fi

echo "2. Получаем domain для этого ID:"
DOMAIN=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT domain FROM potential_shops WHERE id = '$FIRST_ID';" | tr -d ' ')
echo "   Domain: $DOMAIN"
echo ""

echo "3. Пробуем обновить вручную через SQL:"
docker-compose exec -T postgres psql -U postgres -d izborator << SQL
UPDATE potential_shops
SET status = 'classified',
    confidence_score = 0.85,
    classified_at = NOW(),
    updated_at = NOW()
WHERE id = '$FIRST_ID'
RETURNING id, domain, status, confidence_score;
SQL

echo ""
echo "4. Проверяем результат:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT id, domain, status, confidence_score 
  FROM potential_shops 
  WHERE id = '$FIRST_ID';
"

echo ""
echo "✅ Тест завершен"

