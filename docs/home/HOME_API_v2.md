# Home API v2 — /api/v1/home

Цель: описать API, которое отдаёт HomeModel v2
для SSR и client-side рендера главной страницы.

---

## Endpoint

GET /api/v1/home

---

## Query params

- tenant_id (string, required)
- locale (string, required)

---

## Response 200

```json
{
  "version": "2",
  "tenant_id": "default",
  "locale": "sr",

  "hero": {
    "title": "Pronađite proizvode i usluge",
    "subtitle": "Uporedite ponude za minut",
    "searchPlaceholder": "Šta tražite?",
    "showTypeToggle": true,
    "showCitySelect": false,
    "defaultType": "all"
  },

  "featuredCategories": [
    {
      "category_id": "elektronika",
      "title": "Elektronika",
      "href": "/catalog?category=elektronika",
      "priority": "primary",
      "order": 1,
      "icon_key": "electronics"
    }
  ]
}

Errors

400 — missing tenant_id / locale

404 — tenant not found

500 — internal error

SSR & Caching
Cache policy

Cache-Control:

public

max-age=60

s-maxage=300

stale-while-revalidate=600

Cache keys

tenant_id

locale

Backend responsibilities

Читает canonical_tree_v1.json

Применяет FEATURED_RULES

Формирует HomeModel v2

Логирует:

tenant_id

locale

featuredCategories.length

response time

Frontend responsibilities

Делает fetch

Рендерит Home UI

Не вычисляет категории

Не меняет порядок

Fallback behavior

Если API недоступен:

SSR рендерит Home без featuredCategories

поиск остаётся доступным

Definition of Done

API считается готовым, если:

Home UI может отрендериться полностью

Нет логики выбора категорий на фронте

Кеш работает tenant-aware
