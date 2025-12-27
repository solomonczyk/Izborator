# Тестирование AutoConfig для табличных данных

## Что было улучшено

### 1. Промпт для AI (`backend/internal/ai/client.go`)
- ✅ Добавлены детальные инструкции для работы с табличными данными
- ✅ Уточнено, что селекторы должны извлекать МНОЖЕСТВО элементов
- ✅ Добавлены примеры селекторов для таблиц
- ✅ Поддержка div-based списков

### 2. Валидация селекторов (`backend/internal/autoconfig/service.go`)
- ✅ Добавлен параметр `siteType` в функцию `validateSelectors`
- ✅ Для `service_provider`: собираются ВСЕ элементы (не только первый)
- ✅ Проверка количества найденных элементов
- ✅ Проверка соотношения количества имен и цен
- ✅ Улучшено логирование

## Как протестировать

### Вариант 1: Через docker-compose (рекомендуется)

```bash
# 1. Проверить, есть ли классифицированные кандидаты
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT COUNT(*) as classified_count 
FROM potential_shops 
WHERE status = 'classified';
"

# 2. Если есть кандидаты, запустить AutoConfig на одном
docker-compose run --rm backend ./autoconfig -limit 1

# 3. Проверить результат
docker-compose exec -T postgres psql -U postgres -d izborator -c "
SELECT 
    name, 
    base_url, 
    selectors->>'name' as name_selector,
    selectors->>'price' as price_selector
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC 
LIMIT 1;
"
```

### Вариант 2: Через скрипт

```bash
# Полный тест-драйв (Discovery -> Classifier -> AutoConfig)
./scripts/test/test-autoconfig-chain.sh

# Детальная проверка результатов
./scripts/check/check-autoconfig-details.sh
```

### Вариант 3: Ручной запуск (если есть .env с OPENAI_API_KEY)

```bash
cd backend
go run cmd/autoconfig/main.go -limit 1
```

## Что проверить в логах

При запуске AutoConfig для `service_provider` должны быть логи:

```
✅ Validation successful
  site_type: "service_provider"
  names_count: 5  (должно быть > 1 для таблиц)
  prices_count: 5
  first_name: "Услуга 1"
  first_price: "1000 RSD"
```

Если найдена только одна услуга:
```
⚠️ Service provider found only one service, might not be a table
  names_count: 1
  prices_count: 1
```

## Ожидаемые результаты

### Для e-commerce (как раньше):
- Находит один товар
- Валидирует name и price
- Создает магазин с селекторами

### Для service_provider (новое):
- Находит МНОЖЕСТВО услуг (если таблица)
- Валидирует, что найдено несколько элементов
- Проверяет соотношение имен и цен
- Создает магазин с селекторами для таблиц

## Примеры селекторов для таблиц

AI должен генерировать селекторы типа:
```json
{
  "name": "table tbody tr td:first-child",
  "price": "table tbody tr td:last-child",
  "image": "",
  "description": ""
}
```

Или для div-based списков:
```json
{
  "name": "div.service-item .service-name",
  "price": "div.service-item .service-price",
  "image": "",
  "description": ""
}
```

## Troubleshooting

### Проблема: "no candidates available"
**Решение:** Нужно сначала запустить Discovery и Classifier:
```bash
docker-compose run --rm backend ./discovery
docker-compose run --rm backend ./classifier -classify-all -limit 10
```

### Проблема: "AI generation failed"
**Решение:** Проверить OPENAI_API_KEY в .env файле

### Проблема: "validation failed"
**Решение:** Проверить логи - возможно, селекторы не находят данные на странице

