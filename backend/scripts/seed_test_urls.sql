-- Тестовые URL для парсинга по категориям
-- Обновляй эти URL, если они перестали работать

-- ВАЖНО: Сначала найди валидные URL на gigatron.rs для каждой категории
-- Инструкции: см. backend/scripts/test_urls_by_category.md

-- 1. Mobilni telefoni (Мобильные телефоны)
-- Категория ID: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb
-- ЗАМЕНИ НА РЕАЛЬНЫЙ URL:
-- INSERT INTO raw_products (shop_id, shop_name, external_id, url, name, price, currency, in_stock, processed)
-- VALUES (
--     'shop-001',
--     'Gigatron Test',
--     'test-phone-001',
--     'https://gigatron.rs/mobilni-telefoni/apple-iphone-15-128gb-black-mtp03zda-573380', -- ЗАМЕНИ НА РЕАЛЬНЫЙ URL
--     'Test Phone',
--     0, -- Цена будет заполнена парсером
--     'RSD',
--     true,
--     false
-- )
-- ON CONFLICT (shop_id, external_id) DO NOTHING;

-- 2. Laptopovi (Ноутбуки)
-- Категория ID: cccccccc-cccc-cccc-cccc-cccccccccccc
-- ЗАМЕНИ НА РЕАЛЬНЫЙ URL:
-- INSERT INTO raw_products (shop_id, shop_name, external_id, url, name, price, currency, in_stock, processed)
-- VALUES (
--     'shop-001',
--     'Gigatron Test',
--     'test-laptop-001',
--     'https://gigatron.rs/laptopovi/lenovo-ideapad-3-15-82h7000vra', -- ЗАМЕНИ НА РЕАЛЬНЫЙ URL
--     'Test Laptop',
--     0, -- Цена будет заполнена парсером
--     'RSD',
--     true,
--     false
-- )
-- ON CONFLICT (shop_id, external_id) DO NOTHING;

-- ПРИМЕЧАНИЕ: Этот скрипт не выполняется автоматически
-- Используй его как шаблон для добавления тестовых URL
-- Или используй напрямую команду воркера:
-- go run cmd/worker/main.go -url "ТВОЙ_URL" -shop "shop-001"

