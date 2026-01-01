# Domain Pack Draft (goods, services)

**Статус:** черновик  
**Дата:** 2026-01-01  
**Цель:** зафиксировать минимальные Domain Pack для двух вертикалей без привязки к коду.

---

## Общие определения

**semantic_type** — смысловая роль атрибута (price, duration и т.д.).  
**facet_type** — тип UI-фасета (range, enum, boolean, text).  
**UI facets default** — набор фасетов, который UI показывает по умолчанию, если доступен.

Пример базовых semantic_type:  
`title`, `price`, `currency`, `availability`, `duration`, `rating`, `location`, `category`, `brand`, `image`, `description`, `specs`.

---

## Domain Pack: goods

**required semantic types**
- `title`
- `price`
- `currency`

**optional semantic types**
- `brand`
- `specs`
- `availability`
- `rating`
- `image`
- `description`
- `category`
- `location`

**UI facets default**
| semantic_type | facet_type | Примечание |
|---|---|---|
| price | range | основной фильтр |
| category | enum | из categories |
| location | enum | город/регион |
| brand | enum | если есть |
| rating | range | если есть |
| availability | boolean | если есть |

---

## Domain Pack: services

**required semantic types**
- `title`
- **one_of:** `duration` **или** `availability`

**optional semantic types**
- `price`
- `currency`
- `rating`
- `location`
- `service_area`
- `provider_name`
- `image`
- `description`
- `category`

**UI facets default**
| semantic_type | facet_type | Примечание |
|---|---|---|
| category | enum | из categories |
| location / service_area | enum | город/район/зона |
| duration | range | если есть |
| price | range | если есть |
| availability | boolean | если есть |
| rating | range | если есть |

---

## Примечания
- `one_of` означает: достаточно любого из перечисленных semantic_type для прохождения валидации.  
- Эти пакеты описывают **поведение**, а не конкретные источники данных.  
- Следующий шаг: согласовать список semantic_type и закрепить их в `attributes` как метаданные.

---

## Validation rules v1

goods.valid_if = title AND price (currency optional/defaultable)

services.valid_if = title AND location AND (duration OR availability OR price)

---

## Semantic Types Registry (v0)

| semantic_type | kind | notes |
|---|---|---|
| title | core | required for all |
| location | core | city / area |
| category | core | taxonomy |
| price | attribute | numeric |
| currency | derivative | defaultable |
| duration | attribute | minutes/hours |
| availability | attribute | boolean/schedule |
| rating | attribute | numeric |
| brand | attribute | vendor/brand |
| image | attribute | url |
| description | attribute | text |
| specs | attribute | key/value |

---

## Semantic → DB mapping (current)

| semantic_type | source table | field / path |
|---|---|---|
| title | products / raw_products | products.name (canonical), raw_products.name (source) |
| location | cities / product_prices / shops | cities.id + product_prices.city_id, shops.default_city_id |
| category | categories / products / raw_products | categories.id/slug, products.category_id (+ products.category legacy), raw_products.category |
| price | product_prices / raw_products | product_prices.price, raw_products.price |
| currency | product_prices / raw_products | product_prices.currency, raw_products.currency |
| duration | — | not stored yet |
| availability | product_prices / raw_products | product_prices.in_stock, raw_products.in_stock |
| rating | — | not stored yet |
| brand | products / raw_products | products.brand, raw_products.brand |
| image | products / raw_products | products.image_url, raw_products.image_urls (JSONB) |
| description | products / raw_products | products.description, raw_products.description |
| specs | products / raw_products | products.specs (JSONB), raw_products.specs (JSONB) |

category.canonical = products.category_id; fallback: raw_products.category -> categories.slug resolver; legacy ignore

location.canonical = product_prices.city_id; fallback: shops.default_city_id; else null

---

## Semantic Validation Result (v1)

```json
{
  "domain": "services",
  "valid": false,
  "missing_semantic": [
    "location",
    "duration|availability|price"
  ],
  "present_semantic": [
    "title"
  ],
  "notes": "price missing but optional; no location detected"
}
```

---

## Quality Score v1

quality_score = 0.5*valid_rate + 0.3*semantic_coverage + 0.2*normalization_success  
valid_rate = доля valid объектов; semantic_coverage = 1 - (avg_missing_semantic / required_semantic_count); normalization_success = доля нормализованных значений

---

## CI Quality Gates v1

goods:
  valid_rate >= 0.95
  quality_score >= 0.85

services:
  valid_rate >= 0.80
  quality_score >= 0.70

---

## Pipeline Integration Points (v1)

- normalize(): emits present_semantic[]
- validate(): produces Semantic Validation Result
- metrics(): aggregates quality_score per domain
- CI: fails if Quality Gates not met
- UI: builds facets from semantic_type + facet_type
