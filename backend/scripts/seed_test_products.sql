-- Тестовые товары для демонстрации каталога
-- Эти товары будут отображаться в каталоге без парсинга

-- Сначала убедимся, что категории существуют
-- Используем существующие ID категорий из seed_data.sql

-- 1. Товар в категории "Sport i rekreacija" (ID: 55555555-5555-5555-5555-555555555555)
INSERT INTO products (id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at)
VALUES (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'Nike Air Max 90',
    'Классические кроссовки Nike Air Max 90 с амортизацией Air',
    'Nike',
    'Sport i rekreacija',
    '55555555-5555-5555-5555-555555555555',
    'https://via.placeholder.com/400x400?text=Nike+Air+Max+90',
    '{"size": "42", "color": "Black/White", "material": "Leather"}'::jsonb,
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET updated_at = NOW();

-- 2. Товар в категории "Mobilni telefoni" (ID: 11111111-1111-1111-1111-111111111201)
INSERT INTO products (id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at)
VALUES (
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'Samsung Galaxy S24',
    'Флагманский смартфон Samsung с камерой 200MP',
    'Samsung',
    'Mobilni telefoni',
    '11111111-1111-1111-1111-111111111201',
    'https://via.placeholder.com/400x400?text=Samsung+Galaxy+S24',
    '{"storage": "256GB", "ram": "8GB", "screen": "6.2 inch", "camera": "200MP"}'::jsonb,
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET updated_at = NOW();

-- 3. Товар в категории "Laptopovi" (ID: 11111111-1111-1111-1111-111111111202)
INSERT INTO products (id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at)
VALUES (
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    'Lenovo IdeaPad 3',
    'Ноутбук Lenovo IdeaPad 3 с процессором AMD Ryzen 5',
    'Lenovo',
    'Laptopovi',
    '11111111-1111-1111-1111-111111111202',
    'https://via.placeholder.com/400x400?text=Lenovo+IdeaPad+3',
    '{"cpu": "AMD Ryzen 5", "ram": "8GB", "storage": "512GB SSD", "screen": "15.6 inch"}'::jsonb,
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET updated_at = NOW();

-- 4. Товар в категории "Televizori" (ID: 11111111-1111-1111-1111-111111111203)
INSERT INTO products (id, name, description, brand, category, category_id, image_url, specs, created_at, updated_at)
VALUES (
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'Samsung 55" QLED TV',
    'Телевизор Samsung 55 дюймов с технологией QLED',
    'Samsung',
    'Televizori',
    '11111111-1111-1111-1111-111111111203',
    'https://via.placeholder.com/400x400?text=Samsung+55+QLED',
    '{"size": "55 inch", "resolution": "4K UHD", "smart_tv": "Yes", "hdr": "Yes"}'::jsonb,
    NOW(),
    NOW()
)
ON CONFLICT (id) DO UPDATE SET updated_at = NOW();

-- Добавляем цены для товаров (Gigatron)
-- Магазин ID: a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11

-- Цена для Nike Air Max 90
INSERT INTO product_prices (id, product_id, shop_id, price, currency, url, in_stock, city_id)
VALUES (
    gen_random_uuid(),
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    12990.00,
    'RSD',
    'https://gigatron.rs/sport/nike-air-max-90',
    true,
    NULL
)
ON CONFLICT (product_id, shop_id, city_id) DO UPDATE SET price = EXCLUDED.price;

-- Цена для Samsung Galaxy S24
INSERT INTO product_prices (id, product_id, shop_id, price, currency, url, in_stock, city_id)
VALUES (
    gen_random_uuid(),
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    89990.00,
    'RSD',
    'https://gigatron.rs/mobilni-telefoni/samsung-galaxy-s24',
    true,
    NULL
)
ON CONFLICT (product_id, shop_id, city_id) DO UPDATE SET price = EXCLUDED.price;

-- Цена для Lenovo IdeaPad 3
INSERT INTO product_prices (id, product_id, shop_id, price, currency, url, in_stock, city_id)
VALUES (
    gen_random_uuid(),
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    59990.00,
    'RSD',
    'https://gigatron.rs/laptopovi/lenovo-ideapad-3',
    true,
    NULL
)
ON CONFLICT (product_id, shop_id, city_id) DO UPDATE SET price = EXCLUDED.price;

-- Цена для Samsung 55" QLED TV
INSERT INTO product_prices (id, product_id, shop_id, price, currency, url, in_stock, city_id)
VALUES (
    gen_random_uuid(),
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    149990.00,
    'RSD',
    'https://gigatron.rs/televizori/samsung-55-qled',
    true,
    NULL
)
ON CONFLICT (product_id, shop_id, city_id) DO UPDATE SET price = EXCLUDED.price;

