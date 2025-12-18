# Текущий шаг (2025-01-XX)

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

### ⏳ Шаг 2 — Протестировать /browse с категориями
- [x] Исправлен баг: category_id читается в searchViaPostgres
- [x] Исправлен баг: category_id добавляется в Meilisearch indexer
- [ ] GET /api/v1/products/browse?category=mobilni-telefoni работает (требует перезапуска API)
- [ ] GET /api/v1/products/browse?category=laptopovi работает
- [ ] GET /api/v1/products/browse (без фильтра) работает
- [ ] Проверка структуры BrowseResult
- [ ] Проверка fallback при несуществующем slug

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
