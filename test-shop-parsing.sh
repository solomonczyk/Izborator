#!/bin/bash
# Скрипт для тестирования парсинга автоматически созданных магазинов

set -e

cd ~/Izborator

echo "🧪 Тестирование парсинга автоматически созданных магазинов"
echo "=========================================="

# Получаем ID и base_url автоматически созданных магазинов
echo ""
echo "📋 Получаем список магазинов для тестирования..."
SHOPS=$(docker exec -i izborator_postgres psql -U postgres -d izborator -t -c "
SELECT id || '|' || base_url || '|' || name
FROM shops
WHERE is_auto_configured = true
ORDER BY created_at DESC;
")

if [ -z "$SHOPS" ]; then
    echo "❌ Не найдено автоматически созданных магазинов"
    exit 1
fi

# Для каждого магазина находим тестовый URL товара
echo "$SHOPS" | while IFS='|' read -r shop_id base_url shop_name; do
    # Убираем пробелы
    shop_id=$(echo "$shop_id" | tr -d '[:space:]')
    base_url=$(echo "$base_url" | tr -d '[:space:]')
    shop_name=$(echo "$shop_name" | tr -d '[:space:]')
    
    if [ -z "$shop_id" ]; then
        continue
    fi
    
    echo ""
    echo "🔍 Тестируем магазин: $shop_name"
    echo "   ID: $shop_id"
    echo "   URL: $base_url"
    
    # Получаем селекторы для этого магазина
    SELECTORS=$(docker exec -i izborator_postgres psql -U postgres -d izborator -t -c "
    SELECT selectors->>'name' || '|' || selectors->>'price' || '|' || selectors->>'image'
    FROM shops
    WHERE id = '$shop_id';
    ")
    
    echo "   Селекторы: $SELECTORS"
    
    # Пробуем найти тестовый URL товара из shop_config_attempts
    TEST_URL=$(docker exec -i izborator_postgres psql -U postgres -d izborator -t -c "
    SELECT html_sample
    FROM shop_config_attempts
    WHERE shop_id = '$shop_id'
    AND status = 'success'
    ORDER BY created_at DESC
    LIMIT 1;
    " | grep -o 'https://[^[:space:]]*' | head -1)
    
    # Если не нашли в attempts, пробуем найти из potential_shops
    if [ -z "$TEST_URL" ]; then
        TEST_URL=$(docker exec -i izborator_postgres psql -U postgres -d izborator -t -c "
        SELECT metadata->>'product_url'
        FROM potential_shops
        WHERE id IN (
            SELECT potential_shop_id
            FROM shop_config_attempts
            WHERE shop_id = '$shop_id'
            LIMIT 1
        );
        " | tr -d '[:space:]')
    fi
    
    # Если все еще нет URL, используем base_url + типичный путь
    if [ -z "$TEST_URL" ] || [ "$TEST_URL" = "null" ]; then
        # Для istyle.rs - пробуем найти MacBook
        if [[ "$base_url" == *"istyle.rs"* ]]; then
            TEST_URL="https://istyle.rs/products/13-incni-macbook-air-m3-sa-8-jezgarnim-cpu-om-8-jezgarnim-gpu-om-8gb-objedinjene-memorije-i-256gb-ssd-om-space-gray-copy-1"
        # Для stana.rs - пробуем найти товар из каталога
        elif [[ "$base_url" == *"stana.rs"* ]]; then
            TEST_URL="https://stana.rs/psi/oprema-za-pse/kreveti-za-pse/"
        else
            TEST_URL="$base_url"
        fi
    fi
    
    echo "   Тестовый URL: $TEST_URL"
    
    # Запускаем парсинг через worker
    echo ""
    echo "🚀 Запуск парсинга..."
    docker-compose run --rm worker ./worker -url "$TEST_URL" -shop "$shop_id" 2>&1 | tee /tmp/parsing-test-$shop_id.log
    
    # Проверяем результат
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        echo "✅ Парсинг успешен для $shop_name"
        
        # Проверяем, что товар сохранился
        PRODUCT_COUNT=$(docker exec -i izborator_postgres psql -U postgres -d izborator -t -c "
        SELECT COUNT(*)
        FROM raw_products
        WHERE shop_id = '$shop_id';
        " | tr -d '[:space:]')
        
        echo "   Сохранено товаров: $PRODUCT_COUNT"
    else
        echo "❌ Парсинг не удался для $shop_name"
        echo "   Смотри логи: /tmp/parsing-test-$shop_id.log"
    fi
    
    echo ""
    echo "---"
done

echo ""
echo "📊 Итоговая статистика:"
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    s.name as shop_name,
    COUNT(rp.shop_id) as total_products,
    COUNT(rp.shop_id) FILTER (WHERE rp.processed = true) as processed,
    COUNT(rp.shop_id) FILTER (WHERE rp.processed = false) as unprocessed
FROM shops s
LEFT JOIN raw_products rp ON rp.shop_id = s.id
WHERE s.is_auto_configured = true
GROUP BY s.id, s.name
ORDER BY s.created_at DESC;
"

echo ""
echo "✅ Тестирование завершено!"
