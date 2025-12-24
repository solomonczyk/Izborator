# 🏭 Project Horizon - Запуск "Фабрики" на Продакшене

## Цель

Наполнить базу реальными магазинами без единой строчки кода вручную, используя полный конвейер:
**Discovery → Classifier → AutoConfig**

## Подготовка

### ✅ Проверка перед запуском

1. **Dockerfile готов** - все бинарники (discovery, classifier, autoconfig) включены
2. **OpenAI ключ добавлен** - в `.env` на сервере есть `OPENAI_API_KEY`
3. **Google API ключи добавлены** - в `.env` есть `GOOGLE_API_KEY` и `GOOGLE_CX`
4. **Деплой завершен** - последний коммит задеплоен на сервер

## Запуск на сервере

### Вариант 1: Автоматический скрипт (рекомендуется)

```bash
# Подключись к серверу
ssh root@152.53.227.37

# Перейди в директорию проекта
cd ~/Izborator

# Запусти скрипт
bash run-harvest.sh
```

### Вариант 2: Вручную (пошагово)

```bash
# 1. Discovery - поиск новых доменов
docker-compose run --rm backend ./discovery

# 2. Classifier - классификация найденных доменов
docker-compose run --rm backend ./classifier -classify-all

# 3. AutoConfig - генерация селекторов для 5 магазинов
docker-compose run --rm backend ./autoconfig -limit 5
```

## Проверка результатов

После завершения AutoConfig, проверь созданные магазины:

```bash
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    name, 
    base_url, 
    is_active,
    is_auto_configured,
    ai_config_model,
    selectors->>'name' as name_selector,
    selectors->>'price' as price_selector,
    created_at
FROM shops 
WHERE is_auto_configured = true 
ORDER BY created_at DESC;
"
```

## Ожидаемый результат

После успешного запуска ты должен увидеть:

1. ✅ **Новые записи в `potential_shops`** со статусом `classified`
2. ✅ **Новые магазины в `shops`** с `is_auto_configured = true`
3. ✅ **Валидные селекторы** в JSON формате (name, price, image, description)
4. ✅ **Статус `configured`** в `potential_shops` для обработанных магазинов

## Статистика

Проверь общую статистику:

```bash
# Статистика по статусам
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count
FROM potential_shops
GROUP BY status
ORDER BY status;
"

# Попытки конфигурации
docker exec -i izborator_postgres psql -U postgres -d izborator -c "
SELECT 
    status,
    COUNT(*) as count,
    MAX(created_at) as last_attempt
FROM shop_config_attempts
GROUP BY status
ORDER BY status;
"
```

## Troubleshooting

### Discovery не находит домены
- Проверь `GOOGLE_API_KEY` и `GOOGLE_CX` в `.env`
- Проверь лимиты Google Custom Search API
- Попробуй запустить вручную: `docker-compose run --rm backend ./discovery`

### Classifier не классифицирует
- Проверь логи: `docker-compose logs backend`
- Убедись, что есть кандидаты со статусом `new`
- Попробуй запустить с лимитом: `docker-compose run --rm backend ./classifier -classify-all -limit 10`

### AutoConfig не генерирует конфиги
- Проверь `OPENAI_API_KEY` в `.env`
- Проверь баланс OpenAI API
- Проверь логи на ошибки Scout/Validation
- Попробуй запустить на 1 магазине: `docker-compose run --rm backend ./autoconfig -limit 1`

## Стоимость

- **Discovery**: ~$0.01 за 100 запросов (Google Custom Search API)
- **Classifier**: Бесплатно (локальная обработка)
- **AutoConfig**: ~$0.01 за магазин (OpenAI GPT-4o-mini)

**Итого**: ~$0.01-0.02 за успешную конфигурацию магазина.

## Следующие шаги

После успешного запуска "Фабрики":

1. ✅ Проверь созданные магазины в БД
2. ✅ Протестируй парсинг одного из магазинов
3. ✅ Настрой автоматический запуск (через cron или GitHub Actions)
4. ✅ Мониторь результаты и оптимизируй процесс

## Автоматизация

Для автоматического запуска "Фабрики" можно:

1. **GitHub Actions** - создать workflow с расписанием
2. **Cron на сервере** - запускать скрипт по расписанию
3. **Docker Compose** - добавить как отдельный сервис

Пример cron задачи (каждую неделю):
```bash
0 3 * * 1 cd ~/Izborator && bash run-harvest.sh >> /var/log/harvest.log 2>&1
```

