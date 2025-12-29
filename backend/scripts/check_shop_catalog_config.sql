-- Проверка конфигурации каталогов для всех магазинов
-- Этот скрипт показывает, какие магазины имеют настройки для discovery

SELECT 
    id,
    name,
    base_url,
    is_active,
    CASE 
        WHEN selectors ? 'catalog_url' THEN selectors->>'catalog_url'
        ELSE '❌ Не указан'
    END as catalog_url,
    CASE 
        WHEN selectors ? 'catalog_product_link' THEN selectors->>'catalog_product_link'
        ELSE '❌ Не указан (используется дефолтный)'
    END as catalog_product_link,
    CASE 
        WHEN selectors ? 'catalog_next_page' THEN selectors->>'catalog_next_page'
        ELSE '❌ Не указан (используется дефолтный)'
    END as catalog_next_page,
    CASE 
        WHEN is_active = true AND selectors ? 'catalog_url' THEN '✅ Готов к discovery'
        WHEN is_active = true AND NOT (selectors ? 'catalog_url') THEN '⚠️ Активен, но нет catalog_url'
        WHEN is_active = false THEN '❌ Неактивен'
        ELSE '❓ Неизвестный статус'
    END as discovery_status
FROM shops
ORDER BY 
    CASE 
        WHEN is_active = true AND selectors ? 'catalog_url' THEN 1
        WHEN is_active = true THEN 2
        ELSE 3
    END,
    name;

-- Детальная информация о селекторах для активных магазинов
SELECT 
    name,
    jsonb_pretty(selectors) as selectors_json
FROM shops
WHERE is_active = true
ORDER BY name;

