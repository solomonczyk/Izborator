-- 0016_add_category_translations.down.sql
-- Удаление полей переводов для категорий

ALTER TABLE categories
    DROP COLUMN IF EXISTS name_ru,
    DROP COLUMN IF EXISTS name_en,
    DROP COLUMN IF EXISTS name_hu,
    DROP COLUMN IF EXISTS name_zh;

