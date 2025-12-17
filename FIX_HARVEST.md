# 🔧 Исправление проблем Harvest

## Проблемы:
1. ❌ Миграции не применены (таблица `shop_config_attempts` не существует)
2. ❌ Classifier не обработал 85 кандидатов (все еще статус "new")
3. ❌ AutoConfig не запустился (нет classified кандидатов)

## Решение:

### Шаг 1: Применить миграции

```bash
cd ~/Izborator
docker-compose run --rm backend ./migrate
```

**Ожидаемый результат:** "Migration up finished successfully"

### Шаг 2: Загрузить переменные окружения

```bash
export $(cat .env | grep -v '^#' | xargs)
```

### Шаг 3: Запустить Classifier

```bash
docker-compose run --rm backend ./classifier -classify-all
```

**Ожидаемый результат:** Логи классификации и обновление статусов

### Шаг 4: Проверить количество classified

```bash
docker-compose exec -T postgres psql -U postgres -d izborator -c "SELECT status, COUNT(*) FROM potential_shops GROUP BY status;"
```

**Ожидаемый результат:** Должны быть записи со статусом "classified"

### Шаг 5: Запустить AutoConfig

```bash
docker-compose run --rm \
  -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
  -e OPENAI_MODEL="gpt-4o-mini" \
  backend ./autoconfig -limit 5
```

### Шаг 6: Проверить результаты

```bash
bash check-harvest-results.sh
```

## Быстрый скрипт (все в одном):

```bash
cd ~/Izborator

# 1. Миграции
echo "📦 Применение миграций..."
docker-compose run --rm backend ./migrate

# 2. Переменные
export $(cat .env | grep -v '^#' | xargs)

# 3. Classifier
echo "🔍 Запуск Classifier..."
docker-compose run --rm backend ./classifier -classify-all

# 4. AutoConfig
echo "🤖 Запуск AutoConfig..."
docker-compose run --rm \
  -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
  -e OPENAI_MODEL="gpt-4o-mini" \
  backend ./autoconfig -limit 5

# 5. Проверка
echo "📊 Проверка результатов..."
bash check-harvest-results.sh
```

