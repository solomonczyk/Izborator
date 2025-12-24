#!/bin/bash
# Скрипт для пересборки backend и тестирования Classifier

set -e

echo "🔧 Пересборка Backend и тестирование Classifier"
echo "================================================"
echo ""

# Шаг 1: Обновление кода
echo "📥 Шаг 1: Обновление кода..."
git fetch origin --prune
git reset --hard origin/main
git clean -fd
echo "✅ Код обновлен"
echo ""

# Шаг 2: Пересборка backend
echo "🔨 Шаг 2: Пересборка backend контейнера..."
docker-compose build --no-cache backend
if [ $? -ne 0 ]; then
  echo "❌ Ошибка сборки backend"
  exit 1
fi
echo "✅ Backend пересобран"
echo ""

# Шаг 3: Перезапуск контейнеров
echo "🔄 Шаг 3: Перезапуск контейнеров..."
docker-compose up -d
sleep 10
echo "✅ Контейнеры перезапущены"
echo ""

# Шаг 4: Статистика ДО Classifier
echo "📊 Шаг 4: Статистика potential_shops (ДО Classifier)..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT status, COUNT(*) as count
  FROM potential_shops
  GROUP BY status
  ORDER BY status;
"
echo ""

# Шаг 5: Запуск Classifier с детальными логами
echo "🔍 Шаг 5: Запуск Classifier..."
echo "Логи будут сохранены в /tmp/classifier.log"
docker-compose run --rm backend ./classifier -classify-all 2>&1 | tee /tmp/classifier.log
CLASSIFIER_EXIT=$?
echo ""

if [ $CLASSIFIER_EXIT -eq 0 ]; then
  echo "✅ Classifier завершился успешно"
else
  echo "⚠️  Classifier завершился с кодом $CLASSIFIER_EXIT"
  echo ""
  echo "📋 Последние 50 строк логов:"
  tail -50 /tmp/classifier.log
  echo ""
  echo "📋 Поиск ошибок в логах:"
  grep -i "error\|failed\|update" /tmp/classifier.log | tail -20 || echo "Ошибок не найдено"
fi
echo ""

# Шаг 6: Статистика ПОСЛЕ Classifier
echo "📊 Шаг 6: Статистика potential_shops (ПОСЛЕ Classifier)..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT status, COUNT(*) as count
  FROM potential_shops
  GROUP BY status
  ORDER BY status;
"
echo ""

# Шаг 7: Проверка конкретных записей
echo "🔍 Шаг 7: Проверка нескольких записей (первые 5)..."
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT domain, status, confidence_score, classified_at
  FROM potential_shops
  ORDER BY updated_at DESC
  LIMIT 5;
"
echo ""

echo "✅ Проверка завершена!"
echo ""
echo "💡 Если статусы не обновились, проверь:"
echo "   1. Логи Classifier: tail -100 /tmp/classifier.log"
echo "   2. Логи backend: docker-compose logs backend | grep -i error"
echo "   3. Проверь, что domain совпадает в таблице и в коде"

