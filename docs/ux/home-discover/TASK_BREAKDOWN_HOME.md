Home / Discover — Task List (v1.0)

ЭПИК H-01: Home / Discover (UX-driven)
🔹 EPIC H-01.1 — Page Skeleton
H-01.1.1 Создать страницу Home

Путь: /[locale]/

Добавить базовый layout (Header / Main / Footer)

DoD

Страница рендерится

Header и Footer подключены

Нет бизнес-логики

H-01.1.2 Safe Center (Hero container)

Зафиксировать центральный контейнер

Задать размеры Safe Center

DoD

Контейнер всегда по центру

Ничего не перекрывает его

🔹 EPIC H-01.2 — HeroSearch
H-01.2.1 Компонент HeroSearch

Input

Кнопка “Найти”

DoD

Enter работает

Кнопка работает

Keyboard OK

H-01.2.2 TypeToggle (feature-flag)

goods / services / all

Без fetch

DoD

Toggle работает

Не ломает submit

H-01.2.3 CitySelect (feature-flag)

Опционально

Только UI

DoD

Можно скрыть флагом

Не влияет на остальное

🔹 EPIC H-01.3 — CategoryCard (атом)
H-01.3.1 CategoryCard UI

Title

Hint

Link

DoD

Вся карточка кликабельна

Focus / hover / active есть

H-01.3.2 A11y для карточки

tabIndex

aria-label

focus-ring

DoD

Keyboard-навигация проходит

🔹 EPIC H-01.4 — FloatingCategoryCloud
H-01.4.1 Layout Engine (без motion)

Раскладка по Layout Rules

Ограничение количества карточек

DoD

Карточки не перекрывают центр

Позиции стабильны

H-01.4.2 Responsive behavior

Desktop: cloud

Mobile: grid

DoD

Mobile без hover

Grid читаемый

🔹 EPIC H-01.5 — Motion Layer
H-01.5.1 Подключить animation tokens

Duration

Easing

Amplitude

DoD

Нет магических чисел

H-01.5.2 Proximity / Hover / Active

Реализация по spec

DoD

Нет слежения за курсором

FPS стабильный

H-01.5.3 Reduced Motion

prefers-reduced-motion

DoD

Все движения отключаются

🔹 EPIC H-01.6 — Data Wiring
H-01.6.1 Подключить /api/v1/home

Fetch HomeModel

Передача tenant / locale

DoD

UI рендерится из данных

Нет логики на фронте

H-01.6.2 Loading / Skeleton

Skeleton cards

Нет layout shift

DoD

Поиск доступен сразу

🔹 EPIC H-01.7 — QA & Accessibility
H-01.7.1 QA checklist

Пройти все AC из Шага 9

DoD

Нет блокеров

H-01.7.2 Cross-device

Desktop

Mobile

Keyboard

DoD

Все UX-флоу закрыты

🔹 EPIC H-01.8 — Performance & Polish
H-01.8.1 Perf audit

Main thread

Re-render

DoD

Нет лагов

H-01.8.2 Visual polish

Отступы

Контраст

Тени

DoD

UI выглядит цельно

📌 Ключевая точка остановки (MVP)

Можно запускаться после:

EPIC H-01.6 (Data Wiring)

Дальше — качество, не функционал.
