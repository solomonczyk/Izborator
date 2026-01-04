# Home — SSR & SEO Strategy

Цель: определить поведение главной страницы
с точки зрения SSR, SEO и производительности.

---

## Общие принципы

1. Home всегда SSR
2. Search — главный SEO-инструмент
3. Featured категории — навигационные, не SEO-лендинги
4. Никакой динамической генерации контента на клиенте

---

## SSR поведение

### Что рендерится на сервере
- HeroSearch (полностью)
- FeaturedCategories (карточки)
- Основная разметка страницы

### Что НЕ рендерится на сервере
- Hover / motion эффекты
- Proximity-анимации
- Client-only UX-улучшения

---

## Fallback при ошибках API

Если `/api/v1/home` недоступен:
- SSR рендерит:
  - HeroSearch
  - пустой блок featured
- Никаких ошибок 5xx пользователю
- Ошибка логируется сервером

---

## SEO-стратегия

### Indexing
- Home индексируется
- Featured категории НЕ считаются SEO-лендингами
- Основные SEO-страницы:
  - /catalog
  - /catalog?category=...
  - /product/[id]

---

### Meta tags (Home)

Обязательные:
- `<title>` — общий, брендовый
- `<meta description>` — короткий, не продающий
- `<link rel="canonical">` — на саму Home

Запрещено:
- Динамические title на основе featured
- Keyword stuffing

---

## Structured data

v1:
- Не добавляем schema.org

v2+ (опционально):
- WebSite
- SearchAction (для поиска)

---

## Performance

### TTFB
- Цель: < 300ms (edge)
- Допустимо: < 600ms

### LCP
- HeroSearch — LCP элемент
- Без изображений в первом экране

---

## Caching

- SSR с revalidation
- CDN cache по tenant + locale
- HomeModel кешируется отдельно

---

## Accessibility & SEO

- H1 один
- Семантическая разметка
- Нет скрытого текста
- Reduced motion не влияет на контент

---

## Definition of Done

SSR/SEO стратегия считается готовой, если:
- Home индексируется корректно
- Нет SEO-конфликтов
- SSR не ломается при ошибках API
- Performance метрики в пределах нормы
