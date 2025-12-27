#!/bin/bash
# Скрипт для проверки данных в продакшене для тестирования AutoConfig

echo "🔍 Проверка данных для тестирования AutoConfig"
echo "=============================================="
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. Проверка потенциальных магазинов
echo -e "${BLUE}1️⃣ Статистика по potential_shops:${NC}"
psql $DATABASE_URL -c "
SELECT 
    status,
    COUNT(*) as count,
    COUNT(*) FILTER (WHERE metadata->>'site_type' = 'service_provider') as service_providers,
    COUNT(*) FILTER (WHERE metadata->>'site_type' = 'ecommerce') as ecommerce,
    MAX(confidence_score) as max_score,
    AVG(confidence_score) as avg_score
FROM potential_shops
GROUP BY status
ORDER BY status;
" 2>/dev/null || echo "⚠️  Не удалось подключиться к БД. Проверьте DATABASE_URL"

echo ""

# 2. Классифицированные кандидаты (готовы для AutoConfig)
echo -e "${BLUE}2️⃣ Классифицированные кандидаты (готовы для AutoConfig):${NC}"
CLASSIFIED_COUNT=$(psql $DATABASE_URL -t -A -c "
SELECT COUNT(*) 
FROM potential_shops 
WHERE status = 'classified';
" 2>/dev/null | tr -d ' \n')

if [ -n "$CLASSIFIED_COUNT" ] && [ "$CLASSIFIED_COUNT" -gt "0" ]; then
    echo -e "${GREEN}✅ Найдено $CLASSIFIED_COUNT классифицированных кандидатов${NC}"
    echo ""
    
    # Показываем примеры
    echo "   Примеры кандидатов:"
    psql $DATABASE_URL -c "
    SELECT 
        domain,
        status,
        confidence_score,
        metadata->>'site_type' as site_type,
        discovered_at
    FROM potential_shops 
    WHERE status = 'classified'
    ORDER BY confidence_score DESC, discovered_at DESC
    LIMIT 5;
    " 2>/dev/null
    
    # Количество service_provider
    SERVICE_PROVIDER_COUNT=$(psql $DATABASE_URL -t -A -c "
    SELECT COUNT(*) 
    FROM potential_shops 
    WHERE status = 'classified' 
    AND metadata->>'site_type' = 'service_provider';
    " 2>/dev/null | tr -d ' \n')
    
    if [ -n "$SERVICE_PROVIDER_COUNT" ] && [ "$SERVICE_PROVIDER_COUNT" -gt "0" ]; then
        echo -e "${GREEN}   ✅ Из них service_provider: $SERVICE_PROVIDER_COUNT (отлично для тестирования таблиц!)${NC}"
    else
        echo -e "${YELLOW}   ⚠️  Service providers не найдены. Нужно запустить Discovery для поиска услуг.${NC}"
    fi
else
    echo -e "${RED}❌ Классифицированных кандидатов нет (0)${NC}"
    echo -e "${YELLOW}   Нужно запустить:${NC}"
    echo "   1. Discovery (поиск кандидатов)"
    echo "   2. Classifier (классификация)"
fi

echo ""

# 3. Уже созданные магазины через AutoConfig
echo -e "${BLUE}3️⃣ Магазины, созданные через AutoConfig:${NC}"
AUTOCONFIG_COUNT=$(psql $DATABASE_URL -t -A -c "
SELECT COUNT(*) 
FROM shops 
WHERE is_auto_configured = true;
" 2>/dev/null | tr -d ' \n')

if [ -n "$AUTOCONFIG_COUNT" ] && [ "$AUTOCONFIG_COUNT" -gt "0" ]; then
    echo -e "${GREEN}✅ Найдено $AUTOCONFIG_COUNT автоматически созданных магазинов${NC}"
    echo ""
    echo "   Последние созданные:"
    psql $DATABASE_URL -c "
    SELECT 
        name,
        base_url,
        is_active,
        ai_config_model,
        CASE 
            WHEN selectors->>'name' IS NOT NULL THEN '✅'
            ELSE '❌'
        END as has_name,
        CASE 
            WHEN selectors->>'price' IS NOT NULL THEN '✅'
            ELSE '❌'
        END as has_price,
        created_at
    FROM shops 
    WHERE is_auto_configured = true 
    ORDER BY created_at DESC 
    LIMIT 5;
    " 2>/dev/null
else
    echo -e "${YELLOW}⚠️  Автоматически созданных магазинов нет${NC}"
fi

echo ""

# 4. Рекомендации
echo -e "${BLUE}4️⃣ Рекомендации для тестирования:${NC}"
echo ""

if [ -z "$CLASSIFIED_COUNT" ] || [ "$CLASSIFIED_COUNT" -eq "0" ]; then
    echo -e "${YELLOW}📋 Для получения данных выполните:${NC}"
    echo ""
    echo "   1. Запустить Discovery для поиска сайтов услуг:"
    echo "      ./backend/discovery -max-results 100"
    echo ""
    echo "   2. Запустить Classifier для классификации:"
    echo "      ./backend/classifier -classify-all -limit 50"
    echo ""
    echo "   3. Запустить AutoConfig для тестирования:"
    echo "      ./backend/autoconfig -limit 5"
    echo ""
elif [ -n "$SERVICE_PROVIDER_COUNT" ] && [ "$SERVICE_PROVIDER_COUNT" -gt "0" ]; then
    echo -e "${GREEN}✅ Отлично! Есть данные для тестирования табличных данных${NC}"
    echo ""
    echo "   Запустите AutoConfig:"
    echo "   ./backend/autoconfig -limit 3"
    echo ""
    echo "   Или для тестирования только service_provider:"
    echo "   (нужно будет добавить фильтр в код)"
else
    echo -e "${YELLOW}⚠️  Есть классифицированные кандидаты, но нет service_provider${NC}"
    echo ""
    echo "   Для получения service_provider запустите Discovery с запросами для услуг:"
    echo "   ./backend/discovery -max-results 200"
    echo ""
    echo "   Discovery уже содержит запросы для услуг (стоматология, красота, ремонт и т.д.)"
fi

echo ""
echo -e "${GREEN}✅ Проверка завершена!${NC}"

