#!/bin/bash
# Скрипт для исправления проблем и запуска полного конвейера

set -e

echo "🔧 Project Horizon - Исправление и запуск"
echo "=========================================="
echo ""

cd ~/Izborator

# Шаг 1: Применение миграций
echo "📦 Шаг 1: Применение миграций..."
docker-compose run --rm backend ./migrate 2>&1 | tail -10 || echo "⚠️  Миграции уже применены или ошибка"
echo ""

# Шаг 2: Проверка таблицы
echo "✅ Шаг 2: Проверка таблицы shop_config_attempts..."
if docker-compose exec -T postgres psql -U postgres -d izborator -c "\d shop_config_attempts" > /dev/null 2>&1; then
  echo "✅ Таблица shop_config_attempts существует"
else
  echo "❌ Таблица shop_config_attempts не существует"
  exit 1
fi
echo ""

# Шаг 3: Загрузка переменных окружения
echo "🔍 Шаг 3: Загрузка переменных окружения..."
if [ -f .env ]; then
  export $(cat .env | grep -v '^#' | xargs)
  echo "✅ Переменные окружения загружены"
else
  echo "❌ Файл .env не найден"
  exit 1
fi
echo ""

# Шаг 4: Статистика до Classifier
echo "📊 Шаг 4: Статистика potential_shops (до Classifier)..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT 
      status,
      COUNT(*) as count
  FROM potential_shops
  GROUP BY status
  ORDER BY status;
"
echo ""

# Шаг 5: Запуск Classifier
echo "🔍 Шаг 5: Запуск Classifier..."
docker-compose run --rm backend ./classifier -classify-all 2>&1 | tee /tmp/classifier.log
CLASSIFIER_EXIT=$?
if [ $CLASSIFIER_EXIT -eq 0 ]; then
  echo "✅ Classifier завершен успешно"
else
  echo "⚠️  Classifier завершился с кодом $CLASSIFIER_EXIT"
  echo "Последние 30 строк логов:"
  tail -30 /tmp/classifier.log
fi
echo ""

# Шаг 6: Статистика после Classifier
echo "📊 Шаг 6: Статистика potential_shops (после Classifier)..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT 
      status,
      COUNT(*) as count
  FROM potential_shops
  GROUP BY status
  ORDER BY status;
"
echo ""

# Шаг 7: Проверка classified магазинов
CLASSIFIED_COUNT=$(docker-compose exec -T postgres psql -U postgres -d izborator -t -c "SELECT COUNT(*) FROM potential_shops WHERE status = 'classified';" | tr -d ' ')

if [ "$CLASSIFIED_COUNT" -gt "0" ]; then
  echo "✅ Найдено $CLASSIFIED_COUNT classified магазинов"
  echo ""
  
  # Шаг 8: Запуск AutoConfig
  echo "🤖 Шаг 8: Запуск AutoConfig (обрабатываем 5 магазинов)..."
  docker-compose run --rm \
    -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
    -e OPENAI_MODEL="${OPENAI_MODEL:-gpt-4o-mini}" \
    backend ./autoconfig -limit 5 2>&1 | tee /tmp/autoconfig.log
  AUTOCONFIG_EXIT=$?
  if [ $AUTOCONFIG_EXIT -eq 0 ]; then
    echo "✅ AutoConfig завершен успешно"
  else
    echo "⚠️  AutoConfig завершился с кодом $AUTOCONFIG_EXIT"
    echo "Последние 30 строк логов:"
    tail -30 /tmp/autoconfig.log
  fi
  echo ""
  
  # Шаг 9: Финальные результаты
  echo "📊 Шаг 9: Финальные результаты..."
  echo ""
  echo "Созданные магазины (AutoConfig):"
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
    LIMIT 10;
  " || echo "⚠️  Не удалось получить результаты"
  echo ""
else
  echo "⚠️  Нет classified магазинов для обработки AutoConfig"
  echo ""
fi

echo "✅ Все шаги завершены!"

