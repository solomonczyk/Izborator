# ROADMAP — izborator (frontend+backend)

> Цель: понятная карта дальнейших шагов от текущего состояния до “готово к продакшену + масштабирование по tenant/доменам”.
> Формат: чек-лист по этапам. Можно вести прогресс галочками.

## Где мы сейчас (срез на текущий момент)

✅ Закрыто:
- Next обновлён, `npm audit` чистый
- E2E Playwright проходит, webServer корректно поднимает Next
- Facet schema → UI (gating) → CI contract check
- SSR оптимизация: параллелизация fetch, гейтинг категорий/городов, кеширование where applicable
- Production hardening: Playwright в CI, env separation
- Observability v1: структурированный `catalog_ssr` лог (ms + facets_count) + warn-thresholds
- Multi-tenant plumbing: `tenant_id` обязателен для facets + включён в SSR лог + CI enforcement
- Tenant isolation v1: `tenant_id` обязателен для browse/categories/cities + тесты обновлены
- Soft-limits v1: глобальные и per-tenant overrides (`TENANT_LIMITS_JSON`) + warn на parse fail

---

## Этап 1 — Стабилизация и эксплуатационная готовность (Ops-ready)

- [ ] **Tenant health snapshot**
  - Цель: быстрый “снимок здоровья” tenant’а (counts/limits/over-limit) для поддержки.
  - Опции:
    - Вариант A: internal endpoint `/api/internal/tenant/health`
    - Вариант B: отдельный structured log по запросу/по расписанию

- [ ] **Hard limits v1 (защита от перегруза)**
  - Цель: не просто warn, а управляемое ограничение.
  - Правила (пример):
    - если `facets_count` > limit → урезать до top-N (или 422 с сообщением)
    - если `brands_count` > limit → ограничить выдачу брендов (пагинация/поиск)
  - Важно: добавить контракт в тесты.

- [ ] **Rate limiting / abuse protection**
  - Цель: tenant не может DDoS’нуть browse/facets.
  - Подход: token-bucket per tenant + общий лимит.
  - Логи: `rate_limited` с tenant_id.

- [ ] **Error taxonomy**
  - Единый формат ошибок (code/message/details) для всех API.
  - CI: проверка, что ошибки соответствуют схеме.

---

## Этап 2 — Performance / DX v2 (ощутимая скорость)

- [ ] **Performance budgets**
  - Определить бюджеты: SSR ms (p95), facets latency, browse latency.
  - CI gate (мягкий): warn при превышении бюджета.

- [ ] **Cache strategy per tenant**
  - Кеширование facets/schema и справочников (categories/cities) по tenant_id.
  - Ключи кеша: `{tenant_id}:{domain}:{type}:{locale}`.
  - Revalidate: разный TTL для “справочники” vs “динамика”.

- [ ] **Prefetch v2**
  - Сейчас: warm-up fetch.
  - Дальше: prefetch на hover/intent + debounce, чтобы не спамить API.

- [ ] **Parallel fetch audit**
  - Проверить, что лишних await-цепочек нет.
  - Вынести независимые запросы в `Promise.allSettled`/`all` там, где безопасно.

---

## Этап 3 — Domain scaling v2 (домены как продуктовые пакеты)

- [ ] **Domain pack schema versioning**
  - Версионирование domain packs (v1/v2) + миграции.
  - CI: валидатор JSON schema.

- [ ] **Add 2–3 “стресс-домена”**
  - Один без brand/location.
  - Один с большим набором facets.
  - Один с кастомными правилами сортировки/ранжирования.

- [ ] **UI regression matrix**
  - Автопрогон “каждый домен × goods/services” на контракт/рендер.

---

## Этап 4 — Production hardening v2 (безопасность, релизы, деградации)

- [ ] **Env policy**
  - Чётко разделить env: local/dev/stage/prod.
  - Док: какие env обязательны, какие дефолтятся, какие запрещены на prod.

- [ ] **Security headers / middleware audit**
  - CSP (минимальная, не ломая Next dev)
  - HSTS (prod)
  - Referrer-Policy, X-Content-Type-Options, etc.

- [ ] **CI/CD release flow**
  - Versioning (semver)
  - Changelog generation
  - “Release candidate” pipeline + smoke tests

- [ ] **Degradation modes**
  - Если facets недоступны → каталог должен работать (без фильтров) + warn.
  - Если cities/categories недоступны → скрыть фильтры + warn.

---

## Этап 5 — Analytics и бизнес-метрики (чтобы управлять продуктом)

- [ ] **Facet usage analytics**
  - Какие фасеты реально используют по tenant/domain.
  - События: facet_opened, facet_applied, facet_cleared (privacy-safe).

- [ ] **Search/Filter funnels**
  - conversion: просмотр → фильтр → карточка → действие.

- [ ] **Tenant dashboards**
  - Срез: latency, errors, usage, over-limit counts.

---

## Этап 6 — Multi-tenant productization (под клиентов/партнёров)

- [ ] **Tenant provisioning**
  - Создание/удаление tenant, секреты, дефолтные пакеты.

- [ ] **Billing hooks (заготовка)**
  - Лимиты/тарифы → hard limits.
  - Usage metrics → счётчик.

- [ ] **Tenant isolation audit**
  - Проверить все endpoints: везде ли tenant_id обязателен.
  - Unit/integration tests: “no tenant → 400” и “wrong tenant → no data leak”.

---

## Definition of Done (что считается “концом проекта”)

Проект можно считать “готовым”, когда:
- [ ] Все внешние API требуют `tenant_id` там, где есть данные tenant’а
- [ ] CI: unit + integration + Playwright стабильно зелёные
- [ ] Есть бюджеты производительности + предупреждения в логах
- [ ] Есть degradation-mode при падении зависимостей
- [ ] Есть базовая безопасность (headers + rate limiting)
- [ ] Есть документация env + запуск локально/в docker + релизный процесс

---

## Быстрый порядок выполнения (рекомендованный)

1) Tenant health snapshot  
2) Hard limits v1  
3) Rate limiting  
4) Cache strategy per tenant  
5) Performance budgets  
6) Domain pack versioning + stress-domains  
7) Degradation modes  
8) Security headers  
9) CI/CD release flow  
10) Analytics (facet usage)  
11) Tenant provisioning + billing hooks

