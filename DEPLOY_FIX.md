# 🔧 Инструкция по деплою исправления UpdatePotentialShop

## Проблема
Classifier обрабатывал кандидатов, но статусы не обновлялись в базе данных из-за ошибки в `UpdatePotentialShop`.

## Исправления
1. Исправлен метод `UpdatePotentialShop` в `classifier_adapter.go`
2. Улучшено логирование ошибок
3. Исправлена миграция (тип shop_id)

## Деплой

### Вариант 1: Через Git (рекомендуется)

```bash
# На локальной машине
git add backend/cmd/classifier/main.go backend/internal/storage/classifier_adapter.go backend/migrations/0006_discovery_tables.up.sql
git commit -m "Fix UpdatePotentialShop: handle NULL metadata and improve error logging"
git push
```

GitHub Actions автоматически задеплоит изменения на сервер.

### Вариант 2: Вручную на сервере

```bash
# На сервере
ssh root@152.53.227.37
cd ~/Izborator
git pull
docker-compose build backend
docker-compose run --rm backend ./classifier -classify-all
```

## Проверка результатов

После деплоя и запуска Classifier:

```bash
bash check-pipeline-results.sh
```

Должно показать:
- ✅ Статусы обновлены (classified, pending_review, rejected)
- ✅ Нет записей со статусом "new" (или меньше 85)

## Если проблема сохраняется

Проверь логи:
```bash
docker-compose logs backend | grep -i "Failed to update shop" | tail -10
```

