#!/bin/bash
# Скрипт для запуска AutoConfig на сервере с правильной загрузкой переменных окружения

set -e

cd ~/Izborator

echo "🤖 Запуск AutoConfig..."
echo "=========================================="

# Проверяем наличие .env файла
if [ ! -f .env ]; then
  echo "❌ Файл .env не найден!"
  echo "⚠️  Создайте .env из .env.example и заполните значения"
  exit 1
fi

# Загружаем переменные из .env
echo "📝 Загрузка переменных из .env..."
export $(cat .env | grep -v '^#' | xargs)

# Проверяем наличие OPENAI_API_KEY
if [ -z "$OPENAI_API_KEY" ] || [ "$OPENAI_API_KEY" = "your_openai_api_key_here" ]; then
  echo "❌ OPENAI_API_KEY не настроен в .env файле!"
  echo "⚠️  Откройте .env и установите реальный ключ OpenAI"
  echo "📝 Создайте ключ на https://platform.openai.com/account/api-keys"
  exit 1
fi

# Проверяем формат ключа (должен начинаться с sk-)
if [[ ! "$OPENAI_API_KEY" =~ ^sk- ]]; then
  echo "⚠️  Предупреждение: OPENAI_API_KEY не начинается с 'sk-'"
  echo "   Убедитесь, что ключ правильный"
fi

# Удаляем пробелы и переносы строк из ключа (на случай, если они есть)
OPENAI_API_KEY=$(echo "$OPENAI_API_KEY" | tr -d '[:space:]')

# Показываем первые и последние символы для проверки (безопасно)
KEY_PREVIEW="${OPENAI_API_KEY:0:10}...${OPENAI_API_KEY: -4}"
echo "✅ OPENAI_API_KEY загружен (${KEY_PREVIEW})"
echo "📏 Длина ключа: ${#OPENAI_API_KEY} символов"
echo ""

# Запускаем AutoConfig
echo "🚀 Запуск AutoConfig (limit=10)..."
docker-compose run --rm \
  -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
  -e OPENAI_MODEL="${OPENAI_MODEL:-gpt-4o-mini}" \
  backend ./autoconfig -limit 10

echo ""
echo "✅ AutoConfig завершен!"
echo ""

# Показываем результаты
echo "📊 Результаты:"
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    id,
    name, 
    base_url, 
    is_active, 
    is_auto_configured,
    created_at 
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC
LIMIT 10;
" || echo "⚠️  Не удалось получить результаты"

