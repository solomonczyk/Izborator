Home / Discover UI — Implementation Plan (v1.0)

0) Принципы реализации

Идём сверху вниз: страница → компоненты → motion → polish

Каждый шаг:

маленький

проверяемый

не ломает архитектуру

Сначала структура и UX, потом анимации

Этап 1 — Каркас страницы (Foundation)
Цель

Получить статичную, но правильную главную страницу.

Задачи

Создать страницу /[locale]/ (Home)

Разметить layout:

Header

HeroSearch (центр)

Placeholder для FloatingCategoryCloud

Footer

Зафиксировать Safe Center (CSS)

Definition of Done

Поиск в центре

Карточки пока без анимаций

Ничего не перекрывается

Работает desktop + mobile layout

Риски

❌ Рано лезть в анимации

❌ Пытаться “сразу красиво”

Этап 2 — HeroSearch (Primary UX)
Цель

Сделать идеальный поиск, который не ломается ни при чём.

Задачи

Реализовать HeroSearch как отдельный компонент

Подключить:

SearchInput

TypeToggle (feature-flag)

CitySelect (feature-flag)

Навигация:

submit → /catalog?...

Definition of Done

Enter работает

Кнопка “Найти” работает

Keyboard-first сценарий закрыт

Риски

❌ Логика в компоненте

❌ Зависимость от API

Этап 3 — CategoryCard (Atomic UI)
Цель

Сделать идеальную карточку без контекста страницы.

Задачи

Реализовать CategoryCard

Состояния:

idle

hover

focus

active

Проверить:

a11y

кликабельность всей карточки

Definition of Done

Карточка работает в Storybook / из mock

Нет “магических” стилей

Props = спецификации

Этап 4 — FloatingCategoryCloud (Layout Engine)
Цель

Собрать карточки вокруг поиска, без движения.

Задачи

Реализовать FloatingCategoryCloud

Раскладка:

positions по правилам Layout Rules

Responsive:

desktop: cloud

mobile: grid

Definition of Done

Карточки не перекрывают центр

Количество ограничено

Позиции стабильны

Риски

❌ Случайный layout

❌ CSS, который “живёт своей жизнью”

Этап 5 — Motion Layer (по токенам)
Цель

Добавить движение как усиление, не как смысл.

Задачи

Подключить animation tokens

Реализовать:

proximity

hover

active

Добавить prefers-reduced-motion

Definition of Done

Motion отключаем

Нет layout shift

FPS стабильный

Риски

❌ Слишком сильные эффекты

❌ Следование за курсором

Этап 6 — Data Wiring (HomeModel)
Цель

Подключить реальные данные без изменения UX.

Задачи

Подключить /api/v1/home

Отрисовать HomeModel

Skeleton states

Definition of Done

Home рендерится из данных

Нет логики “на фронте”

Empty state корректный

Этап 7 — Accessibility & QA
Цель

Закрыть все UX-флоу и критерии.

Задачи

Пройти QA checklist (Шаг 9)

Проверить:

keyboard

mobile

reduced motion

Поправить контраст/фокус

Definition of Done

Все AC выполнены

Нет блокеров

Этап 8 — Performance & Polish
Цель

Довести до production feel.

Задачи

Проверить main-thread

Убрать лишние re-render

Проверить Lighthouse

Definition of Done

Нет лагов

Нет визуального шума

UX “тихий и уверенный”

Итоговая карта этапов
1. Page skeleton
2. HeroSearch
3. CategoryCard
4. Floating layout
5. Motion
6. Data
7. QA
8. Polish

Где можно остановиться и уже “запускаться”

После Этапа 6 продукт уже:

usable

понятный

масштабируемый

Этапы 7–8 — про качество, не про выживание.
