# Docker Compose для локального окружения

Этот файл поднимает все необходимые сервисы для разработки проекта Izborator.

## Сервисы

- **PostgreSQL** (порт 5432) — основная база данных
- **Meilisearch** (порт 7700) — поисковый движок
- **Redis** (порт 6379) — кеш и очереди
- **InfluxDB** (порт 8086) — история цен (time-series)

## Запуск

```bash
# Запустить все сервисы
docker-compose up -d

# Показать статус
docker-compose ps

# Показать логи
docker-compose logs -f

# Остановить все сервисы
docker-compose down

# Остановить и удалить volumes (очистить данные)
docker-compose down -v
```

## Настройка

Переменные окружения можно задать через `.env` файл или через переменные окружения системы.

Пример `.env` файла:

```env
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=izborator
MEILISEARCH_API_KEY=masterKey123
INFLUX_USER=admin
INFLUX_PASSWORD=adminpassword
INFLUX_ORG=izborator_org
INFLUX_BUCKET=price_history
```

## Применение миграций

После запуска контейнеров примени миграции:

```bash
cd backend
go run cmd/migrate/main.go -up
```

## Настройка индекса Meilisearch

После применения миграций настрой индекс:

```bash
cd backend
go run cmd/indexer/main.go -setup
go run cmd/indexer/main.go -sync
```

## Доступ к сервисам

- **PostgreSQL**: `localhost:5432`
- **Meilisearch**: `http://localhost:7700`
- **Redis**: `localhost:6379`
- **InfluxDB**: `http://localhost:8086`

## Health Checks

Все сервисы имеют health checks для автоматической проверки готовности.

