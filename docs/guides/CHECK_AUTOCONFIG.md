# 🔍 Проверка результатов AutoConfig

## Шаг 1: Проверка базы данных

Выполни на сервере:

```bash
# Проверка автоматически созданных магазинов
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    id,
    name, 
    base_url, 
    is_active, 
    is_auto_configured,
    created_at 
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC;
"

# Статистика по магазинам
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    COUNT(*) FILTER (WHERE is_auto_configured = true) as auto_configured_count,
    COUNT(*) FILTER (WHERE is_auto_configured = false) as manual_count,
    COUNT(*) as total_shops
FROM shops;
"

# Проверка shop_config_attempts
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count
FROM shop_config_attempts
GROUP BY status
ORDER BY status;
"
```

**Или используй готовый скрипт:**

```bash
chmod +x check-autoconfig-results.sh
./check-autoconfig-results.sh
```

## Шаг 2: Проверка конфигурации .env

### Проверка docker-compose.yml

✅ **Уже исправлено:** В `docker-compose.yml` уже есть правильная конфигурация:

```yaml
services:
  backend:
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}  # ✅ Правильно
      
  worker:
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}  # ✅ Правильно
```

### Проверка .env файла на сервере

Убедись, что на сервере есть `.env` файл в корне проекта с:

```bash
# На сервере
cd ~/Izborator
cat .env | grep OPENAI_API_KEY
```

Должно быть:
```
OPENAI_API_KEY=sk-proj-... (твой реальный ключ)
```

### Если .env отсутствует или неполный

1. Скопируй `.env.example` в `.env`:
   ```bash
   cp .env.example .env
   ```

2. Отредактируй `.env` и добавь реальные значения:
   ```bash
   nano .env
   # Или
   vi .env
   ```

3. Убедись, что все переменные заполнены:
   - `OPENAI_API_KEY` - обязательно
   - `DB_PASSWORD` - обязательно
   - `MEILISEARCH_API_KEY` - обязательно
   - И другие секреты

## Шаг 3: Пересоздание контейнеров (если нужно)

Если нужно применить изменения в `.env`:

```bash
# Пересоздать контейнеры с новыми переменными
docker-compose up -d --force-recreate backend worker

# Проверить логи
docker-compose logs backend | tail -20
docker-compose logs worker | tail -20
```

## Шаг 4: Проверка работы воркера в фоне

```bash
# Проверить, что worker запущен
docker-compose ps worker

# Проверить логи worker
docker-compose logs worker --tail=50

# Если worker не запущен, запустить
docker-compose up -d worker
```

## ✅ Ожидаемые результаты

После успешного AutoConfig:

1. **В таблице `shops`:**
   - Должны быть записи с `is_auto_configured = true`
   - `is_active = true` (если конфиг успешно создан)
   - `base_url` заполнен
   - `name` заполнен

2. **В таблице `shop_config_attempts`:**
   - Записи со статусом `success` для успешных попыток
   - Записи со статусом `failed` для неудачных (если были)

3. **В логах:**
   - `✨ SUCCESS! Config generated` для успешных магазинов
   - Информация о созданных селекторах

## 🔧 Если магазины не созданы

1. Проверь логи AutoConfig:
   ```bash
   docker-compose logs backend | grep -i autoconfig
   ```

2. Проверь наличие classified кандидатов:
   ```bash
   docker exec -i izborator_postgres psql -U postgres -d izborator -c "
   SELECT status, COUNT(*) 
   FROM potential_shops 
   GROUP BY status;
   "
   ```

3. Если нет classified кандидатов - запусти Classifier:
   ```bash
   docker-compose run --rm backend ./classifier -classify-all
   ```

4. Затем запусти AutoConfig:
   ```bash
   docker-compose run --rm \
     -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
     backend ./autoconfig -limit 5
   ```

