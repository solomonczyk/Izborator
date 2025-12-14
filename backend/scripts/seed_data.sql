-- seed_data.sql
-- Наполнение справочников: города и категории
-- Привязка тестового iPhone к категории и городу

-- 1. Очистка (на случай повторного запуска)
-- ВАЖНО: CASCADE удалит связанные данные, используй осторожно!
-- TRUNCATE TABLE cities, categories CASCADE;

-- Вместо TRUNCATE используем DELETE для безопасной очистки
DELETE FROM product_prices WHERE city_id IS NOT NULL;
DELETE FROM products WHERE category_id IS NOT NULL;
DELETE FROM categories WHERE id IS NOT NULL;
DELETE FROM cities WHERE id IS NOT NULL;

-- 2. Вставляем ГОРОДА
-- ВАЖНО: Используем правильную кодировку UTF-8 для сербских символов
INSERT INTO cities (id, slug, name_sr, sort_order, is_active) VALUES
('11111111-1111-1111-1111-111111111111', 'beograd', 'Beograd', 1, true),
('22222222-2222-2222-2222-222222222222', 'novi-sad', 'Novi Sad', 2, true),
('33333333-3333-3333-3333-333333333333', 'nis', E'Ni\u0161', 3, true)
ON CONFLICT (slug) DO UPDATE SET
    name_sr = EXCLUDED.name_sr,
    sort_order = EXCLUDED.sort_order,
    is_active = EXCLUDED.is_active;

-- 3. Вставляем КАТЕГОРИИ
-- Корневая: Электроника
INSERT INTO categories (id, slug, code, name_sr, name_sr_lc, level, parent_id, is_active, sort_order) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'elektronika', 'electronics', 'Elektronika', 'elektronika', 1, NULL, true, 1)
ON CONFLICT (slug) DO UPDATE SET
    code = EXCLUDED.code,
    name_sr = EXCLUDED.name_sr,
    name_sr_lc = EXCLUDED.name_sr_lc,
    level = EXCLUDED.level,
    parent_id = EXCLUDED.parent_id,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order;

-- Дочерние: Мобильные телефоны и Ноутбуки (родитель - Электроника)
INSERT INTO categories (id, slug, code, name_sr, name_sr_lc, level, parent_id, is_active, sort_order) VALUES
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'mobilni-telefoni', 'mobile-phones', 'Mobilni telefoni', 'mobilni telefoni', 2, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', true, 1),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'laptopovi', 'laptops', 'Laptopovi', 'laptopovi', 2, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', true, 2)
ON CONFLICT (slug) DO UPDATE SET
    code = EXCLUDED.code,
    name_sr = EXCLUDED.name_sr,
    name_sr_lc = EXCLUDED.name_sr_lc,
    level = EXCLUDED.level,
    parent_id = EXCLUDED.parent_id,
    is_active = EXCLUDED.is_active,
    sort_order = EXCLUDED.sort_order;

-- 4. СВЯЗЫВАЕМ СУЩЕСТВУЮЩИЙ ТОВАР
-- Находим наш iPhone (по названию) и привязываем к "Mobilni telefoni"
UPDATE products 
SET category_id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'
WHERE name ILIKE '%iPhone%'
  AND category_id IS NULL;

-- 5. ОБНОВЛЯЕМ ЦЕНУ (Привязываем к городу Белград)
-- Находим цену для iPhone и ставим city_id = Beograd
UPDATE product_prices
SET city_id = '11111111-1111-1111-1111-111111111111'
WHERE product_id IN (SELECT id FROM products WHERE name ILIKE '%iPhone%')
  AND city_id IS NULL;

-- 6. Проверка результата (опционально)
-- SELECT 'Cities:' as info, COUNT(*) as count FROM cities;
-- SELECT 'Categories:' as info, COUNT(*) as count FROM categories;
-- SELECT 'Products with category:' as info, COUNT(*) as count FROM products WHERE category_id IS NOT NULL;
-- SELECT 'Prices with city:' as info, COUNT(*) as count FROM product_prices WHERE city_id IS NOT NULL;

