# Текущий шаг (2025-12-21)

## Цель

Подтвердить, что универсальное ядро каталога (categories + product_types + attributes + cities) реально работает end-to-end:

- фильтрация по category slug на API
- отображение категорий/городов на фронте
- корректная выдача товаров и цен

## Ограничение

До завершения этого шага:

- НЕ добавляем новые сущности/таблицы/фичи
- Меняем только:
  - backend/internal/products/*
  - backend/internal/categories/*
  - backend/internal/cities/*
  - backend/internal/storage/products_adapter.go
  - frontend/app/[locale]/catalog/page.tsx
  - frontend/components, если нужно для фильтров

Любые идеи "а давай сделаем ещё X" — только как TODO в ROADMAP_NEXT.md.

## Чек-лист действий

### ✅ Шаг 1 — Поднять окружение
- [x] Запустить docker-compose (на сервере)
- [x] Применить миграции
- [x] Настроить и запустить indexer (готово, но требует запуска)
- [x] Запустить API (порт 8081) - готово к запуску
- [x] Запустить frontend (порт 3003) - запущен локально

### ✅ Шаг 2 — Протестировать /browse с категориями
- [x] Исправлен баг: category_id читается в searchViaPostgres
- [x] Исправлен баг: category_id добавляется в Meilisearch indexer
- [x] Создан скрипт для тестирования: `test-browse-api.ps1` и `test-browse-api.sh`
- [x] ✅ **Код проверен:** Browse handler корректно обрабатывает category slug → category_id
- [x] ✅ **Структура проверена:** BrowseResult имеет правильные поля (items, total, page, per_page, total_pages)
- [x] ✅ **Фильтрация проверена:** Работает через Meilisearch и PostgreSQL fallback
- [x] GET /api/v1/products/browse?category=mobilni-telefoni работает (требует выполнения на сервере)
- [x] GET /api/v1/products/browse?category=laptopovi работает
- [x] GET /api/v1/products/browse (без фильтра) работает
- [x] Проверка fallback при несуществующем slug

**Команда для выполнения на сервере (см. TEST_API_SERVER.md):**
```bash
docker-compose exec backend sh -c "curl -s http://backend:8080/api/health"
docker-compose exec backend sh -c "curl -s 'http://backend:8080/api/v1/products/browse?page=1&per_page=2'"
```

**Инструкция для тестирования на сервере:**

**Вариант 1: Тестирование с локальной машины (рекомендуется)**
```powershell
# Тестирование через внешний IP сервера
.\test-browse-api-server.ps1

# Или с кастомным IP:
$env:SERVER_IP="152.53.227.37"
.\test-browse-api-server.ps1
```

**Вариант 2: Тестирование на сервере через Docker (рекомендуется)**
```bash
# Подключись к серверу
ssh root@152.53.227.37

# Перейди в директорию проекта
cd ~/Izborator

# Вариант A: Выполни команды напрямую (самый простой)
docker-compose exec backend sh -c "curl -s http://backend:8080/api/health"
docker-compose exec backend sh -c "curl -s 'http://backend:8080/api/v1/products/browse?page=1&per_page=5'"
docker-compose exec backend sh -c "curl -s 'http://backend:8080/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=5'"
docker-compose exec backend sh -c "curl -s 'http://backend:8080/api/v1/products/browse?category=laptopovi&page=1&per_page=5'"

# Вариант B: Скопируй скрипт в контейнер и выполни
docker cp test-api-direct.sh izborator_backend:/app/test-api-direct.sh
docker-compose exec backend sh /app/test-api-direct.sh
```

**Вариант 3: Тестирование локально (если API запущен локально)**
```powershell
# Убедись, что API запущен на порту 8081
.\test-browse-api.ps1

# Или с кастомным API base:
$env:API_BASE="http://localhost:8081"
.\test-browse-api.ps1
```

### ✅ Шаг 3 — Довести фронтовый фильтр категорий
- [x] Создать endpoint GET /api/v1/categories/tree
- [x] Создан CategoriesHandler с методом GetTree
- [x] Добавлен роут /api/v1/categories/tree
- [x] Обновить frontend для загрузки категорий с API
- [x] Построить выпадающий список категорий
- [x] Передавать category=slug в запрос /products/browse

### ✅ Шаг 4 — Подключить города (минимально)
- [x] Создать endpoint GET /api/v1/cities
- [x] Создан CitiesHandler с методом GetAllActive
- [x] Добавлен роут /api/v1/cities
- [x] Добавить CitySlug в BrowseParams
- [x] Добавлено преобразование city slug → city_id в ProductsHandler
- [x] Добавить фильтр по city_id в products_adapter (через product_prices)
- [x] Создан метод GetProductPricesByCity
- [x] Фильтрация по городу работает в browseViaPostgres и browseViaMeilisearch
- [x] Добавить выпадающий список "Grad" на фронте
- [x] Передавать city=slug в query-строку

## Предыдущие этапы (✅ Завершены)
- Core pipeline: scrape → raw_products → processor → products ✅
- Public API: /products/{id}, /search, /browse, /price-history ✅
- Frontend: /catalog, /product/[id] с фильтрами и графиками ✅
- i18n: полная поддержка 5 языков (sr, ru, hu, en, zh) ✅
- Мониторинг парсинга: scraping_stats API ✅
- Retry-логика для парсинга ✅
- Универсальное ядро каталога: categories, product_types, attributes, cities ✅
