-- seed_catalog_core_sr.sql
-- Стартовые категории, типы, атрибуты и города для сербского рынка

-- 1. Разделы (level = 1)
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order)
VALUES
    ('11111111-1111-1111-1111-111111111111', NULL, 'elektronika', 'ELEKTRONIKA', 'Elektronika', 'elektronika', 1, 10),
    ('22222222-2222-2222-2222-222222222222', NULL, 'hrana-i-pice', 'HRANA_PICE', 'Hrana i piće', 'hrana i piće', 1, 20),
    ('33333333-3333-3333-3333-333333333333', NULL, 'moda', 'MODA', 'Moda i obuća', 'moda i obuća', 1, 30),
    ('44444444-4444-4444-4444-444444444444', NULL, 'dom-i-basta', 'DOM_BASTA', 'Dom i bašta', 'dom i bašta', 1, 40),
    ('55555555-5555-5555-5555-555555555555', NULL, 'sport', 'SPORT', 'Sport i rekreacija', 'sport i rekreacija', 1, 50),
    ('66666666-6666-6666-6666-666666666666', NULL, 'auto-moto', 'AUTO_MOTO', 'Auto-moto', 'auto-moto', 1, 60)
ON CONFLICT (code) DO NOTHING;

-- 2. Категории внутри разделов (level = 2)
INSERT INTO categories (id, parent_id, slug, code, name_sr, name_sr_lc, level, sort_order)
VALUES
    -- Elektronika
    ('11111111-1111-1111-1111-111111111201', '11111111-1111-1111-1111-111111111111', 'mobilni-telefoni', 'PHONES', 'Mobilni telefoni', 'mobilni telefoni', 2, 10),
    ('11111111-1111-1111-1111-111111111202', '11111111-1111-1111-1111-111111111111', 'laptopovi', 'LAPTOPS', 'Laptopovi', 'laptopovi', 2, 20),
    ('11111111-1111-1111-1111-111111111203', '11111111-1111-1111-1111-111111111111', 'televizori', 'TVS', 'Televizori', 'televizori', 2, 30),
    -- Hrana i piće
    ('22222222-2222-2222-2222-222222222201', '22222222-2222-2222-2222-222222222222', 'mleko-i-mlecni-proizvodi', 'MILK_DAIRY', 'Mleko i mlečni proizvodi', 'mleko i mlečni proizvodi', 2, 10),
    -- Moda
    ('33333333-3333-3333-3333-333333333201', '33333333-3333-3333-3333-333333333333', 'patike', 'SNEAKERS', 'Patike', 'patike', 2, 10),
    -- Dom i bašta
    ('44444444-4444-4444-4444-444444444201', '44444444-4444-4444-4444-444444444444', 'namestaj', 'FURNITURE', 'Nameštaj', 'nameštaj', 2, 10)
ON CONFLICT (code) DO NOTHING;

-- 3. Product types
INSERT INTO product_types (id, code, name_sr)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'SMARTPHONE', 'Pametni telefon'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'LAPTOP', 'Laptop'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', 'TV', 'Televizor'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', 'MILK', 'Mleko'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa5', 'SNEAKERS', 'Patike'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa6', 'FURNITURE', 'Komad nameštaja')
ON CONFLICT (code) DO NOTHING;

-- 4. Category ↔ product_types
INSERT INTO category_product_types (category_id, product_type_id)
VALUES
    ('11111111-1111-1111-1111-111111111201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1'),
    ('11111111-1111-1111-1111-111111111202', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2'),
    ('11111111-1111-1111-1111-111111111203', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3'),
    ('22222222-2222-2222-2222-222222222201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4'),
    ('33333333-3333-3333-3333-333333333201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa5'),
    ('44444444-4444-4444-4444-444444444201', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa6')
ON CONFLICT DO NOTHING;

-- 5. Attributes
INSERT INTO attributes (id, code, name_sr, data_type, unit_sr, is_filterable, is_sortable)
VALUES
    -- Общие
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb001', 'BRAND', 'Brend', 'string', NULL, TRUE, FALSE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb002', 'COLOR', 'Boja', 'string', NULL, TRUE, FALSE),
    -- Для смартфонов / ноутбуков / ТВ
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb101', 'RAM', 'RAM', 'int', 'GB', TRUE, TRUE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb102', 'INTERNAL_STORAGE', 'Unutrašnja memorija', 'int', 'GB', TRUE, TRUE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb103', 'SCREEN_SIZE', 'Dijagonala ekrana', 'float', 'inča', TRUE, FALSE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb104', 'CPU', 'Procesor', 'string', NULL, TRUE, FALSE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb105', 'GPU', 'Grafička kartica', 'string', NULL, TRUE, FALSE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb106', 'OS', 'Operativni sistem', 'string', NULL, TRUE, FALSE),
    -- Для молока
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb201', 'FAT_PERCENT', 'Procenat masti', 'float', '%', TRUE, TRUE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb202', 'VOLUME_L', 'Zapremina', 'float', 'l', TRUE, TRUE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb203', 'MILK_TYPE', 'Vrsta mleka', 'string', NULL, TRUE, FALSE),
    -- Для обуви
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb301', 'SIZE_EU', 'Veličina (EU)', 'float', NULL, TRUE, TRUE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb302', 'GENDER', 'Pol', 'string', NULL, TRUE, FALSE),
    -- Для мебели
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb401', 'WIDTH_CM', 'Širina', 'float', 'cm', TRUE, FALSE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb402', 'HEIGHT_CM', 'Visina', 'float', 'cm', TRUE, FALSE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb403', 'DEPTH_CM', 'Dubina', 'float', 'cm', TRUE, FALSE),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb404', 'MATERIAL', 'Materijal', 'string', NULL, TRUE, FALSE)
ON CONFLICT (code) DO NOTHING;

-- 6. product_type ↔ attributes
-- SMARTPHONE
INSERT INTO product_type_attributes (product_type_id, attribute_id, is_required, sort_order)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb001', FALSE, 10),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb101', TRUE, 20),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb102', TRUE, 30),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb103', FALSE, 40),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb106', FALSE, 50)
ON CONFLICT DO NOTHING;

-- LAPTOP
INSERT INTO product_type_attributes (product_type_id, attribute_id, is_required, sort_order)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb001', FALSE, 10),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb101', TRUE, 20),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb102', TRUE, 30),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb103', FALSE, 40),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb104', FALSE, 50),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb105', FALSE, 60)
ON CONFLICT DO NOTHING;

-- TV
INSERT INTO product_type_attributes (product_type_id, attribute_id, is_required, sort_order)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb001', FALSE, 10),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb103', TRUE, 20)
ON CONFLICT DO NOTHING;

-- MILK
INSERT INTO product_type_attributes (product_type_id, attribute_id, is_required, sort_order)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb201', TRUE, 10),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb202', TRUE, 20),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb203', FALSE, 30)
ON CONFLICT DO NOTHING;

-- SNEAKERS
INSERT INTO product_type_attributes (product_type_id, attribute_id, is_required, sort_order)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa5', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb001', FALSE, 10),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa5', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb301', TRUE, 20),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa5', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb002', FALSE, 30),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa5', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb302', FALSE, 40)
ON CONFLICT DO NOTHING;

-- FURNITURE
INSERT INTO product_type_attributes (product_type_id, attribute_id, is_required, sort_order)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa6', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb404', FALSE, 10),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa6', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb401', FALSE, 20),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa6', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb402', FALSE, 30),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa6', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbb403', FALSE, 40)
ON CONFLICT DO NOTHING;

-- 7. Города Сербии
INSERT INTO cities (id, slug, name_sr, region_sr, sort_order)
VALUES
    ('99999999-0000-0000-0000-000000000001', 'beograd', 'Beograd', 'Beogradski okrug', 10),
    ('99999999-0000-0000-0000-000000000002', 'novi-sad', 'Novi Sad', 'Južno-bački okrug', 20),
    ('99999999-0000-0000-0000-000000000003', 'nis', 'Niš', 'Nišavski okrug', 30),
    ('99999999-0000-0000-0000-000000000004', 'kragujevac', 'Kragujevac', 'Šumadijski okrug', 40),
    ('99999999-0000-0000-0000-000000000005', 'subotica', 'Subotica', 'Severno-bački okrug', 50)
ON CONFLICT (slug) DO NOTHING;


