# Izborator Frontend

Next.js фронтенд для Izborator - агрегатора цен на товары.

## Установка

```bash
npm install
```

## Запуск

```bash
npm run dev
```

Откройте [http://localhost:3000](http://localhost:3000) в браузере.

## Страницы

- `/` - Главная страница
- `/catalog` - Каталог товаров с поиском
- `/catalog?q=motorola` - Поиск по запросу

## Переменные окружения

Создайте файл `.env.local`:

```
NEXT_PUBLIC_API_BASE=http://localhost:8080
```

## API

Фронтенд использует следующие endpoints:

- `GET /api/v1/products/browse` - Каталог товаров с фильтрами
- `GET /api/v1/products/{id}` - Детали товара
- `GET /api/v1/products/search?q=query` - Поиск товаров
