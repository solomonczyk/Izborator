-- Быстрое исправление: создание таблицы cities если её нет
-- Это временное решение, лучше применить полную миграцию 0005

-- Проверяем и создаем таблицу cities если её нет
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'cities') THEN
        CREATE TABLE cities (
            id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            slug        TEXT NOT NULL UNIQUE,
            name_sr     TEXT NOT NULL,
            region_sr   TEXT NULL,
            sort_order  INTEGER NOT NULL DEFAULT 100,
            is_active   BOOLEAN NOT NULL DEFAULT TRUE,
            created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
        
        CREATE INDEX IF NOT EXISTS idx_cities_slug ON cities(slug);
        CREATE INDEX IF NOT EXISTS idx_cities_is_active ON cities(is_active);
        
        RAISE NOTICE 'Table cities created';
    ELSE
        RAISE NOTICE 'Table cities already exists';
    END IF;
END $$;

-- Добавляем несколько тестовых городов если таблица пустая
INSERT INTO cities (slug, name_sr, region_sr, sort_order, is_active)
VALUES
    ('beograd', 'Beograd', 'Grad Beograd', 10, true),
    ('novi-sad', 'Novi Sad', 'Južna Bačka', 20, true),
    ('nis', 'Niš', 'Nišavski okrug', 30, true)
ON CONFLICT (slug) DO NOTHING;


