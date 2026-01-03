Home / Discover API Specification (v1.0)

1) Общие принципы API

Все запросы tenant-aware

Locale влияет на тексты (если отдаём локализованный HomeModel)

API отдаёт готовые intent-ссылки (href)

UI не делает вычислений логики “какие карточки показывать”

2) Endpoints

2.1 GET /api/v1/home

Назначение: получить HomeModel для отрисовки главной.

Query params

tenant_id (string, required)

locale (string, optional) — если не берётся из маршрута/Accept-Language

пример: sr, ru, en

Response 200 (JSON)
{
  "version": "1",
  "tenant_id": "test-tenant",
  "locale": "sr",
  "hero": {
    "title": "Найдите товары и услуги",
    "subtitle": "Сравнивайте цены и предложения",
    "searchPlaceholder": "Что вы ищете?",
    "showTypeToggle": true,
    "showCitySelect": true,
    "defaultType": "all"
  },
  "categoryCards": [
    {
      "id": "electronics",
      "title": "Электроника",
      "hint": "Телефоны, ноутбуки, гаджеты",
      "icon_key": "electronics",
      "href": "/catalog?type=good&category=electronics",
      "priority": "primary",
      "weight": 10,
      "domain": "good",
      "analytics_id": "home_card_electronics"
    }
  ]
}

Errors

400 BAD_REQUEST — отсутствует tenant_id

404 NOT_FOUND — tenant не найден / витрина не настроена

500 INTERNAL_SERVER_ERROR — неожиданные ошибки

2.2 GET /api/v1/home/meta (опционально v1)

Назначение: лёгкий endpoint для prefetch/boot, если home тяжёлый.

Ответ:

версия

минимальные флаги

количество карточек

3) HTTP caching strategy

3.1 Серверный кеш (рекомендуемо)

/api/v1/home допускает кеширование, потому что:

меняется редко

зависит от tenant/locale

Заголовки (пример)

Cache-Control: public, max-age=60, s-maxage=300, stale-while-revalidate=600

Пояснение:

max-age=60 — браузер 1 мин

s-maxage=300 — CDN/edge 5 мин

stale-while-revalidate=600 — ещё 10 мин отдаём старое, пока обновляем

3.2 Важно про multi-tenant

Кеш должен быть key’ed по:

tenant_id

locale

4) Tenant rules

4.1 Как передаётся tenant

v1: query param tenant_id обязателен.

Возможные v2-модели:

subdomain tenancy: tenantA.example.com

path tenancy: /t/tenantA

Но в v1 фиксируем:

query-param (простой, прозрачный, тестируемый)

5) Locale rules

Вариант A (рекомендую v1)

Locale берётся из URL маршрута Next (/[locale]/) и прокидывается в API как параметр.

Вариант B

Locale вычисляется по Accept-Language, если параметр не передан.

6) Security / Privacy

tenant_id валидируется (whitelist/exists)

API не должен раскрывать внутренние конфиги целиком

Не отдаём сырые файлы “domain pack” наружу, только собранный HomeModel

7) Observability (логирование)

Каждый запрос /api/v1/home логируется структурировано:

event: home_model

tenant_id

locale

cards_count

ms

level: warn если ms > threshold

(порог берём из env, по аналогии с catalog_ssr)

8) Definition of Done (API Shape)

Готово, если:

- есть зафиксированный контракт /api/v1/home
- определены обязательные параметры tenant/locale
- определены коды ошибок
- определена стратегия кеширования
- определены правила наблюдаемости

Конец документа
