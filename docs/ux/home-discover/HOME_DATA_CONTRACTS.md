Home / Discover Data Contracts (v1.0)

1) Источники данных (Source of Truth)

1.1 Что является источником истины

Список “витринных категорий” для главной — конфиг/Domain Pack (а не БД)

Локализация текстов — i18n словари

Ссылки/intent — формируются из конфигурации (без вычислений на UI)

1.2 Multi-tenant

Для каждого tenant_id может быть:

- свой набор карточек
- свой порядок
- свои приоритеты
- свои “featured” категории

2) Данные для страницы Home (минимально необходимые)

2.1 Контракт HomeModel (payload для рендера)

type HomeModel = {
  tenant_id: string
  locale: string

  hero: {
    title: string
    subtitle?: string
    searchPlaceholder: string
    showTypeToggle: boolean
    showCitySelect: boolean
    defaultType: "all" | "good" | "service"
  }

  categoryCards: HomeCategoryCard[]
}

2.2 Контракт карточки категории

type HomeCategoryCard = {
  id: string

  title: string
  hint?: string

  icon_key?: string // ключ для подстановки иконки на фронте (без SVG из бэка)

  href: string // готовая ссылка в /catalog с параметрами

  priority?: "primary" | "secondary" // влияет на положение ближе/дальше от центра
  weight?: number // сортировка в пределах priority

  domain?: "good" | "service" | "all" // для аналитики/фильтра (не логика)
  analytics_id?: string
}

Важно: href приходит готовым.
UI не должен собирать query params на лету (кроме поиска).

3) Правила формирования href (Intent rules)

3.1 Примеры

/catalog?type=good&category=electronics

/catalog?type=service&category=repair

/catalog?type=good&brand=nike (если хотите на главной “вход в бренд”)

3.2 Запрещено

- ссылаться на category_id
- формировать сложные “умные” фильтры на UI
- подмешивать tenant в url на UI, если он решается иначе (зависит от вашей tenancy-модели)

4) Порядок, приоритеты и раскладка (для layout engine)

UI должен уметь разложить карточки, не зная бизнес-смысла.

4.1 Логика порядка

priority=primary — ближе к центру

weight — сортировка

fallback — порядок массива

4.2 Ограничения

- первичных карточек: 2–4
- всего карточек: 6–12 (v1 рекомендуем 8)

5) Состояния загрузки (Loading states)

5.1 Если HomeModel ещё не готов

HeroSearch рендерится сразу (статично)

FloatingCategoryCloud показывает skeleton cards (6–8)

5.2 Если categoryCards пустой

облако скрыть

вместо него показать “Popular searches / Tips” (опционально)

6) Локализация (i18n)

6.1 Где должны жить строки

Hero текст: i18n

Title/Hint карточек:

вариант A (рекомендую): уже локализованные строки в HomeModel

вариант B: ключи title_key, hint_key + i18n на фронте

v1 рекомендация: A (проще и быстрее).

7) Версионирование контракта

Чтобы не ломать фронт при изменениях:

type HomeModel = {
  version: "1"
  ...
}

Если меняется структура:

увеличиваем version

фронт поддерживает 1–2 версии

8) Definition of Done (Data Contracts)

Готово, если:

- Home можно полностью отрисовать из HomeModel
- UI не содержит бизнес-логики формирования карточек
- multi-tenant допускает разные витрины
- есть ограничения по количеству и приоритетам
- поддержан loading/skeleton сценарий

Конец документа
