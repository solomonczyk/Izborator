# 📁 Структура проекта Izborator

## 🗂 Общая структура

```
Izborator/
├── 📄 README.md                    # Главная страница проекта
├── 📄 LICENSE                      # Лицензия
│
├── 📂 backend/                     # Go Backend
│   ├── 📂 cmd/                     # Точки входа приложений
│   │   ├── api/                    # HTTP API сервер
│   │   ├── worker/                 # Worker для обработки задач
│   │   ├── indexer/                # Индексатор для Meilisearch
│   │   └── migrate/                # Миграции БД
│   │
│   ├── 📂 internal/                # Внутренние модули
│   │   ├── app/                    # Инициализация приложения
│   │   ├── config/                 # Конфигурация
│   │   ├── logger/                 # Логирование
│   │   │
│   │   ├── 📂 http/                # HTTP слой
│   │   │   ├── handlers/           # HTTP handlers
│   │   │   ├── middleware/         # Middleware
│   │   │   └── router/             # Роутинг
│   │   │
│   │   ├── 📂 storage/             # Адаптеры хранилища
│   │   │   ├── postgres.go         # PostgreSQL
│   │   │   ├── redis.go            # Redis
│   │   │   ├── meilisearch.go      # Meilisearch
│   │   │   └── *_adapter.go        # Адаптеры для модулей
│   │   │
│   │   └── 📂 [modules]/           # Бизнес-логика модулей
│   │       ├── models.go           # Модели данных
│   │       ├── module.go           # Интерфейсы
│   │       ├── impl.go             # Реализация
│   │       ├── errors.go           # Ошибки модуля
│   │       └── *_test.go           # Тесты
│   │
│   ├── 📂 migrations/              # SQL миграции
│   │   ├── 0001_*.up.sql          # Миграции вверх
│   │   ├── 0001_*.down.sql        # Миграции вниз
│   │   └── README.md
│   │
│   ├── 📂 scripts/                 # Вспомогательные скрипты
│   │   ├── seed_*.sql             # Seed данные
│   │   ├── setup_*.ps1            # Скрипты настройки
│   │   └── README.md
│   │
│   ├── 📄 go.mod                   # Go зависимости
│   ├── 📄 go.sum                   # Go checksums
│   ├── 📄 env.example              # Пример .env файла
│   └── 📄 README.md                # Backend документация
│
├── 📂 frontend/                    # Next.js Frontend
│   ├── 📂 app/                     # Next.js App Router
│   ├── 📂 components/              # React компоненты
│   ├── 📂 messages/                # i18n переводы
│   ├── 📂 public/                  # Статические файлы
│   ├── 📄 package.json             # NPM зависимости
│   └── 📄 README.md                # Frontend документация
│
├── 📂 docs/                        # Документация
│   ├── README.md                   # Индекс документации
│   └── CATALOG_DESIGN.md           # Дизайн каталога
│
├── 📄 docker-compose.yml           # Docker Compose конфигурация
├── 📄 .gitignore                   # Git ignore правила
│
└── 📄 [документация].md            # Документы в корне
    ├── STRATEGY.md                 # Стратегия проекта
    ├── PLAN.md                     # План разработки
    ├── DEVELOPMENT_LOG.md           # Дневник разработки
    ├── IMPROVEMENTS.md             # Список улучшений
    ├── START_COMMANDS.md           # Команды запуска
    ├── STATUS.md                   # Статус проекта
    └── ...
```

## 📋 Описание директорий

### Backend (`backend/`)

#### `cmd/` - Точки входа
- **api/** - HTTP API сервер (REST API)
- **worker/** - Worker для обработки задач парсинга и обработки товаров
- **indexer/** - Индексатор для Meilisearch
- **migrate/** - CLI для миграций БД

#### `internal/` - Внутренние модули

**HTTP слой:**
- `http/handlers/` - HTTP handlers для endpoints
- `http/middleware/` - Middleware (CORS, logging, cache, recovery)
- `http/router/` - Настройка роутинга

**Storage адаптеры:**
- `storage/postgres.go` - PostgreSQL подключение
- `storage/redis.go` - Redis подключение
- `storage/meilisearch.go` - Meilisearch подключение
- `storage/*_adapter.go` - Адаптеры для каждого модуля

**Бизнес-логика модулей:**
Каждый модуль содержит:
- `models.go` - Структуры данных
- `module.go` - Интерфейсы (Storage, Service)
- `impl.go` - Реализация бизнес-логики
- `errors.go` - Ошибки модуля
- `*_test.go` - Unit тесты

**Модули:**
- `products/` - Управление товарами
- `scraper/` - Парсинг товаров с сайтов
- `processor/` - Обработка сырых данных
- `matching/` - Сопоставление товаров
- `pricehistory/` - История цен
- `categories/` - Категории товаров
- `cities/` - Города
- `attributes/` - Атрибуты товаров
- `producttypes/` - Типы товаров
- `scrapingstats/` - Статистика парсинга

#### `migrations/` - SQL миграции
- Нумерованные файлы миграций (0001, 0002, ...)
- `.up.sql` - применение миграции
- `.down.sql` - откат миграции

#### `scripts/` - Вспомогательные скрипты
- SQL seed скрипты
- PowerShell скрипты настройки
- SQL скрипты для исправлений

### Frontend (`frontend/`)

#### `app/` - Next.js App Router
- `[locale]/` - Локализованные страницы
- `catalog/` - Каталог товаров
- `product/[id]/` - Страница товара

#### `components/` - React компоненты
- Переиспользуемые UI компоненты

#### `messages/` - i18n переводы
- JSON файлы с переводами для разных языков

### Документация (`docs/`)

- `README.md` - Индекс всей документации
- `CATALOG_DESIGN.md` - Дизайн каталога

## 🚫 Файлы, которые НЕ должны быть в репозитории

Следующие файлы игнорируются через `.gitignore`:

- `*.exe` - Скомпилированные бинарники
- `.env` - Переменные окружения
- `node_modules/` - NPM зависимости
- `.next/` - Next.js build
- `*.log` - Логи
- `*.tmp` - Временные файлы
- `docker-data/` - Docker volumes

## 📝 Соглашения по именованию

### Go файлы
- `models.go` - Модели данных
- `module.go` - Интерфейсы
- `impl.go` - Реализация
- `errors.go` - Ошибки
- `*_test.go` - Тесты
- `*_adapter.go` - Storage адаптеры

### SQL файлы
- `000N_*.up.sql` - Миграция вверх
- `000N_*.down.sql` - Миграция вниз
- `seed_*.sql` - Seed данные

### Документация
- `README.md` - Описание модуля/директории
- `*.md` - Документация в Markdown

## 🔄 Workflow

1. **Разработка:** Создаём модули в `internal/`
2. **Миграции:** Добавляем в `migrations/`
3. **Тесты:** Пишем рядом с кодом `*_test.go`
4. **Документация:** Обновляем `DEVELOPMENT_LOG.md`

## 📚 Дополнительная информация

- Полная документация: [docs/README.md](./docs/README.md)
- Статус проекта: [STATUS.md](./STATUS.md)
- Команды запуска: [START_COMMANDS.md](./START_COMMANDS.md)

