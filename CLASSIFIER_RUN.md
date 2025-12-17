# 🔍 Запуск Классификатора

## Локальный запуск (для тестирования)

```bash
cd backend
go run ./cmd/classifier/main.go -domain "gigatron.rs"
```

## Запуск в Docker (на сервере)

### 1. Классификация всех найденных доменов

```bash
docker-compose run --rm backend ./classifier -classify-all
```

### 2. Классификация с лимитом (первые N доменов)

```bash
docker-compose run --rm backend ./classifier -classify-all -limit 10
```

### 3. Тест одного домена

```bash
docker-compose run --rm backend ./classifier -domain "gigatron.rs"
```

### 4. Тест на предопределенном списке

```bash
docker-compose run --rm backend ./classifier -test-list
```

## Переменные окружения

Убедись, что в `.env` файле (в корне проекта) заданы:
```env
GOOGLE_API_KEY=твой_ключ
GOOGLE_CX=твой_cx_id
```

Эти переменные автоматически передаются в контейнер через `docker-compose.yml`.

## Результаты

После классификации:
- Домены со статусом `classified` - это магазины (score >= 0.8)
- Домены со статусом `pending_review` - требуют ручной проверки (score >= 0.5)
- Домены со статусом `rejected` - не магазины (score < 0.5)

Проверить результаты можно через SQL:
```sql
SELECT domain, status, confidence_score, classified_at 
FROM potential_shops 
WHERE status IN ('classified', 'pending_review')
ORDER BY confidence_score DESC;
```

