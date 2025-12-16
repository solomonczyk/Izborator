-- Обновление URL изображений на рабочий сервис placehold.co
UPDATE products 
SET image_url = REPLACE(image_url, 'via.placeholder.com', 'placehold.co')
WHERE image_url LIKE '%via.placeholder.com%';

