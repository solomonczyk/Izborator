#!/bin/bash
# Скрипт для автоматического деплоя Izborator на VPS

set -e  # Остановка при ошибке

echo "🚀 Начинаем деплой Izborator..."

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Проверка Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker не установлен. Устанавливаю...${NC}"
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    newgrp docker
    echo -e "${GREEN}✅ Docker установлен${NC}"
else
    echo -e "${GREEN}✅ Docker уже установлен${NC}"
fi

# Проверка Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose не установлен. Устанавливаю...${NC}"
    sudo apt update
    sudo apt install docker-compose-plugin -y
    echo -e "${GREEN}✅ Docker Compose установлен${NC}"
else
    echo -e "${GREEN}✅ Docker Compose уже установлен${NC}"
fi

# Проверка архитектуры
ARCH=$(uname -m)
echo -e "${YELLOW}📋 Архитектура: $ARCH${NC}"

# Клонирование репозитория (если еще не клонирован)
if [ ! -d "Izborator" ]; then
    echo -e "${YELLOW}📥 Клонирую репозиторий...${NC}"
    git clone git@github.com:solomonczyk/Izborator.git
    cd Izborator
else
    echo -e "${YELLOW}📥 Обновляю репозиторий...${NC}"
    cd Izborator
    git pull
fi

# Создание .env файла (если не существует)
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}📝 Создаю .env файл...${NC}"
    cat > .env << EOF
# Database
DB_USER=postgres
DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
DB_NAME=izborator

# Meilisearch
MEILISEARCH_API_KEY=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)

# Server
LOG_LEVEL=info
EOF
    echo -e "${GREEN}✅ .env файл создан${NC}"
    echo -e "${YELLOW}⚠️  ВАЖНО: Сохрани пароли из .env файла!${NC}"
else
    echo -e "${GREEN}✅ .env файл уже существует${NC}"
fi

# Остановка старых контейнеров (если есть)
echo -e "${YELLOW}🛑 Останавливаю старые контейнеры...${NC}"
docker-compose down 2>/dev/null || true

# Запуск миграций
echo -e "${YELLOW}🗄️  Запускаю миграции...${NC}"
docker-compose run --rm backend ./migrate up || echo -e "${YELLOW}⚠️  Миграции пропущены (возможно, уже выполнены)${NC}"

# Сборка и запуск контейнеров
echo -e "${YELLOW}🔨 Собираю Docker образы...${NC}"
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    echo -e "${YELLOW}📋 Обнаружена ARM64 архитектура, использую buildx...${NC}"
    docker buildx create --use --name multiarch 2>/dev/null || docker buildx use multiarch
    docker-compose build --platform linux/arm64
else
    docker-compose build
fi

echo -e "${YELLOW}🚀 Запускаю контейнеры...${NC}"
docker-compose up -d

# Ожидание готовности сервисов
echo -e "${YELLOW}⏳ Ожидаю готовности сервисов (30 секунд)...${NC}"
sleep 30

# Проверка статуса
echo -e "${YELLOW}📊 Статус контейнеров:${NC}"
docker-compose ps

# Проверка здоровья API
echo -e "${YELLOW}🏥 Проверяю здоровье API...${NC}"
sleep 5
if curl -f http://localhost:8080/api/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Backend API работает!${NC}"
else
    echo -e "${RED}⚠️  Backend API еще не готов, проверь логи: docker-compose logs backend${NC}"
fi

# Проверка Frontend
if curl -f http://localhost:3000 > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Frontend работает!${NC}"
else
    echo -e "${RED}⚠️  Frontend еще не готов, проверь логи: docker-compose logs frontend${NC}"
fi

echo -e "${GREEN}🎉 Деплой завершен!${NC}"
echo -e "${YELLOW}📋 Полезные команды:${NC}"
echo -e "  Логи: ${GREEN}docker-compose logs -f${NC}"
echo -e "  Статус: ${GREEN}docker-compose ps${NC}"
echo -e "  Остановка: ${GREEN}docker-compose down${NC}"
echo -e "  Перезапуск: ${GREEN}docker-compose restart${NC}"

