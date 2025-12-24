# 🔒 РУКОВОДСТВО ПО БЕЗОПАСНОСТИ

## API Ключи и Секреты

### Получение ключей

1. **OpenAI API Key**
   - Сайт: https://platform.openai.com/api-keys
   - Тип: Секретный
   - Использование: AutoConfig модуль
   - Переменная окружения: `OPENAI_API_KEY`

2. **Google API Key**
   - Сайт: https://cloud.google.com/docs/authentication/api-keys
   - Тип: Может быть публичным (ограничить по IP)
   - Использование: Discovery модуль (поиск магазинов)
   - Переменная окружения: `GOOGLE_API_KEY`

3. **Meilisearch Master Key**
   - По умолчанию: `masterKey123` (ТОЛЬКО для dev!)
   - Production: Сгенерировать в панели администратора
   - Переменная окружения: `MEILISEARCH_MASTER_KEY`

### Хранение и защита

❌ **НИКОГДА:**
- Не коммитьте реальные ключи в Git
- Не кладите .env файлы в репо
- Не логируйте чувствительные данные
- Не отправляйте ключи в сообщениях/чатах

✅ **ВСЕГДА:**
- Используйте `.env.example` для примеров
- Добавляйте `.env` и `.env.local` в `.gitignore`
- Используйте GitHub Secrets для CI/CD
- Ротируйте ключи каждые 3 месяца
- Используйте разные ключи для dev/staging/prod

### Проверка утечек в истории Git

```bash
# Сканирование истории на утечки
git log -p | grep -i "api_key\|secret\|password"

# Используйте git-secrets
git secrets --install
git secrets --register-aws
git secrets --scan

# Удаление чувствительных данных из истории
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch .env' \
  --prune-empty --tag-name-filter cat -- --all
```

## Переменные окружения

### Backend (.env)

```bash
# API Configuration
API_HOST=0.0.0.0
API_PORT=3002
WORKER_CONCURRENCY=10

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=izborator
DB_PASSWORD=secure-password-here
DB_NAME=izborator_prod

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# Meilisearch
MEILISEARCH_HOST=meilisearch:7700
MEILISEARCH_MASTER_KEY=masterKey123

# External APIs
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=AIza...

# Logging
LOG_LEVEL=info

# Secret Key для sessions
SECRET_KEY=generate-random-key-here
```

### Frontend (.env.local)

```bash
NEXT_PUBLIC_API_BASE=http://api:3002
NEXT_PUBLIC_ENV=production
```

## Требования к паролям БД

- Минимум 16 символов
- Включать: заглавные, строчные, цифры, спецсимволы
- Не использовать словари
- Генерировать: `openssl rand -base64 16`

## Процесс ротации ключей

### Каждые 3 месяца:
1. Сгенерировать новый ключ
2. Обновить в production
3. Обновить в GitHub Secrets
4. Удалить старый ключ
5. Документировать в changelog

## Мониторинг утечек

- GitHub: Secret scanning (Settings → Security → Secret scanning)
- AWS: GuardDuty
- Независимо: https://haveibeenpwned.com/

## Ответственность

- **DevOps:** Управление infrastructure secrets
- **Backend:** Управление API ключами
- **Frontend:** Отсутствие чувствительных данных в коде
- **Все:** Проверка перед commit
