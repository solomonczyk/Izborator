# Izborator — План разработки проекта

## Этап 1 — Core Pipeline (Завершено)
1. Реализация парсеров магазинов ✔  
2. Сохранение сырых данных в raw_products ✔  
3. Processor: matching + нормализация ✔  
4. Products & Prices: модели и хранилище ✔  

## Этап 2 — Public API (Завершено)
5. Endpoint: GET /api/v1/products/{id} ✔  
6. Search API через Meilisearch ✔  
7. Browse API с фильтрами и сортировкой ✔  

## Этап 3 — Frontend Core (Завершено)
8. Страница каталога (Next.js): /catalog ✔  
9. Страница товара: /product/[id] ✔  
10. Главная страница ✔  

## Этап 3.5 — Multilingual Support (i18n) (Текущий)
11. Backend i18n модуль (internal/i18n/)  
12. Middleware для определения языка  
13. Frontend i18n (next-intl)  
14. Локализация UI строк (5 языков: sr, ru, hu, en, zh)  
15. Автоматический выбор языка (Accept-Language, query param, URL)  

## Этап 4 — Функции сравнения и аналитики
11. Price-history API  
12. Графики изменений цены  
13. Сравнение товаров  

## Этап 5 — Инфраструктура
14. Очередь задач (RabbitMQ / Redis Queue)  
15. Планировщик (Cron worker)  
16. Мониторинг & логирование (Grafana + Loki)  

## Этап 6 — Production-ready
17. Docker-compose прод конфигурация  
18. Kubernetes манифесты  
19. CI/CD (GitHub Actions)  
