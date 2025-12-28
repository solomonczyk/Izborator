#!/bin/bash
# Скрипт для проверки пароля PostgreSQL на сервере

echo "🔍 Проверка настроек PostgreSQL"
echo "================================="
echo ""

# Проверяем, какой пароль используется в .env
if [ -f ~/Izborator/.env ]; then
    echo "📄 Найден .env файл в корне проекта:"
    echo ""
    grep -E "^DB_|^POSTGRES_" ~/Izborator/.env | grep -v PASSWORD || echo "⚠️  Переменные DB_* не найдены"
    echo ""
    
    DB_PASSWORD=$(grep "^DB_PASSWORD=" ~/Izborator/.env | cut -d'=' -f2)
    if [ -n "$DB_PASSWORD" ]; then
        echo "✅ DB_PASSWORD найден в .env (длина: ${#DB_PASSWORD} символов)"
    else
        echo "❌ DB_PASSWORD не найден в .env"
    fi
else
    echo "❌ Файл .env не найден в ~/Izborator/.env"
fi

echo ""
echo "📊 Проверка контейнера PostgreSQL:"
docker ps --filter "name=izborator_postgres" --format "table {{.Names}}\t{{.Status}}"

echo ""
echo "💡 Если пароль не совпадает, нужно:"
echo "   1. Проверить пароль в .env файле"
echo "   2. Или пересоздать контейнер postgres с правильным паролем"
echo ""

