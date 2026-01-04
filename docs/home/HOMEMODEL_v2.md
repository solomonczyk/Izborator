# HomeModel v2 — Data Contract

Цель: определить финальный контракт данных для главной страницы.
HomeModel v2 — единственный источник данных для Home UI.

---

## Источники данных

- canonical_tree_v1.json
- FEATURED_RULES.md
- tenant / locale context

UI **не читает** каноническое дерево напрямую.

---

## Структура HomeModel

```ts
type HomeModelV2 = {
  version: "2"
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

  featuredCategories: FeaturedCategory[]

  searchPresets?: SearchPreset[]
}

FeaturedCategory
type FeaturedCategory = {
  category_id: string
  title: string
  href: string

  priority: "primary" | "secondary"
  order: number

  icon_key?: string
}

SearchPreset (опционально, v2+)
type SearchPreset = {
  label: string
  query: string
  type?: "good" | "service"
  category_id?: string
}

Правила формирования HomeModel

Backend:

читает canonical tree

применяет FEATURED_RULES

формирует href

Frontend:

рендерит

не принимает решений

Ошибки и fallback

Если featuredCategories пуст:

Home показывает только поиск

Если часть категорий недоступна:

пропускаем, порядок сохраняем

Версионирование

Любое несовместимое изменение → новая версия

UI поддерживает 1 активную версию

Definition of Done

HomeModel v2 считается готовым, если:

Home UI может отрисоваться полностью

UI не знает о дереве категорий

можно менять Featured без изменений UI
