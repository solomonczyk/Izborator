-- Тестовые данные для Izborator
-- Выполните: docker exec -i izborator_postgres psql -U postgres -d izborator < backend/scripts/seed_test_data.sql

-- Вставляем тестовые товары
INSERT INTO products (id, name, description, brand, category, image_url, specs)
VALUES 
    (
        '11111111-1111-1111-1111-111111111111'::UUID,
        'Motorola G72 8/256GB Gray',
        'Смартфон Motorola G72 с 8GB RAM и 256GB памяти',
        'Motorola',
        'Смартфоны',
        'https://gigatron.rs/media/catalog/product/cache/image/700x700/9df78eab33525d08d6e5fb8d27136e95/m/o/motorola-g72-8-256-gray.jpg',
        '{"RAM": "8GB", "Storage": "256GB", "Color": "Gray"}'::JSONB
    ),
    (
        '22222222-2222-2222-2222-222222222222'::UUID,
        'Samsung Galaxy A54 128GB Black',
        'Смартфон Samsung Galaxy A54 с 128GB памяти',
        'Samsung',
        'Смартфоны',
        'https://gigatron.rs/media/catalog/product/cache/image/700x700/9df78eab33525d08d6e5fb8d27136e95/s/a/samsung-galaxy-a54.jpg',
        '{"RAM": "6GB", "Storage": "128GB", "Color": "Black"}'::JSONB
    ),
    (
        '33333333-3333-3333-3333-333333333333'::UUID,
        'iPhone 15 Pro 256GB Natural Titanium',
        'Смартфон Apple iPhone 15 Pro с 256GB памяти',
        'Apple',
        'Смартфоны',
        'https://gigatron.rs/media/catalog/product/cache/image/700x700/9df78eab33525d08d6e5fb8d27136e95/i/p/iphone-15-pro.jpg',
        '{"RAM": "8GB", "Storage": "256GB", "Color": "Natural Titanium"}'::JSONB
    )
ON CONFLICT (id) DO NOTHING;

-- Вставляем тестовые цены
INSERT INTO product_prices (product_id, shop_id, shop_name, price, currency, url, in_stock)
VALUES 
    (
        '11111111-1111-1111-1111-111111111111'::UUID,
        '550e8400-e29b-41d4-a716-446655440000'::UUID,
        'Gigatron',
        29999.00,
        'RSD',
        'https://gigatron.rs/proizvod/motorola-g72-8-256-gray-840023252556',
        TRUE
    ),
    (
        '22222222-2222-2222-2222-222222222222'::UUID,
        '550e8400-e29b-41d4-a716-446655440000'::UUID,
        'Gigatron',
        34999.00,
        'RSD',
        'https://gigatron.rs/proizvod/samsung-galaxy-a54-128gb-black',
        TRUE
    ),
    (
        '33333333-3333-3333-3333-333333333333'::UUID,
        '550e8400-e29b-41d4-a716-446655440000'::UUID,
        'Gigatron',
        129999.00,
        'RSD',
        'https://gigatron.rs/proizvod/iphone-15-pro-256gb-natural-titanium',
        TRUE
    )
ON CONFLICT (product_id, shop_id) DO UPDATE SET
    price = EXCLUDED.price,
    updated_at = NOW();

-- Выводим результат
SELECT 'Тестовые данные добавлены!' AS status;
SELECT COUNT(*) AS products_count FROM products;
SELECT COUNT(*) AS prices_count FROM product_prices;

