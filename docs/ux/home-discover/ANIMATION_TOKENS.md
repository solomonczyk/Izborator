Animation & Motion Tokens (v1.0)

0) Принципы

Один источник истины для motion

Малые амплитуды (UX > эффект)

Детерминизм: одинаковый ввод → одинаковое движение

A11y-first: всё отключаемо

1) Глобальные токены (Design Tokens)
1.1 Время (Duration)
--motion-instant: 80ms;
--motion-fast: 120ms;
--motion-base: 160ms;
--motion-slow: 220ms;


Правила:

Hover/Proximity: fast | base

Click/Press: instant | fast

Появление: base | slow

1.2 Кривые (Easing)
--ease-out-soft: cubic-bezier(0.16, 1, 0.3, 1);
--ease-out-base: cubic-bezier(0.2, 0, 0, 1);
--ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);


Правила:

Hover/Proximity: ease-out-soft

Press/Active: ease-out-base

Entrance/Exit: ease-in-out

1.3 Амплитуды (Amplitude)
--move-xs: 4px;
--move-sm: 8px;
--move-md: 12px;

--scale-hover: 1.03;
--scale-press: 0.98;

--rotate-xs: 2deg;
--rotate-sm: 4deg;


Ограничения:

translate ≤ --move-md

rotate ≤ --rotate-sm

scale ≤ 1.05 (запрещено превышать)

2) Карточки категорий (CategoryCard)
2.1 Idle → Proximity
transform: translate(var(--move-sm)) rotate(var(--rotate-xs));
transition: transform var(--motion-base) var(--ease-out-soft);

2.2 Hover
transform: scale(var(--scale-hover));
transition: transform var(--motion-fast) var(--ease-out-soft);

2.3 Active (Press)
transform: scale(var(--scale-press));
transition: transform var(--motion-instant) var(--ease-out-base);

3) HeroSearch (центр)

Правило:
HeroSearch не анимируется по позиции.

Допустимо:

focus-ring

shadow усиление

placeholder fade (≤ motion-fast)

4) Появление карточек (Entrance)
opacity: 0 → 1;
transition: opacity var(--motion-slow) var(--ease-in-out);


Запрещено:

slide-in

stagger по оси

“разлёт” карточек

5) Reduced Motion
5.1 Триггер
@media (prefers-reduced-motion: reduce) {
  * {
    transition: none !important;
    animation: none !important;
  }
}

5.2 Разрешено при reduce

color

box-shadow

outline

6) Производительность

Только transform + opacity

Все расчёты координат:

через requestAnimationFrame

throttle до 30fps при нагрузке

Запрещено:

top/left

width/height

box-shadow анимация

7) Тестируемость (QA hooks)

Для e2e:

атрибуты data-motion="proximity|hover|active"

можно отключить motion через env:

NEXT_PUBLIC_DISABLE_MOTION=true

8) Definition of Done (Motion Tokens)

Готово, если:

все анимации используют токены

нет “магических чисел” в компонентах

reduce-motion полностью отключает движение

motion легко тюнится из одного места

Конец документа
