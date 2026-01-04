# Catalog API Contract (v1)

Цель: фронтенд (Catalog page) должен получать данные единообразно и предсказуемо.
Backend base URL (dev): http://backend:8080

---

## 1) Common

### 1.1 Headers
- Accept: application/json
- Content-Type: application/json (только для POST)
- X-Request-Id: optional (для трассировки)
- Accept-Language: optional (например: en, ru, sr)

### 1.2 Error model (единый для всех эндпоинтов)
Ответы 4xx/5xx должны быть в формате:

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "Human readable message",
    "details": {
      "field": "q",
      "reason": "too_short"
    }
  }
}
```

Коды:

INVALID_ARGUMENT (400)

UNAUTHORIZED (401)

FORBIDDEN (403)

NOT_FOUND (404)

CONFLICT (409)

RATE_LIMITED (429)

INTERNAL (500)

2) Catalog Search (основной)
GET /api/v1/catalog
Query params

q: string (optional) — текстовый поиск

minLen: 2 (если задан)

type: "all" | "goods" | "services" (optional, default: "all")

category_id: string (optional) — выбранная категория (id)

brand: string (optional) — бренд (строка)

city: string (optional) — город (строка)

price_from: number (optional, default: 0)

price_to: number (optional, default: 1000000)

sort: string (optional, default: "price_asc")

allowed:

relevance

price_asc

price_desc

newest

page: number (optional, default: 1)

page_size: number (optional, default: 20, max: 100)

200 Response
{
  "query": {
    "q": "smartphon",
    "type": "all",
    "category_id": null,
    "brand": null,
    "city": null,
    "price_from": 0,
    "price_to": 1000000,
    "sort": "price_asc",
    "page": 1,
    "page_size": 20
  },
  "paging": {
    "page": 1,
    "page_size": 20,
    "total": 0,
    "total_pages": 0
  },
  "items": []
}

CatalogItem (item schema)
{
  "id": "string",
  "type": "goods|services",
  "title": "string",
  "category": {
    "id": "string",
    "title": "string",
    "path": ["string"]
  },
  "price": {
    "amount": 12345,
    "currency": "RSD"
  },
  "location": {
    "city": "Belgrade",
    "country": "RS"
  },
  "brand": "Samsung",
  "image": {
    "url": "string",
    "width": 1200,
    "height": 800
  },
  "seller": {
    "id": "string",
    "name": "string"
  },
  "url": "string",
  "created_at": "2026-01-04T10:00:00Z"
}

Empty result is NOT an error

Если ничего не найдено — 200 + items: [] + total: 0.

400 (invalid query)

q задан, но короче 2 символов

price_from > price_to

sort вне allowed

page/page_size вне диапазона

3) Filters data (для дропдаунов)
GET /api/v1/catalog/filters

Возвращает наборы значений для фильтров.
Поддерживает зависимость от type/category/q (опционально), чтобы фильтры были релевантны.

Query params

q: string (optional)

type: "all" | "goods" | "services" (optional)

category_id: string (optional)

200 Response
{
  "categories": [
    { "id": "string", "title": "string", "path": ["string"] }
  ],
  "brands": ["Samsung", "Lenovo", "Nike"],
  "cities": ["Belgrade", "Novi Sad"],
  "price": { "min": 0, "max": 1000000 },
  "sort": [
    { "value": "price_asc", "label": "Price: Low to High" },
    { "value": "price_desc", "label": "Price: High to Low" },
    { "value": "newest", "label": "Newest" },
    { "value": "relevance", "label": "Relevance" }
  ]
}

4) Optional: Categories tree (если нужно отдельно)
GET /api/v1/categories

Возвращает дерево категорий (для Home/Cloud и для фильтра Category).

200 Response
{
  "items": [
    {
      "id": "string",
      "title": "string",
      "slug": "string",
      "children": [
        { "id": "string", "title": "string", "slug": "string", "children": [] }
      ]
    }
  ]
}

5) Frontend expectations (важно)

Фронт НЕ должен получать 400 при валидном q и дефолтных фильтрах.

/catalog/filters должен отдавать 200 даже если items=0 (и фильтры пустые допустимы).

Ошибка 400 должна быть структурированной (Error model).

Все currency/price поля должны быть числами, не строками.

Везде стабильные ключи и типы, без "Loading..."-заглушек от API.
