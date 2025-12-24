# 🚀 Development Setup Guide

## Quick Start для разработчиков

### 1. Клонирование и setup

```bash
git clone git@github.com:solomonczyk/izborator.git
cd Izborator

# Скопируйте .env файлы
cp .env.example .env
cp backend/.env.example backend/.env

# Установите pre-commit hook
./install-pre-commit-hook.bat  # На Windows
chmod +x install-pre-commit-hook.sh && ./install-pre-commit-hook.sh  # На Linux/Mac
```

### 2. Получите API ключи

**OpenAI (для AutoConfig):**
1. Перейдите на https://platform.openai.com/api-keys
2. Создайте новый API key
3. Скопируйте в `backend/.env` под переменную `OPENAI_API_KEY`

**Google API (для Discovery):**
1. Создайте Cloud Project в Google Cloud Console
2. Включите Custom Search API
3. Создайте API key
4. Скопируйте в `.env` под переменные `GOOGLE_API_KEY` и `GOOGLE_CX`

### 3. Запуск через Docker Compose

```bash
# Запустите все сервисы
docker-compose up -d

# Проверьте статус
docker-compose ps

# Посмотрите логи
docker-compose logs -f backend
```

### 4. Запуск без Docker (локальная разработка)

**Требования:**
- Go 1.24+
- Node.js 20+
- PostgreSQL 15+
- Redis 7+
- Meilisearch v1.3+

**Backend:**
```bash
cd backend
go mod download
go run cmd/api/main.go
```

**Frontend:**
```bash
cd frontend
npm ci
npm run dev
```

---

## 🔐 Правила безопасности

### ✅ DO's

- ✅ Копируйте `.env.example` → `.env` (локально)
- ✅ Добавляйте реальные ключи **только в локальный** `.env` файл
- ✅ Используйте environment переменные для production (GitHub Secrets, AWS Secrets, и т.д.)
- ✅ Запускайте `git commit` - pre-commit hook проверит вас
- ✅ Регулярно ротируйте API ключи (раз в 3 месяца)

### ❌ DON'Ts

- ❌ Не коммитьте файлы содержащие API ключи
- ❌ Не добавляйте реальные значения в `.env.example`
- ❌ Не используйте простые пароли (особенно для Meilisearch и PostgreSQL в prod)
- ❌ Не обходите pre-commit hook (`--no-verify` использовать только в крайних случаях)
- ❌ Не шарьте API ключи в Slack, Discord, email, и т.д.

---

## 🐛 Troubleshooting

### `docker-compose up` падает на "Dirty database version"

Решение:
```bash
docker-compose down -v  # Удалить все volumes
docker-compose up -d    # Запустить заново
```

### PostgreSQL не может подключиться

Проверьте что DB_PASSWORD установлен в `.env`:
```bash
cat .env | grep DB_PASSWORD
```

### OpenAI API возвращает ошибку аутентификации

1. Проверьте что ключ скопирован полностью (должен начинаться с `sk-proj-`)
2. Проверьте что ключ активен (не удалён/отключен)
3. Проверьте лимиты на использование в https://platform.openai.com/account/billing/overview

### Frontend не может подключиться к backend

1. Проверьте что backend запущен: `docker-compose ps`
2. Проверьте что `NEXT_PUBLIC_API_BASE` правильно установлен:
   - Local dev: `http://localhost:8080`
   - Docker: `http://backend:8080`
3. Посмотрите логи: `docker-compose logs backend`

---

## 📚 Полезные ссылки

- [Izborator Strategy](./STRATEGY.md)
- [Security Guidelines](./SECURITY_GUIDELINES.md)
- [Docker Compose Documentation](./docker-compose.README.md)
- [Backend README](./backend/README.md)
- [Frontend README](./frontend/README.md)

---

## 🤝 Что делать если найдёте проблему

1. **Проверьте GitHub Issues** - может кто-то уже решил проблему
2. **Посмотрите DEVELOPMENT_LOG.md** - может быть решение там
3. **Создайте Issue** на GitHub с:
   - Описанием проблемы
   - Шагами для воспроизведения
   - Логами ошибок
   - Your environment (OS, Go версия, Node версия)

---

**Last Updated:** 2025-12-21
