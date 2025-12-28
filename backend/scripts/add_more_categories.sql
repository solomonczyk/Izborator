-- add_more_categories.sql
-- Добавление недостающих родительских категорий для полного каталога товаров и услуг

-- Услуги (важно - для сервисов)
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    ('77777777-7777-7777-7777-777777777777', NULL, 'usluge', 'SERVICES', 'Usluge', 'usluge', 1, 70, 'Услуги')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Красота и здоровье
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    ('88888888-8888-8888-8888-888888888888', NULL, 'lepota-i-zdravlje', 'BEAUTY_HEALTH', 'Lepota i zdravlje', 'lepota i zdravlje', 1, 80, 'Красота и здоровье')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Детские товары
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    ('99999999-9999-9999-9999-999999999999', NULL, 'deciji-proizvodi', 'KIDS', 'Dečiji proizvodi', 'dečiji proizvodi', 1, 90, 'Детские товары')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Книги и медиа
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL, 'knjige-i-medija', 'BOOKS_MEDIA', 'Knjige i medija', 'knjige i medija', 1, 100, 'Книги и медиа')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Строительные материалы и инструменты
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', NULL, 'gradevina-i-alati', 'CONSTRUCTION_TOOLS', 'Građevina i alati', 'građevina i alati', 1, 110, 'Строительство и инструменты')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Игрушки и игры
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', NULL, 'igracke-i-igre', 'TOYS_GAMES', 'Igračke i igre', 'igračke i igre', 1, 120, 'Игрушки и игры')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Хобби и коллекционирование
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', NULL, 'hobiji-i-kolekcionarstvo', 'HOBBY_COLLECTIBLES', 'Hobiji i kolekcionarstvo', 'hobiji i kolekcionarstvo', 1, 130, 'Хобби и коллекционирование')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

-- Добавляем дочерние категории для услуг (примеры)
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order, name_ru)
VALUES
    -- Услуги
    ('77777777-7777-7777-7777-777777777701', '77777777-7777-7777-7777-777777777777', 'frizerske-usluge', 'HAIRDRESSING', 'Frizerske usluge', 'frizerske usluge', 2, 10, 'Парикмахерские услуги'),
    ('77777777-7777-7777-7777-777777777702', '77777777-7777-7777-7777-777777777777', 'zdravstvene-usluge', 'MEDICAL', 'Zdravstvene usluge', 'zdravstvene usluge', 2, 20, 'Медицинские услуги'),
    ('77777777-7777-7777-7777-777777777703', '77777777-7777-7777-7777-777777777777', 'prevozne-usluge', 'TRANSPORT', 'Prevozne usluge', 'prevozne usluge', 2, 30, 'Транспортные услуги'),
    ('77777777-7777-7777-7777-777777777704', '77777777-7777-7777-7777-777777777777', 'popravke-i-servisi', 'REPAIR_SERVICE', 'Popravke i servisi', 'popravke i servisi', 2, 40, 'Ремонт и сервис'),
    ('77777777-7777-7777-7777-777777777705', '77777777-7777-7777-7777-777777777777', 'pravne-usluge', 'LEGAL', 'Pravne usluge', 'pravne usluge', 2, 50, 'Юридические услуги'),
    -- Еда и напитки - добавляем еще подкатегории
    ('22222222-2222-2222-2222-222222222202', '22222222-2222-2222-2222-222222222222', 'meso-i-mesni-proizvodi', 'MEAT', 'Meso i mesni proizvodi', 'meso i mesni proizvodi', 2, 20, 'Мясо и мясные изделия'),
    ('22222222-2222-2222-2222-222222222203', '22222222-2222-2222-2222-222222222222', 'hleb-i-pekarski-proizvodi', 'BAKERY', 'Hleb i pekarski proizvodi', 'hleb i pekarski proizvodi', 2, 30, 'Хлеб и хлебобулочные изделия'),
    ('22222222-2222-2222-2222-222222222204', '22222222-2222-2222-2222-222222222222', 'pice', 'BEVERAGES', 'Piće', 'piće', 2, 40, 'Напитки'),
    -- Мода и обувь - добавляем еще подкатегории
    ('33333333-3333-3333-3333-333333333202', '33333333-3333-3333-3333-333333333333', 'garderoba', 'CLOTHING', 'Garderoba', 'garderoba', 2, 20, 'Одежда'),
    ('33333333-3333-3333-3333-333333333203', '33333333-3333-3333-3333-333333333333', 'accessories', 'ACCESSORIES', 'Aksesoari', 'aksesoari', 2, 30, 'Аксессуары')
ON CONFLICT (code) DO UPDATE SET name_ru = EXCLUDED.name_ru;

