# Инструкция по созданию тестовой базы данных

## Проблема
PowerShell не может подключиться к Docker через pipe. Выполните команды вручную.

## Способ 1: Через терминал Docker Desktop или Git Bash

Откройте терминал (Git Bash, WSL, или терминал Docker Desktop) и выполните:

```bash
cd F:/Dev/Projects/Izborator

# 1. Создание базы данных
docker exec -i izborator_postgres psql -U postgres < backend/scripts/create_test_db.sql

# 2. Применение схемы
docker exec -i izborator_postgres psql -U postgres -d izborator < backend/scripts/create_test_db_in_izborator.sql

# 3. Добавление тестовых данных
docker exec -i izborator_postgres psql -U postgres -d izborator < backend/scripts/seed_test_data.sql

# 4. Проверка
docker exec -i izborator_postgres psql -U postgres -d izborator -c "SELECT COUNT(*) as shops FROM shops; SELECT COUNT(*) as products FROM products; SELECT COUNT(*) as prices FROM product_prices;"
```

## Способ 2: Через docker-compose (если доступен)

```bash
cd F:/Dev/Projects/Izborator

# 1. Создание базы данных
docker-compose exec -T postgres psql -U postgres < backend/scripts/create_test_db.sql

# 2. Применение схемы
docker-compose exec -T postgres psql -U postgres -d izborator < backend/scripts/create_test_db_in_izborator.sql

# 3. Добавление тестовых данных
docker-compose exec -T postgres psql -U postgres -d izborator < backend/scripts/seed_test_data.sql
```

## Способ 3: Через psql напрямую (если PostgreSQL доступен локально)

Если у вас установлен PostgreSQL локально и он доступен на порту 5432:

```bash
# 1. Создание базы данных
psql -U postgres -f backend/scripts/create_test_db.sql

# 2. Применение схемы
psql -U postgres -d izborator -f backend/scripts/create_test_db_in_izborator.sql

# 3. Добавление тестовых данных
psql -U postgres -d izborator -f backend/scripts/seed_test_data.sql
```

## Что будет создано

- База данных `izborator`
- Таблицы: `shops`, `raw_products`, `products`, `product_prices`
- Тестовый магазин Gigatron
- 3 тестовых товара с ценами:
  - Motorola G72 8/256GB Gray — 29,999 RSD
  - Samsung Galaxy A54 128GB Black — 34,999 RSD
  - iPhone 15 Pro 256GB Natural Titanium — 129,999 RSD

## После выполнения

После успешного выполнения скриптов:
1. Обновите `.env` файл: `DB_PORT=5433` (если используете Docker)
2. Запустите API: `go run cmd/api/main.go`
3. Проверьте endpoint: `http://localhost:8080/api/v1/products/browse`

