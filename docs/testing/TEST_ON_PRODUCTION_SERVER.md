# Тестирование AutoConfig на продакшен сервере

## Подключение к серверу

```bash
ssh root@152.53.227.37
# или
ssh root@v2202508292476370494.powersrv.de
```

## Шаг 1: Проверка данных

После подключения к серверу:

```bash
cd ~/Izborator
chmod +x scripts/production/check-autoconfig-on-server.sh
./scripts/production/check-autoconfig-on-server.sh
```

Скрипт покажет:
- ✅ Сколько классифицированных кандидатов есть
- ✅ Сколько из них service_provider (для тестирования таблиц)
- ✅ Какие магазины уже созданы через AutoConfig
- ✅ Рекомендации по следующим шагам

## Шаг 2: Если данных недостаточно

### Получить больше кандидатов:

```bash
cd ~/Izborator
docker compose run --rm backend ./discovery -max-results 200
```

### Классифицировать найденные сайты:

```bash
docker compose run --rm backend ./classifier -classify-all -limit 100
```

## Шаг 3: Тестирование AutoConfig

### Быстрый тест (3 кандидата):

```bash
cd ~/Izborator
chmod +x scripts/production/test-autoconfig-on-server.sh
./scripts/production/test-autoconfig-on-server.sh
```

Или вручную:

```bash
docker compose run --rm backend ./autoconfig -limit 3
```

### Что проверить в логах:

Для `service_provider` должны быть логи:
```
✅ Validation successful
  site_type: "service_provider"
  names_count: 5  (должно быть > 1 для таблиц)
  prices_count: 5
  first_name: "Стрижка мужская"
  first_price: "1500 RSD"
```

### Проверка результатов:

```bash
docker compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    name,
    base_url,
    selectors->>'name' as name_selector,
    selectors->>'price' as price_selector,
    created_at
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC 
LIMIT 5;
"
```

## Шаг 4: Проверка качества селекторов для таблиц

```bash
docker compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    name,
    base_url,
    selectors->>'name' as name_selector,
    selectors->>'price' as price_selector
FROM shops 
WHERE is_auto_configured = true
AND (
    selectors->>'name' LIKE '%table%' 
    OR selectors->>'name' LIKE '%tr%'
    OR selectors->>'name' LIKE '%td%'
)
ORDER BY created_at DESC;
"
```

## Troubleshooting

### Проблема: "no candidates available"
**Решение:** 
```bash
# Проверить наличие кандидатов
docker compose exec -T postgres psql -U postgres -d izborator -c "SELECT COUNT(*) FROM potential_shops WHERE status = 'classified';"

# Если нет - запустить Discovery и Classifier
docker compose run --rm backend ./discovery -max-results 200
docker compose run --rm backend ./classifier -classify-all -limit 100
```

### Проблема: "AI generation failed"
**Решение:** 
```bash
# Проверить OPENAI_API_KEY
docker compose exec backend env | grep OPENAI_API_KEY

# Проверить логи
docker compose logs backend | grep -i "ai\|openai"
```

### Проблема: "validation failed" для service_provider
**Решение:**
```bash
# Проверить логи детально
docker compose logs backend | grep -A 10 "validation"

# Возможно, селекторы не находят данные на странице
# Проверить, что на странице действительно есть таблица
```

## Автоматизация (Daemon режим)

Для непрерывной обработки:

```bash
# AutoConfig в режиме демона (обрабатывает кандидатов каждые 5 минут)
docker compose run -d backend ./autoconfig -daemon -interval 5m -limit 3
```

## Мониторинг

### Смотреть логи в реальном времени:

```bash
docker compose logs -f backend | grep -i "autoconfig\|validation"
```

### Статистика по созданным магазинам:

```bash
docker compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE selectors->>'name' IS NOT NULL) as with_name,
    COUNT(*) FILTER (WHERE selectors->>'price' IS NOT NULL) as with_price,
    COUNT(*) FILTER (
        WHERE selectors->>'name' IS NOT NULL 
        AND selectors->>'price' IS NOT NULL
    ) as with_both
FROM shops 
WHERE is_auto_configured = true;
"
```

