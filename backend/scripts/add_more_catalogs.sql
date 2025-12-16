-- Добавление дополнительных каталогов для существующих магазинов
-- Можно добавить несколько категорий для каждого магазина

-- Gigatron - дополнительные категории
-- Примечание: Для нескольких категорий можно создать отдельные записи или использовать один catalog_url
-- Сейчас используем один catalog_url, но можно расширить логику для обхода нескольких категорий

-- Laptopovi (Ноутбуки)
UPDATE shops
SET selectors = jsonb_set(
    selectors,
    '{catalog_urls}',
    COALESCE(selectors->'catalog_urls', '[]'::jsonb) || '["https://gigatron.rs/laptop-racunari"]'::jsonb
)
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

-- Televizori (Телевизоры)
UPDATE shops
SET selectors = jsonb_set(
    selectors,
    '{catalog_urls}',
    COALESCE(selectors->'catalog_urls', '[]'::jsonb) || '["https://gigatron.rs/televizori"]'::jsonb
)
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

-- Tehnomanija - дополнительные категории
UPDATE shops
SET selectors = jsonb_set(
    selectors,
    '{catalog_urls}',
    COALESCE(selectors->'catalog_urls', '[]'::jsonb) || '["https://www.tehnomanija.rs/laptop-racunari"]'::jsonb
)
WHERE id = 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380b22';

-- Примечание: Текущая реализация ParseCatalog использует один catalog_url
-- Для поддержки нескольких категорий нужно будет расширить функцию

