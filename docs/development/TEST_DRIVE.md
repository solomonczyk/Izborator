# 🚀 Project Horizon - Финальный Тест-Драйв

## Быстрый старт

### Вариант 1: На сервере (рекомендуется)

```bash
# Подключись к серверу
ssh root@твой_сервер

# Перейди в директорию проекта
cd ~/Izborator

# Запусти тестовый скрипт
bash test-autoconfig-chain.sh
```

### Вариант 2: Вручную на сервере

```bash
# Шаг 1: Discovery (если нужно)
docker-compose run --rm backend ./discovery

# Шаг 2: Classifier
docker-compose run --rm backend ./classifier -classify-all -limit 10

# Шаг 3: AutoConfig (тест на 1 магазине)
docker-compose run --rm backend ./autoconfig -limit 1

# Шаг 4: Проверка результата
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    name, 
    base_url, 
    is_auto_configured,
    ai_config_model,
    selectors->>'name' as name_selector,
    selectors->>'price' as price_selector,
    created_at
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC 
LIMIT 1;
"
```

### Вариант 3: Локально (требует Docker Desktop)

1. **Запусти Docker Desktop**
2. **Запусти БД:**
   ```bash
   docker-compose up -d postgres
   ```
3. **Подожди 10 секунд** (пока БД запустится)
4. **Запусти команды из Варианта 2**

## Что искать в логах AutoConfig

### ✅ Успешный запуск:

```
🤖 Auto-configuring shop domain=example.rs
Found product page url=https://example.rs/product/123
Asking AI for selectors...
✨ SUCCESS! Config generated selectors=map[name:.product-title price:.price ...]
```

### ❌ Возможные ошибки:

**Scout failed:**
```
Scout failed domain=example.rs error=no product link found
```
→ Сайт не имеет очевидных ссылок на товары (нужна ручная настройка)

**AI generation failed:**
```
AI generation failed error=rate limit exceeded
```
→ Превышен лимит OpenAI API (подожди или проверь баланс)

**Validation failed:**
```
Validation failed error=name selector did not extract data
```
→ AI сгенерировал неправильные селекторы (можно попробовать еще раз)

## Проверка результата

После успешного запуска AutoConfig, проверь БД:

```sql
-- Последний созданный магазин
SELECT 
    name, 
    base_url, 
    is_auto_configured,
    ai_config_model,
    selectors
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC 
LIMIT 1;

-- Статистика по статусам
SELECT 
    status,
    COUNT(*) as count
FROM potential_shops
GROUP BY status;

-- Попытки конфигурации
SELECT 
    status,
    COUNT(*) as count,
    MAX(created_at) as last_attempt
FROM shop_config_attempts
GROUP BY status;
```

## Ожидаемый результат

После успешного теста ты должен увидеть:

1. ✅ Новую запись в таблице `shops` с `is_auto_configured = true`
2. ✅ Валидные селекторы в JSON формате:
   ```json
   {
     "name": ".product-title",
     "price": ".price",
     "image": "img.product-image",
     "description": ".product-description"
   }
   ```
3. ✅ Статус в `potential_shops` изменен на `configured`
4. ✅ Запись в `shop_config_attempts` со статусом `success`

## Troubleshooting

### БД не доступна
```bash
# Проверь статус контейнеров
docker-compose ps

# Перезапусти БД
docker-compose restart postgres
```

### OpenAI ключ не работает
```bash
# Проверь переменные окружения
docker-compose exec backend env | grep OPENAI

# Если нет - добавь в .env на сервере
echo "OPENAI_API_KEY=твой_ключ" >> .env
docker-compose restart backend
```

### Нет кандидатов для обработки
```bash
# Проверь количество кандидатов
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT status, COUNT(*) FROM potential_shops GROUP BY status;
"

# Если нет "classified" - запусти Classifier
docker-compose run --rm backend ./classifier -classify-all
```

