-- Обновление URL изображений на рабочий сервис placehold.co
UPDATE products 
SET image_url = REPLACE(
    REPLACE(image_url, 'via.placeholder.com', 'placehold.co'),
    '/400x400?text=',
    '/400x400/000000/FFFFFF.png?text='
)
WHERE image_url LIKE '%via.placeholder.com%' OR image_url LIKE '%placehold.co%';

