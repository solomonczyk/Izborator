# 🛒 Добавление реальных URL для парсинга

## Проблема

Worker работает, но все URL из базы возвращают 404, потому что это тестовые URL, которых нет на реальном сайте.

## Решение: Добавить реальные рабочие URL

### Вариант 1: Через SQL (быстро)

```bash
ssh root@152.53.227.37
cd ~/Izborator

# Подключись к базе
docker exec -it izborator_postgres psql -U postgres -d izborator
```

Затем выполни SQL (замени URL на реальные с Gigatron):

```sql
-- Найди ID магазина Gigatron
SELECT id, name FROM shops WHERE name LIKE '%Gigatron%';

-- Добавь реальные URL (замени на рабочие ссылки с gigatron.rs)
-- Пример: найди реальный товар на gigatron.rs и скопируй его URL
INSERT INTO product_prices (product_id, shop_id, url, price, currency, updated_at)
SELECT 
    p.id,
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid, -- Gigatron ID
    'https://gigatron.rs/реальный-путь-к-товару',
    0, -- Цена будет обновлена при парсинге
    'RSD',
    NOW() - INTERVAL '25 hours' -- Сделаем "устаревшим", чтобы worker сразу его подхватил
FROM products p
WHERE p.name LIKE '%Samsung%' -- Или любой другой товар
LIMIT 1
ON CONFLICT (product_id, shop_id) DO UPDATE SET
    url = EXCLUDED.url,
    updated_at = EXCLUDED.updated_at;
```

### Вариант 2: Через Worker (ручной скрапинг)

```bash
ssh root@152.53.227.37
cd ~/Izborator

# Запусти ручной скрапинг реального URL
docker-compose exec worker ./worker \
  -url "https://gigatron.rs/реальный-путь-к-товару" \
  -shop "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
```

### Вариант 3: Обновить существующие URL

```sql
-- Обнови URL на реальные (примеры - замени на рабочие)
UPDATE product_prices 
SET url = 'https://gigatron.rs/реальный-путь',
    updated_at = NOW() - INTERVAL '25 hours'
WHERE shop_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
  AND url LIKE '%nike-air-max%'
LIMIT 1;
```

## Как найти реальные URL на Gigatron

1. Открой https://gigatron.rs
2. Найди любой товар (например, смартфон, ноутбук)
3. Скопируй полный URL из адресной строки
4. Используй его в SQL или для ручного скрапинга

## Проверка

После добавления URL:

```bash
# Проверь, что URL добавлены
docker exec izborator_postgres psql -U postgres -d izborator -c \
  "SELECT url, updated_at FROM product_prices WHERE url IS NOT NULL LIMIT 5;"

# Подожди ~10 минут (worker проверяет каждые 10 минут)
# Или перезапусти worker, чтобы он сразу проверил
docker-compose restart worker

# Следи за логами
docker-compose logs -f worker
```

## Ожидаемый результат

После добавления реальных URL, в логах должно появиться:

```
✅ Scraping successful
✅ Product parsed & saved
✅ Processed items: 1
```

