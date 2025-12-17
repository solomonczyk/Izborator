# 🔧 Быстрое исправление проблемы с переменными окружения

## Проблема

При запуске `docker-compose run --rm backend ./discovery` переменные окружения из `.env` не передаются автоматически.

## Решение

### Вариант 1: Использовать обновленный скрипт (рекомендуется)

```bash
cd ~/Izborator
git pull  # Получить обновленный скрипт
bash run-harvest.sh
```

### Вариант 2: Передать переменные вручную

```bash
# Discovery
docker-compose run --rm \
  -e GOOGLE_API_KEY="твой_google_api_key" \
  -e GOOGLE_CX="твой_cx_id" \
  backend ./discovery

# Classifier
docker-compose run --rm backend ./classifier -classify-all

# AutoConfig
docker-compose run --rm \
  -e OPENAI_API_KEY="твой_openai_key" \
  -e OPENAI_MODEL="gpt-4o-mini" \
  backend ./autoconfig -limit 5
```

### Вариант 3: Использовать env_file в docker-compose (постоянное решение)

Добавь в `docker-compose.yml` в секцию `backend`:

```yaml
backend:
  # ... существующие настройки
  env_file:
    - ./backend/.env
```

Затем перезапусти:
```bash
docker-compose up -d --build backend
```

### Вариант 4: Экспортировать переменные перед запуском

```bash
# Загрузи переменные из .env
export $(cat backend/.env | grep -v '^#' | xargs)

# Теперь запускай команды
docker-compose run --rm backend ./discovery
docker-compose run --rm backend ./classifier -classify-all
docker-compose run --rm backend ./autoconfig -limit 5
```

## Проверка

Убедись, что переменные доступны:

```bash
# Проверь, что .env файл существует
ls -la backend/.env

# Проверь содержимое (не показывай ключи публично!)
cat backend/.env | grep -E "GOOGLE_API_KEY|GOOGLE_CX|OPENAI_API_KEY"
```

