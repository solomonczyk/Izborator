#!/bin/bash
# Скрипт для исправления dirty migration state

set +e  # Не останавливаться при ошибках

echo "🔧 Исправление dirty migration state"
echo "===================================="
echo ""

cd ~/Izborator

echo "📊 Шаг 1: Проверка текущего состояния миграций..."
MIGRATION_STATUS=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT version, dirty FROM schema_migrations;" 2>/dev/null)
if [ $? -eq 0 ]; then
  echo "$MIGRATION_STATUS"
else
  echo "⚠️  Не удалось получить статус"
fi
echo ""

echo "🔧 Шаг 2: Принудительная установка версии 6 (исправление dirty state)..."
FORCE_OUTPUT=$(docker-compose run --rm backend ./migrate -force 6 2>&1)
FORCE_EXIT=$?
if [ $FORCE_EXIT -eq 0 ]; then
  echo "✅ Версия принудительно установлена"
else
  echo "⚠️  Ошибка при установке версии (возможно, уже исправлено):"
  echo "$FORCE_OUTPUT" | tail -5
fi
echo ""

echo "✅ Шаг 3: Применение миграций..."
MIGRATE_OUTPUT=$(docker-compose run --rm backend ./migrate 2>&1)
MIGRATE_EXIT=$?
if [ $MIGRATE_EXIT -eq 0 ]; then
  echo "✅ Миграции применены успешно"
  echo "$MIGRATE_OUTPUT" | tail -5
else
  echo "⚠️  Ошибка применения миграций:"
  echo "$MIGRATE_OUTPUT" | tail -10
  # Проверяем, может быть миграции уже применены
  if echo "$MIGRATE_OUTPUT" | grep -q "no change"; then
    echo "✅ Миграции уже применены (no change)"
    MIGRATE_EXIT=0
  fi
fi
echo ""

echo "✅ Шаг 4: Проверка таблицы shop_config_attempts..."
if docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts" > /dev/null 2>&1; then
  echo "✅ Таблица shop_config_attempts существует"
  TABLE_EXISTS=1
else
  echo "❌ Таблица shop_config_attempts не существует"
  TABLE_EXISTS=0
fi
echo ""

echo "📊 Шаг 5: Финальный статус миграций..."
FINAL_STATUS=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT version, dirty FROM schema_migrations;" 2>/dev/null)
if [ $? -eq 0 ]; then
  echo "$FINAL_STATUS"
else
  echo "⚠️  Не удалось получить финальный статус"
fi
echo ""

if [ $TABLE_EXISTS -eq 1 ] && [ $MIGRATE_EXIT -eq 0 ]; then
  echo "✅ Dirty state исправлен и миграции применены!"
  exit 0
else
  echo "⚠️  Есть проблемы: таблица существует=$TABLE_EXISTS, миграции применены=$MIGRATE_EXIT"
  exit 1
fi

