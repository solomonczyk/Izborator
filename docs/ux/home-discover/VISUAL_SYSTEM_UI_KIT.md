Visual System & UI Kit (v1.0)

0) Принципы визуальной системы

Спокойный, нейтральный, “умный” интерфейс

UI не кричит → пользователь думает

Контент важнее декора

Все состояния читаемы без анимаций

1) Цветовая система (Color System)

1.1 Base colors

Назначение	Token	Пример
Background	--bg-default	#F8FAFC
Surface	--bg-surface	#FFFFFF
Border	--border-default	#E2E8F0
Text Primary	--text-primary	#0F172A
Text Secondary	--text-secondary	#475569

1.2 Brand / Accent

Назначение	Token	Пример
Primary	--accent-primary	#2563EB
Primary Hover	--accent-primary-hover	#1D4ED8
Focus Ring	--accent-focus	#93C5FD

Один основной акцент.
Никаких “вторых брендов” в v1.

1.3 Semantic colors

Назначение	Token
Success	--color-success
Warning	--color-warning
Error	--color-error

Используются строго для состояний, не для декора.

2) Типографика (Typography)

2.1 Базовый шрифт

Sans-serif

Inter / system-ui / fallback

Читаемый, нейтральный

font-family: Inter, system-ui, -apple-system, sans-serif;

2.2 Типографическая шкала

Назначение	Размер	Вес
H1	32–36px	600
H2	24–28px	600
H3	20–22px	600
Body	16px	400
Small	14px	400
Caption	12px	400

2.3 Правила

Не более 3 размеров на одном экране

Межстрочный интервал: 1.4–1.6

Заголовки без CAPS

3) Сетка и отступы (Layout & Spacing)

3.1 Базовая сетка

Desktop: 12 колонок

Max width: 1280–1440px

Центрирование контейнера

3.2 Spacing scale

Используем 8px scale:

4 / 8 / 12 / 16 / 24 / 32 / 48 / 64

4) Формы и инпуты (Forms)

4.1 Input (Search, Text)

Height: 56px

Radius: 12–16px

Border: 1–2px

Focus:

border + focus ring

обязательно видимый

Состояния:

default

hover

focus

disabled

error

4.2 Button

Типы:

Primary

Secondary

Ghost (редко)

Primary Button:

Height: 48–56px

Radius: 12px

Цвет: --accent-primary

5) Карточки (CategoryCard)

5.1 Визуальные параметры

Background: --bg-surface

Border: 1px --border-default

Radius: 16–24px

Shadow: soft → medium on hover

5.2 Состояния

Idle

Hover (усиление)

Focus (outline, без движения)

Active (press)

6) Иконография

Stroke icons (outline)

Толщина: 1.5–2px

Один стиль на всё приложение

Иконка — опциональна, не обязательна

7) Тени (Elevation)

Уровень	Использование
Shadow 1	Карточки
Shadow 2	Hover
Shadow 3	Dropdown / Popover

Никаких “матовых” или цветных теней.

8) Анимации (связь с Motion Spec)

Default duration: 120–180ms

Easing: ease-out

Только transform и opacity

Все анимации отключаемы через prefers-reduced-motion

9) Dark mode (v1 — опционально)

Поддержка через токены

Без отдельного дизайна

Контраст ≥ WCAG AA

10) Definition of Done (Visual System)

Готово, если:

- все цвета — через токены
- типографика едина
- карточки, кнопки, инпуты выглядят согласованно
- нет “случайных” стилей
- можно расширять без редизайна

Конец документа
