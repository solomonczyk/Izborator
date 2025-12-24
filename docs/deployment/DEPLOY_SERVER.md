# 🚀 Деплой на сервер v2202508292476370494

## Информация о сервере
- **Hostname:** v2202508292476370494.powersrv.de
- **IP:** 152.53.227.37
- **Архитектура:** ARM64
- **RAM:** 8GB
- **CPU:** 6 cores
- **Disk:** 256GB

## Шаг 1: Подключение к серверу

```bash
ssh root@152.53.227.37
# или
ssh root@v2202508292476370494.powersrv.de
```

## Шаг 2: Установка Docker (если не установлен)

```bash
# Обновление системы
apt update && apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Установка Docker Compose
apt install docker-compose-plugin -y

# Проверка
docker --version
docker compose version
```

## Шаг 3: Клонирование репозитория

```bash
# Переходим в домашнюю директорию
cd ~

# Клонируем репозиторий
git clone git@github.com:solomonczyk/Izborator.git
cd Izborator
```

**Если SSH ключ не настроен:**
```bash
# Используй HTTPS (потребуется токен GitHub)
git clone https://github.com/solomonczyk/Izborator.git
cd Izborator
```

## Шаг 4: Создание .env файла

```bash
# Создаем .env файл
cat > .env << 'EOF'
# Database
DB_USER=postgres
DB_PASSWORD=izborator_secure_password_2024
DB_NAME=izborator

# Meilisearch
MEILISEARCH_API_KEY=izborator_meili_master_key_2024

# Server
LOG_LEVEL=info
EOF
```

**⚠️ ВАЖНО:** Замени пароли на более безопасные!

## Шаг 5: Настройка для ARM64

Так как сервер на ARM64, нужно использовать buildx для сборки:

```bash
# Создаем buildx builder для multi-arch
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap
```

## Шаг 6: Запуск миграций

```bash
# Запускаем миграции
docker-compose run --rm backend ./migrate up
```

## Шаг 7: Сборка и запуск

```bash
# Собираем образы для ARM64
docker-compose build --platform linux/arm64

# Или используем автоматический скрипт
chmod +x deploy.sh
./deploy.sh
```

## Шаг 8: Запуск всех сервисов

```bash
# Запускаем все контейнеры
docker-compose up -d

# Проверяем статус
docker-compose ps

# Смотрим логи
docker-compose logs -f
```

## Шаг 9: Проверка работы

```bash
# Проверка Backend API
curl http://localhost:8080/api/health

# Проверка Frontend
curl http://localhost:3000
```

## Настройка Nginx (опционально)

Если хочешь использовать домен:

```bash
# Установка Nginx
apt install nginx -y

# Создание конфига
cat > /etc/nginx/sites-available/izborator << 'EOF'
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
EOF

# Активация
ln -s /etc/nginx/sites-available/izborator /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

## Полезные команды

```bash
# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f backend
docker-compose logs -f worker
docker-compose logs -f frontend

# Перезапуск сервиса
docker-compose restart backend

# Остановка всех сервисов
docker-compose down

# Обновление после изменений
git pull
docker-compose build --platform linux/arm64
docker-compose up -d
```

## Troubleshooting

### Проблема: Ошибка сборки на ARM64
```bash
# Используй buildx
docker buildx build --platform linux/arm64 -t izborator-backend ./backend
```

### Проблема: Контейнеры не запускаются
```bash
# Проверь логи
docker-compose logs

# Проверь, что порты свободны
netstat -tulpn | grep -E '8080|3000|5432|7700'
```

### Проблема: Backend не подключается к БД
```bash
# Проверь переменные окружения
docker-compose exec backend env | grep DB_

# Проверь, что postgres запущен
docker-compose ps postgres
```

## Автоматический деплой (скрипт)

Я создал скрипт `deploy.sh` для автоматического деплоя. Просто запусти:

```bash
chmod +x deploy.sh
./deploy.sh
```

Скрипт автоматически:
- Проверит и установит Docker
- Клонирует/обновит репозиторий
- Создаст .env файл
- Запустит миграции
- Соберет и запустит контейнеры
- Проверит здоровье сервисов

