# 🔍 Диагностика проблемы Classifier

## Проблема
Classifier обрабатывает 85 записей, но статусы не обновляются в базе данных. Все записи остаются со статусом "new".

## Возможные причины

### 1. Backend контейнер не пересобран
**Решение:** Пересобрать backend с флагом `--no-cache`:
```bash
cd ~/Izborator
docker-compose build --no-cache backend
docker-compose up -d
```

### 2. Domain не совпадает
**Проверка:** Сравнить domain в базе и в коде:
```bash
# Проверить domain в базе
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT domain, status FROM potential_shops LIMIT 5;
"

# Проверить логи Classifier на конкретные domain
docker-compose logs backend | grep -i "domain=" | head -10
```

### 3. Ошибки в SQL запросе
**Проверка:** Посмотреть детальные логи ошибок:
```bash
# Логи Classifier
docker-compose run --rm backend ./classifier -classify-all 2>&1 | grep -i "error\|failed\|update"

# Или проверить последние логи
docker-compose logs backend | tail -100 | grep -i "error\|failed"
```

### 4. Проблема с транзакциями
**Проверка:** Убедиться, что транзакции коммитятся:
```sql
-- Проверить, что записи действительно обновляются
SELECT domain, status, updated_at 
FROM potential_shops 
ORDER BY updated_at DESC 
LIMIT 10;
```

## Пошаговая диагностика

### Шаг 1: Пересобрать backend
```bash
cd ~/Izborator
git pull
docker-compose build --no-cache backend
docker-compose up -d
```

### Шаг 2: Проверить код в контейнере
```bash
# Проверить, что код обновлен (опционально - можно посмотреть через docker exec)
docker-compose exec backend cat /app/classifier 2>/dev/null || echo "Бинарник скомпилирован"
```

### Шаг 3: Запустить Classifier с детальными логами
```bash
docker-compose run --rm backend ./classifier -classify-all 2>&1 | tee /tmp/classifier-debug.log
```

### Шаг 4: Проверить ошибки в логах
```bash
# Найти все ошибки обновления
grep -i "failed to update\|error updating\|no rows updated" /tmp/classifier-debug.log

# Проверить конкретные domain, которые не обновились
grep -i "error" /tmp/classifier-debug.log | grep -i "domain="
```

### Шаг 5: Проверить статусы в базе
```bash
docker-compose exec -T postgres psql -U postgres -d izborator -c "
  SELECT status, COUNT(*) as count
  FROM potential_shops
  GROUP BY status
  ORDER BY status;
"
```

### Шаг 6: Тестовое обновление вручную
```bash
# Попробовать обновить одну запись вручную через SQL
docker-compose exec -T postgres psql -U postgres -d izborator << 'SQL'
  -- Получить первый domain
  SELECT domain FROM potential_shops WHERE status = 'new' LIMIT 1;
  
  -- Обновить его вручную
  UPDATE potential_shops 
  SET status = 'test', updated_at = NOW() 
  WHERE domain = (SELECT domain FROM potential_shops WHERE status = 'new' LIMIT 1);
  
  -- Проверить результат
  SELECT domain, status, updated_at 
  FROM potential_shops 
  WHERE status = 'test';
SQL
```

## Использование скрипта для автоматической диагностики

```bash
chmod +x rebuild-and-test-classifier.sh
./rebuild-and-test-classifier.sh
```

Этот скрипт:
1. Обновит код
2. Пересоберет backend с `--no-cache`
3. Запустит Classifier с детальными логами
4. Покажет статистику до и после
5. Покажет примеры записей

## Если проблема сохраняется

1. **Проверить логи backend контейнера:**
   ```bash
   docker-compose logs backend | tail -200
   ```

2. **Проверить подключение к БД:**
   ```bash
   docker-compose exec backend ./migrate status
   ```

3. **Проверить права доступа к таблице:**
   ```bash
   docker-compose exec -T postgres psql -U postgres -d izborator -c "
     SELECT grantee, privilege_type 
     FROM information_schema.role_table_grants 
     WHERE table_name = 'potential_shops';
   "
   ```

4. **Проверить структуру таблицы:**
   ```bash
   docker-compose exec -T postgres psql -U postgres -d izborator -c "\d potential_shops"
   ```

