UI Component Specification (v1.0)
Контракт между UX / Frontend / Backend

1) Общие принципы компонентов

Компоненты:

- презентационные
- без бизнес-логики

Все данные:

- приходят через props
- сериализуемы

Навигация:

- через URL-параметры

Компоненты не знают:

- о фасетах
- о schema
- о tenant-логике

2) HeroSearch — центральный поиск

Назначение

Единая точка фокуса и основной сценарий входа в каталог.

Props (TypeScript)

type HeroSearchProps = {
  initialQuery?: string
  initialType?: "good" | "service" | "all"
  initialCity?: string

  showCitySelect: boolean
  showTypeToggle: boolean

  onSubmit: (params: {
    query?: string
    type?: "good" | "service"
    city?: string
  }) => void
}

Поведение

Enter в input → submit

Кнопка “Найти” → submit

Submit:

формирует URL

делает router.push("/catalog?...")

UX-требования

Input всегда доступен

Placeholder: “Что вы ищете?”

Минимальная высота: 56px

Focus-ring обязателен

Не делает

не валидирует бизнес-правила

не знает tenant

не дергает API

3) FloatingCategoryCloud

Назначение

Визуальный контейнер для карточек категорий.

Props

type FloatingCategoryCloudProps = {
  categories: CategoryCardProps[]
  maxVisible?: number // default 8
  reducedMotion?: boolean
}

Ответственность

позиционирование карточек

передача координат карточкам

НЕ обработка кликов

4) CategoryCard

Назначение

Один навигационный entry-point в каталог.

Props

type CategoryCardProps = {
  id: string

  title: string
  hint?: string
  icon?: ReactNode

  href: string

  priority?: "primary" | "secondary"

  analyticsId?: string
}

Поведение

Hover:

визуальное усиление

Click / Enter:

переход по href

Focus:

outline / halo

без движения

Ограничения

max 2 строки текста

кликабельна вся карточка

cursor: pointer

5) TypeToggle

Назначение

Переключение домена (goods / services).

Props

type TypeToggleProps = {
  value: "good" | "service" | "all"
  onChange: (value: "good" | "service" | "all") => void
}

Поведение

Toggle / segmented control

Keyboard доступен

Не триггерит fetch

6) Навигационный контракт

Категории
/catalog?type=good&category=electronics

Только тип
/catalog?type=service

Поиск
/catalog?q=iphone&type=good

7) Аналитика (v1 — пассивная)

Компоненты могут принимать analyticsId, но:

- не логируют сами
- не знают, куда логируется

8) Error states (UX)

Нет категорий → Cloud не рендерится

Нет type → скрыть toggle

Нет city → скрыть select

9) Definition of Done (Components)

Готово, если:

- каждый компонент можно отрендерить из mock-данных
- props покрывают все сценарии
- нет скрытых зависимостей
- можно Storybook’ить без бэка

Конец документа
