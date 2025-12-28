-- complete_categories.sql
-- Полный набор категорий для каталога товаров и услуг

-- Дополнительные родительские категории
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    -- Животные и домашние питомцы
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', NULL, 'kucni-ljubimci', 'PETS', 'Kućni ljubimci', 'kućni ljubimci', 1, 140, 'Животные и домашние питомцы'),
    -- Офис и канцелярия
    ('ffffffff-ffff-ffff-ffff-ffffffffffff', NULL, 'kancelarija', 'OFFICE', 'Kancelarijski materijal', 'kancelarijski materijal', 1, 150, 'Офис и канцелярия'),
    -- Подарки и сувениры
    ('00000000-0000-0000-0000-000000000001', NULL, 'pokloni-i-suveniri', 'GIFTS', 'Pokloni i suveniri', 'pokloni i suveniri', 1, 160, 'Подарки и сувениры'),
    -- Путешествия и туризм
    ('00000000-0000-0000-0000-000000000002', NULL, 'putovanja-i-turizam', 'TRAVEL', 'Putovanja i turizam', 'putovanja i turizam', 1, 170, 'Путешествия и туризм'),
    -- Финансовые услуги
    ('00000000-0000-0000-0000-000000000003', NULL, 'finansijske-usluge', 'FINANCIAL', 'Finansijske usluge', 'finansijske usluge', 1, 75, 'Финансовые услуги')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Электроника
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('11111111-1111-1111-1111-111111111204', '11111111-1111-1111-1111-111111111111', 'tableti', 'TABLETS', 'Tableti', 'tableti', 2, 40, 'Планшеты'),
    ('11111111-1111-1111-1111-111111111205', '11111111-1111-1111-1111-111111111111', 'slusalice', 'HEADPHONES', 'Slušalice', 'slušalice', 2, 50, 'Наушники'),
    ('11111111-1111-1111-1111-111111111206', '11111111-1111-1111-1111-111111111111', 'kucna-elektronika', 'HOME_ELECTRONICS', 'Kućna elektronika', 'kućna elektronika', 2, 60, 'Домашняя электроника'),
    ('11111111-1111-1111-1111-111111111207', '11111111-1111-1111-1111-111111111111', 'kompjuteri', 'COMPUTERS', 'Računari i kompjuteri', 'računari i kompjuteri', 2, 70, 'Компьютеры')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Еда и напитки (дополняем)
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('22222222-2222-2222-2222-222222222205', '22222222-2222-2222-2222-222222222222', 'voce-i-povrce', 'FRUITS_VEGETABLES', 'Voće i povrće', 'voće i povrće', 2, 50, 'Фрукты и овощи'),
    ('22222222-2222-2222-2222-222222222206', '22222222-2222-2222-2222-222222222222', 'kafe-i-caj', 'COFFEE_TEA', 'Kafa i čaj', 'kafa i čaj', 2, 60, 'Кофе и чай'),
    ('22222222-2222-2222-2222-222222222207', '22222222-2222-2222-2222-222222222222', 'slatkisi', 'SWEETS', 'Slatkiši', 'slatkiši', 2, 70, 'Сладости'),
    ('22222222-2222-2222-2222-222222222208', '22222222-2222-2222-2222-222222222222', 'za-djecu', 'BABY_FOOD', 'Hrana za decu', 'hrana za decu', 2, 80, 'Детское питание')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Мода и обувь (дополняем)
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('33333333-3333-3333-3333-333333333204', '33333333-3333-3333-3333-333333333333', 'obuca', 'FOOTWEAR', 'Obuća', 'obuća', 2, 40, 'Обувь'),
    ('33333333-3333-3333-3333-333333333205', '33333333-3333-3333-3333-333333333333', 'torbe-i-rančevi', 'BAGS', 'Torbe i rančevi', 'torbe i rančevi', 2, 50, 'Сумки и рюкзаки'),
    ('33333333-3333-3333-3333-333333333206', '33333333-3333-3333-3333-333333333333', 'satovi', 'WATCHES', 'Satovi', 'satovi', 2, 60, 'Часы'),
    ('33333333-3333-3333-3333-333333333207', '33333333-3333-3333-3333-333333333333', 'nakit', 'JEWELRY', 'Nakit', 'nakit', 2, 70, 'Украшения')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Дом и сад
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('44444444-4444-4444-4444-444444444202', '44444444-4444-4444-4444-444444444444', 'tekstil', 'HOME_TEXTILES', 'Tekstil za dom', 'tekstil za dom', 2, 20, 'Текстиль для дома'),
    ('44444444-4444-4444-4444-444444444203', '44444444-4444-4444-4444-444444444444', 'posude', 'KITCHENWARE', 'Posuđe', 'posuđe', 2, 30, 'Посуда'),
    ('44444444-4444-4444-4444-444444444204', '44444444-4444-4444-4444-444444444444', 'dekoracija', 'DECORATION', 'Dekoracija', 'dekoracija', 2, 40, 'Декор'),
    ('44444444-4444-4444-4444-444444444205', '44444444-4444-4444-4444-444444444444', 'basta', 'GARDENING', 'Bašta i vrt', 'bašta i vrt', 2, 50, 'Сад и огород')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Спорт и отдых
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('55555555-5555-5555-5555-555555555501', '55555555-5555-5555-5555-555555555555', 'fitness', 'FITNESS', 'Fitness oprema', 'fitness oprema', 2, 10, 'Фитнес'),
    ('55555555-5555-5555-5555-555555555502', '55555555-5555-5555-5555-555555555555', 'kampovanje', 'CAMPING', 'Kampovanje', 'kampovanje', 2, 20, 'Кемпинг'),
    ('55555555-5555-5555-5555-555555555503', '55555555-5555-5555-5555-555555555555', 'bicikli', 'BICYCLES', 'Bicikli', 'bicikli', 2, 30, 'Велосипеды'),
    ('55555555-5555-5555-5555-555555555504', '55555555-5555-5555-5555-555555555555', 'ribolov', 'FISHING', 'Ribolov', 'ribolov', 2, 40, 'Рыбалка')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Авто-мото
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('66666666-6666-6666-6666-666666666601', '66666666-6666-6666-6666-666666666666', 'delovi', 'AUTO_PARTS', 'Delovi za vozila', 'delovi za vozila', 2, 10, 'Запчасти'),
    ('66666666-6666-6666-6666-666666666602', '66666666-6666-6666-6666-666666666666', 'aksesoari', 'AUTO_ACCESSORIES', 'Aksesoari za vozila', 'aksesoari za vozila', 2, 20, 'Аксессуары для авто'),
    ('66666666-6666-6666-6666-666666666603', '66666666-6666-6666-6666-666666666666', 'motocikli', 'MOTORCYCLES', 'Motocikli', 'motocikli', 2, 30, 'Мотоциклы')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Красота и здоровье
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('88888888-8888-8888-8888-888888888801', '88888888-8888-8888-8888-888888888888', 'kozmetika', 'COSMETICS', 'Kozmetika', 'kozmetika', 2, 10, 'Косметика'),
    ('88888888-8888-8888-8888-888888888802', '88888888-8888-8888-8888-888888888888', 'parfemi', 'PERFUMES', 'Parfemi', 'parfemi', 2, 20, 'Парфюмерия'),
    ('88888888-8888-8888-8888-888888888803', '88888888-8888-8888-8888-888888888888', 'zdravstveni-proizvodi', 'HEALTH_PRODUCTS', 'Zdravstveni proizvodi', 'zdravstveni proizvodi', 2, 30, 'Продукты для здоровья'),
    ('88888888-8888-8888-8888-888888888804', '88888888-8888-8888-8888-888888888888', 'vitamini', 'VITAMINS', 'Vitamini i suplementi', 'vitamini i suplementi', 2, 40, 'Витамины и добавки')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Детские товары
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('99999999-9999-9999-9999-999999999901', '99999999-9999-9999-9999-999999999999', 'igracke', 'TOYS', 'Igračke', 'igračke', 2, 10, 'Игрушки'),
    ('99999999-9999-9999-9999-999999999902', '99999999-9999-9999-9999-999999999999', 'odeca-za-decu', 'KIDS_CLOTHING', 'Odeća za decu', 'odeća za decu', 2, 20, 'Детская одежда'),
    ('99999999-9999-9999-9999-999999999903', '99999999-9999-9999-9999-999999999999', 'bebi-oprema', 'BABY_GEAR', 'Bebi oprema', 'bebi oprema', 2, 30, 'Детское оборудование'),
    ('99999999-9999-9999-9999-999999999904', '99999999-9999-9999-9999-999999999999', 'decije-knjige', 'KIDS_BOOKS', 'Dečije knjige', 'dečije knjige', 2, 40, 'Детские книги')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Книги и медиа
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaab001', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'knjige', 'BOOKS', 'Knjige', 'knjige', 2, 10, 'Книги'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaab002', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'filmovi', 'MOVIES', 'Filmovi', 'filmovi', 2, 20, 'Фильмы'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaab003', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'muzika', 'MUSIC', 'Muzika', 'muzika', 2, 30, 'Музыка'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaab004', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'igre', 'VIDEO_GAMES', 'Video igre', 'video igre', 2, 40, 'Видеоигры')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Строительство и инструменты
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb001', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'alati', 'TOOLS', 'Alati', 'alati', 2, 10, 'Инструменты'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb002', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'gradjevinski-materijali', 'BUILDING_MATERIALS', 'Građevinski materijali', 'građevinski materijali', 2, 20, 'Строительные материалы'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb003', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'elektro-oprema', 'ELECTRICAL', 'Elektro oprema', 'elektro oprema', 2, 30, 'Электрооборудование'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb004', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'santehnika', 'PLUMBING', 'Sanitarna tehnika', 'sanitarna tehnika', 2, 40, 'Сантехника')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Животные и домашние питомцы
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeee01', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'hrana-za-ljubimce', 'PET_FOOD', 'Hrana za kućne ljubimce', 'hrana za kućne ljubimce', 2, 10, 'Корм для животных'),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeee02', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'igracke-za-ljubimce', 'PET_TOYS', 'Igračke za ljubimce', 'igračke za ljubimce', 2, 20, 'Игрушки для животных'),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeee03', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'oprema-za-ljubimce', 'PET_GEAR', 'Oprema za ljubimce', 'oprema za ljubimce', 2, 30, 'Оборудование для животных')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Дочерние категории для Офис и канцелярия
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru) VALUES
    ('ffffffff-ffff-ffff-ffff-fffffffff001', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'kancelarijski-materijal', 'STATIONERY', 'Kancelarijski materijal', 'kancelarijski materijal', 2, 10, 'Канцелярия'),
    ('ffffffff-ffff-ffff-ffff-fffffffff002', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'papir', 'PAPER', 'Papir', 'papir', 2, 20, 'Бумага'),
    ('ffffffff-ffff-ffff-ffff-fffffffff003', 'ffffffff-ffff-ffff-ffff-ffffffffffff', 'toneri', 'TONERS', 'Toneri i kartridži', 'toneri i kartridži', 2, 30, 'Тонеры и картриджи')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;
