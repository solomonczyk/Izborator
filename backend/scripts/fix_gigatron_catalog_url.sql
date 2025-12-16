-- Исправление catalog_url для Gigatron на правильный путь
UPDATE shops 
SET selectors = jsonb_set(
    selectors, 
    '{catalog_url}', 
    '"https://gigatron.rs/mobilni-telefoni-tableti-i-oprema/mobilni-telefoni"'::jsonb
) 
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

-- Проверка результата
SELECT 
    name, 
    selectors->>'catalog_url' as catalog_url 
FROM shops 
WHERE id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

