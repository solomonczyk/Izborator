API contracts

Все tenant-aware эндпоинты требуют tenant_id (явно задокументировано: где передаётся).

Ошибки стандартизированы: code, message, details?.

Для /api/v1/products/facets:

tenant_id обязателен

type валидируется по доступным доменам/конфигу

Ответ содержит tenant_id (для трассировки)

Validation & security

tenant_id валидируется (формат + существование/whitelist).

Нет “тихих дефолтов”, которые могут смешивать данные разных tenant.

Логи не содержат PII и не содержат “сырых конфигов”.

Caching

Кэш (если есть) key’ed минимум по: tenant_id + locale + domain/type.

Исключены коллизии ключей между доменами и tenant.

TTL и стратегия инвалидации описаны.

Observability

Структурированные логи для:

catalog_ssr (уже есть)

home_model (если /home присутствует)

facets_fetch (опционально, если нужен трейс)

Warn-пороги управляются через env.

CI / Contract checks

Контракт UI↔schema проверяется в CI по всем доменам.

Отдельная проверка: tenant обязателен (тест/линт/контракт).

Integration tests

Минимум 2 интеграционных теста:

facets без tenant_id → 400 VALIDATION_FAILED

facets с tenant_id и валидным type → 200 + корректная схема

Docs

README/API docs описывают tenant модель (как передавать, примеры).

Примеры curl-запросов для ключевых эндпоинтов.

Definition of Done: все пункты выше выполнены + CI зелёный.
