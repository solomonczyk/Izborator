Motion & Interaction Rules (v1.0)

1) Принципы (не нарушать)

Центр стабилен: HeroSearch никогда не смещается и не перекрывается.

Движение — вторично: анимации усиливают, но не объясняют.

Детерминизм: одинаковое движение при одинаковом вводе.

A11y-first: всё работает без hover.

2) Слои (z-index)

Z0: Background

Z1: FloatingCategoryCloud

Z2: HeroSearch (фиксирован)

Z3: Header / Dropdowns

3) Карточки категорий — состояния

Idle → Proximity → Hover → Focus → Active

Idle

scale: 1

opacity: 0.9

shadow: soft

Proximity (курсор в радиусе 120px)

translate: ±6–12px от центра

rotate: ±2–4°

transition: 120–180ms, ease-out

Hover

scale: 1.03

shadow: medium

opacity: 1

cursor: pointer

Focus (keyboard)

без движения

outline/halo visible

Enter = переход

Active (click/tap)

scale: 0.98 (100ms)

затем навигация

4) Алгоритм реакции на курсор (упрощённо)

Рассчитать вектор card → cursor

Нормализовать

Применить translate от центра HeroSearch

Ограничить смещение max 12px

Запрет: следование карточки за курсором.

5) Ограничения производительности

≤ 12 карточек

Только transform + opacity

requestAnimationFrame для трекинга

Throttle 60fps → 30fps при нагрузке

6) Mobile / Touch

Hover отключён

Idle + Active только

Карточки в сетке

Лёгкое “дыхание” (optional): opacity 0.95↔1, 3–4s

7) Reduced Motion

Если prefers-reduced-motion: reduce:

Все translate/rotate = 0

Оставить только focus/hover цвет и тень

8) Ошибки, которые запрещены

Перекрытие HeroSearch

Следование карточек за курсором

Случайные траектории

Анимации, зависящие от FPS

Hover-only контент

9) Definition of Done (Motion)

Нет layout shift

Input всегда доступен

Keyboard полностью рабочий

FPS стабильный

Reduced Motion проходит QA

Конец документа
