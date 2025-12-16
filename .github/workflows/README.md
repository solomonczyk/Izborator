# GitHub Actions - Автоматический деплой

## Настройка

### 1. Добавь Secrets в GitHub

Перейди в Settings → Secrets and variables → Actions и добавь:

- `SSH_PRIVATE_KEY` - приватный SSH ключ для доступа к серверу
- `SERVER_HOST` - IP адрес или домен сервера (например: `152.53.227.37`)

### 2. Генерация SSH ключа (если еще нет)

На локальной машине:
```bash
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github_actions
```

Скопируй публичный ключ на сервер:
```bash
ssh-copy-id -i ~/.ssh/github_actions.pub root@152.53.227.37
```

Приватный ключ (`~/.ssh/github_actions`) добавь в GitHub Secrets как `SSH_PRIVATE_KEY`.

### 3. Запуск деплоя

**Автоматически:**
- При каждом `git push` в ветку `main` деплой запустится автоматически

**Вручную:**
- Перейди в Actions → Deploy to Production → Run workflow

## Что делает workflow

1. ✅ Клонирует код из репозитория
2. ✅ Подключается к серверу по SSH
3. ✅ Обновляет код (`git pull`)
4. ✅ Генерирует SSL сертификаты (если их нет)
5. ✅ Пересобирает Docker контейнеры
6. ✅ Перезапускает сервисы
7. ✅ Проверяет health check

## Логи

Все логи деплоя можно посмотреть в разделе Actions на GitHub.

