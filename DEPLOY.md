# 🚀 Инструкция по деплою Izborator

## Подготовка сервера (VPS)

### Требования:
- Ubuntu 22.04+ или Debian 12+
- Минимум 2GB RAM, 20GB SSD
- Docker и Docker Compose установлены

### Установка Docker (если не установлен):

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Установка Docker Compose
sudo apt install docker-compose-plugin -y

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER
newgrp docker
```

## Деплой проекта

### 1. Клонирование репозитория

```bash
git clone git@github.com:solomonczyk/Izborator.git
cd Izborator
```

### 2. Настройка переменных окружения

Создай файл `.env` в корне проекта:

```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=izborator

# Meilisearch
MEILISEARCH_HOST=meilisearch
MEILISEARCH_PORT=7700
MEILISEARCH_API_KEY=your_secure_master_key

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Server
SERVER_PORT=8080
LOG_LEVEL=info
```

### 3. Обновление docker-compose.yml

Обнови `docker-compose.yml`, добавив сервисы для backend и frontend:

```yaml
services:
  # ... существующие сервисы (postgres, meilisearch, redis, influxdb) ...

  # Backend API
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: izborator_backend
    restart: always
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - MEILISEARCH_HOST=meilisearch
      - MEILISEARCH_PORT=7700
      - MEILISEARCH_API_KEY=${MEILISEARCH_API_KEY}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - SERVER_PORT=8080
      - LOG_LEVEL=info
    depends_on:
      postgres:
        condition: service_healthy
      meilisearch:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - izborator_network

  # Worker (Daemon)
  worker:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: izborator_worker
    restart: always
    command: ["./worker", "-daemon"]
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - MEILISEARCH_HOST=meilisearch
      - MEILISEARCH_PORT=7700
      - MEILISEARCH_API_KEY=${MEILISEARCH_API_KEY}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - LOG_LEVEL=info
    depends_on:
      - backend
    networks:
      - izborator_network

  # Frontend
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: izborator_frontend
    restart: always
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_BASE=http://localhost:8080
      - NODE_ENV=production
    depends_on:
      - backend
    networks:
      - izborator_network

networks:
  izborator_network:
    driver: bridge
```

### 4. Запуск миграций

```bash
# Запуск миграций через Docker
docker-compose run --rm backend ./migrate up
```

### 5. Запуск всех сервисов

```bash
# Сборка и запуск всех контейнеров
docker-compose up -d --build

# Проверка статуса
docker-compose ps

# Просмотр логов
docker-compose logs -f
```

### 6. Настройка Nginx (опционально)

Если хочешь использовать домен и HTTPS:

```nginx
# /etc/nginx/sites-available/izborator
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Обновление после изменений

```bash
# Остановка контейнеров
docker-compose down

# Обновление кода
git pull

# Пересборка и запуск
docker-compose up -d --build
```

## Мониторинг

```bash
# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f backend
docker-compose logs -f worker
docker-compose logs -f frontend

# Использование ресурсов
docker stats
```

## Резервное копирование

```bash
# Бэкап PostgreSQL
docker exec izborator_postgres pg_dump -U postgres izborator > backup.sql

# Восстановление
docker exec -i izborator_postgres psql -U postgres izborator < backup.sql
```

## Troubleshooting

### Проблема: Контейнеры не запускаются
```bash
# Проверь логи
docker-compose logs

# Проверь, что порты свободны
sudo netstat -tulpn | grep -E '8080|3000|5432|7700'
```

### Проблема: Backend не подключается к БД
- Проверь переменные окружения в `.env`
- Убедись, что `DB_HOST=postgres` (имя сервиса в docker-compose)
- Проверь, что postgres запущен: `docker-compose ps postgres`

### Проблема: Frontend не видит Backend
- Проверь `NEXT_PUBLIC_API_BASE` в окружении frontend
- Убедись, что backend доступен: `curl http://localhost:8080/api/health`

