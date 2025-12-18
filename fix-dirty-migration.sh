#!/bin/bash
# Скрипт для исправления dirty migration state

set -e

echo "🔧 Исправление dirty migration state"
echo "===================================="
echo ""

cd ~/Izborator

echo "📊 Шаг 1: Проверка текущего состояния миграций..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT version, dirty FROM schema_migrations;
" || echo "⚠️  Не удалось получить статус"
echo ""

echo "🔧 Шаг 2: Принудительная установка версии 6 (исправление dirty state)..."
docker-compose run --rm backend ./migrate -force 6
echo ""

echo "✅ Шаг 3: Применение миграций..."
docker-compose run --rm backend ./migrate
echo ""

echo "✅ Шаг 4: Проверка таблицы shop_config_attempts..."
if docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts" > /dev/null 2>&1; then
  echo "✅ Таблица shop_config_attempts существует"
else
  echo "❌ Таблица shop_config_attempts не существует"
  exit 1
fi
echo ""

echo "📊 Шаг 5: Финальный статус миграций..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT version, dirty FROM schema_migrations;
"
echo ""

echo "✅ Dirty state исправлен!"

