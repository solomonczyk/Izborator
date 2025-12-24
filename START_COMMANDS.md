# ▶️ КОМАНДЫ ЗАПУСКА ПРОЕКТА

## Быстрый старт (5 минут)

### 1. Клонировать репо
```bash
git clone https://github.com/your-username/Izborator.git
cd Izborator
```

### 2. Скопировать конфиги
```bash
cp backend/.env.example backend/.env
cp frontend/.env.local.example frontend/.env.local
```

### 3. Запустить сервисы (через Docker Compose)
```bash
docker-compose up -d
```

### 4. Инициализировать БД
```bash
# Запустить миграции
docker-compose exec backend go run cmd/api/main.go -migrate

# Или вручную через контейнер
docker-compose exec postgres psql -U izborator -d izborator -f migrations/001_init.sql
```

### 5. Проверить статус
```bash
# API доступен на http://localhost:3002
curl http://localhost:3002/api/v1/health

# Frontend доступен на http://localhost:3000
open http://localhost:3000

# Meilisearch доступен на http://localhost:7700
open http://localhost:7700
```

---

## Запуск отдельных компонентов

### Backend API

```bash
# Локально (требует Go 1.21+)
cd backend
go run cmd/api/main.go

# Через Docker
docker-compose run --rm backend go run cmd/api/main.go

# На специфичном порту
cd backend
go run cmd/api/main.go -port 8080
```

### Backend Worker (Scraper/Processor)

```bash
# Запустить worker с обработкой
cd backend
go run cmd/worker/main.go -process

# Только сбор статистики
cd backend
go run cmd/worker/main.go -stats

# С определенным процессом
docker-compose run --rm backend ./cmd/worker/main.go -process -shops=1,2,3
```

### Backend Indexer (Meilisearch)

```bash
# Запустить индексацию
cd backend
go run cmd/indexer/main.go

# Через Docker
docker-compose run --rm backend ./indexer
```

### Frontend (Next.js)

```bash
# Development mode (с hot reload)
cd frontend
npm install
npm run dev
# Открыть http://localhost:3000

# Production build
cd frontend
npm run build
npm start

# Или через Docker
docker-compose run --rm frontend npm run dev
```

---

## Запуск сервисов инфраструктуры

### PostgreSQL

```bash
# Запустить контейнер
docker-compose up -d postgres

# Подключиться к БД
docker-compose exec postgres psql -U izborator -d izborator

# Просмотр логов
docker-compose logs -f postgres
```

### Redis

```bash
# Запустить контейнер
docker-compose up -d redis

# Проверить статус
docker-compose exec redis redis-cli ping
# Результат: PONG

# Посмотреть данные в Redis
docker-compose exec redis redis-cli
> KEYS *
> GET key-name
```

### Meilisearch

```bash
# Запустить контейнер
docker-compose up -d meilisearch

# Открыть интерфейс
open http://localhost:7700

# Проверить индексы через API
curl http://localhost:7700/indexes
```

---

## Полезные скрипты

### Проверка здоровья системы

```bash
# Все сразу
./scripts/check/health.sh

# Отдельные компоненты
./scripts/check/db-status.sh
./scripts/check/services.sh
./scripts/check/migrations.sh
```

### Запуск тестов

```bash
# Все backend тесты
cd backend && go test ./... -v

# С покрытием
cd backend && go test ./... -cover

# Определенный пакет
cd backend && go test ./internal/products -v

# Frontend тесты (если настроены)
cd frontend && npm test

# E2E тесты
cd frontend && npm run test:e2e
```

### Логирование и отладка

```bash
# Просмотр логов всех контейнеров
docker-compose logs -f

# Определенного контейнера
docker-compose logs -f backend
docker-compose logs -f postgres

# В реальном времени
docker-compose logs -f --tail=100

# Сохранить логи в файл
docker-compose logs > logs/docker.log 2>&1
```

### Очистка и перезагрузка

```bash
# Остановить все контейнеры
docker-compose down

# Остановить и удалить данные (⚠️ ВНИМАНИЕ!)
docker-compose down -v

# Перестартовать все
docker-compose restart

# Перестартовать определенный сервис
docker-compose restart backend
```

---

## Переменные окружения

### Для локальной разработки

```bash
# backend/.env
API_HOST=0.0.0.0
API_PORT=3002
DB_HOST=localhost
DB_PORT=5432
DB_USER=izborator
DB_PASSWORD=password
REDIS_HOST=localhost
REDIS_PORT=6379
MEILISEARCH_HOST=http://localhost:7700
LOG_LEVEL=debug
```

```bash
# frontend/.env.local
NEXT_PUBLIC_API_BASE=http://localhost:3002
NEXT_PUBLIC_ENV=development
```

---

## Решение проблем

### Backend не запускается

```bash
# Проверить логи
docker-compose logs backend

# Проверить что порт свободен
lsof -i :3002

# Проверить что БД доступна
docker-compose exec postgres pg_isready
```

### Frontend не запускается

```bash
# Очистить кэш Node
rm -rf frontend/node_modules frontend/package-lock.json
npm install

# Проверить переменные окружения
cat frontend/.env.local

# Запустить в debug режиме
cd frontend && npm run dev -- --debug
```

### БД недоступна

```bash
# Проверить что контейнер запущен
docker-compose ps postgres

# Проверить подключение
docker-compose exec postgres psql -U izborator -d izborator -c "SELECT 1"

# Просмотреть логи БД
docker-compose logs postgres
```

---

## Полезные ссылки

- **API Documentation:** http://localhost:3002/swagger
- **Meilisearch Admin:** http://localhost:7700
- **Frontend:** http://localhost:3000
- **Redis CLI:** `docker-compose exec redis redis-cli`
- **PostgreSQL:** `docker-compose exec postgres psql`

---

**Последнее обновление:** 24 декабря 2025
