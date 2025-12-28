-- 0016_add_category_translations.up.sql
-- Добавление полей переводов для категорий

ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS name_ru TEXT NULL,
    ADD COLUMN IF NOT EXISTS name_en TEXT NULL,
    ADD COLUMN IF NOT EXISTS name_hu TEXT NULL,
    ADD COLUMN IF NOT EXISTS name_zh TEXT NULL;

-- Заполняем переводы для существующих категорий
UPDATE categories SET
    name_ru = CASE code
        WHEN 'ELEKTRONIKA' THEN 'Электроника'
        WHEN 'HRANA_PICE' THEN 'Еда и напитки'
        WHEN 'MODA' THEN 'Мода и обувь'
        WHEN 'DOM_BASTA' THEN 'Дом и сад'
        WHEN 'SPORT' THEN 'Спорт и отдых'
        WHEN 'AUTO_MOTO' THEN 'Авто-мото'
        WHEN 'PHONES' THEN 'Мобильные телефоны'
        WHEN 'LAPTOPS' THEN 'Ноутбуки'
        WHEN 'TVS' THEN 'Телевизоры'
        WHEN 'MILK_DAIRY' THEN 'Молоко и молочные продукты'
        WHEN 'SNEAKERS' THEN 'Кроссовки'
        WHEN 'FURNITURE' THEN 'Мебель'
        ELSE name_sr
    END,
    name_en = CASE code
        WHEN 'ELEKTRONIKA' THEN 'Electronics'
        WHEN 'HRANA_PICE' THEN 'Food & Drinks'
        WHEN 'MODA' THEN 'Fashion & Footwear'
        WHEN 'DOM_BASTA' THEN 'Home & Garden'
        WHEN 'SPORT' THEN 'Sports & Recreation'
        WHEN 'AUTO_MOTO' THEN 'Auto & Motorcycle'
        WHEN 'PHONES' THEN 'Mobile Phones'
        WHEN 'LAPTOPS' THEN 'Laptops'
        WHEN 'TVS' THEN 'TVs'
        WHEN 'MILK_DAIRY' THEN 'Milk & Dairy Products'
        WHEN 'SNEAKERS' THEN 'Sneakers'
        WHEN 'FURNITURE' THEN 'Furniture'
        ELSE name_sr
    END,
    name_hu = CASE code
        WHEN 'ELEKTRONIKA' THEN 'Elektronika'
        WHEN 'HRANA_PICE' THEN 'Étel és ital'
        WHEN 'MODA' THEN 'Divat és cipő'
        WHEN 'DOM_BASTA' THEN 'Otthon és kert'
        WHEN 'SPORT' THEN 'Sport és szabadidő'
        WHEN 'AUTO_MOTO' THEN 'Autó és motor'
        WHEN 'PHONES' THEN 'Mobiltelefonok'
        WHEN 'LAPTOPS' THEN 'Laptopok'
        WHEN 'TVS' THEN 'TV-k'
        WHEN 'MILK_DAIRY' THEN 'Tej és tejtermékek'
        WHEN 'SNEAKERS' THEN 'Cipők'
        WHEN 'FURNITURE' THEN 'Bútor'
        ELSE name_sr
    END,
    name_zh = CASE code
        WHEN 'ELEKTRONIKA' THEN '电子产品'
        WHEN 'HRANA_PICE' THEN '食品和饮料'
        WHEN 'MODA' THEN '时尚和鞋类'
        WHEN 'DOM_BASTA' THEN '家居和花园'
        WHEN 'SPORT' THEN '运动和休闲'
        WHEN 'AUTO_MOTO' THEN '汽车和摩托车'
        WHEN 'PHONES' THEN '手机'
        WHEN 'LAPTOPS' THEN '笔记本电脑'
        WHEN 'TVS' THEN '电视'
        WHEN 'MILK_DAIRY' THEN '牛奶和乳制品'
        WHEN 'SNEAKERS' THEN '运动鞋'
        WHEN 'FURNITURE' THEN '家具'
        ELSE name_sr
    END
WHERE name_ru IS NULL OR name_en IS NULL OR name_hu IS NULL OR name_zh IS NULL;

