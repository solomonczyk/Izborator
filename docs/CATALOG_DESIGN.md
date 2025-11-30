# CATALOG_DESIGN — Ядро каталога для любого товара (Serbia-first)

## Цели

- Поддержка любых типов товаров на сербском рынке (техника, еда, одежда, стройматериалы, услуги).
- Сохранение текущего pipeline (scraper → raw_products → processor → products + prices).
- Расширяемость через данные (категории, типы, атрибуты), а не через новые таблицы под каждую нишу.

## Принципы

1. **Product (Товар)** — абстрактная вещь: телефон, молоко, шуруп, кроссовки, страховой полис.
2. **Offer (Предложение)** — цена товара в конкретном магазине/городе/канале (реализовано как `product_prices`).
3. **Category / ProductType / Attributes** — описывают *что это за товар* и *какие у него параметры*.

## Новые сущности

### 1. Categories (Иерархия категорий)

Иерархия категорий (разделы → категории → подкатегории).

```sql
CREATE TABLE categories (
    id          UUID PRIMARY KEY,
    parent_id   UUID NULL REFERENCES categories(id) ON DELETE SET NULL,
    slug        TEXT NOT NULL UNIQUE,
    code        TEXT NOT NULL UNIQUE,
    name_sr     TEXT NOT NULL,
    name_sr_lc  TEXT NOT NULL,
    level       SMALLINT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order  INTEGER NOT NULL DEFAULT 100
);
```

**Уровни:**
- Level 1: Разделы (ELEKTRONIKA, HRANA_PICE, MODA, SPORT, DOM_BASTA, AUTO_MOTO)
- Level 2: Категории (PHONES, LAPTOPS, MILK, SNEAKERS)
- Level 3: Подкатегории (SMARTPHONES, FEATURE_PHONES)

В `products`:
```sql
ALTER TABLE products
    ADD COLUMN category_id UUID NULL REFERENCES categories(id);
```

Поле `category` (строка) **оставляем как сырой тег от парсера**, а `category_id` — нормализованная таксономия.

### 2. Product Types (Типы товаров)

Категория — это «где в дереве каталога лежит», **ProductType** — «какие поля вообще есть у этого вида товара».

```sql
CREATE TABLE product_types (
    id          UUID PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    name_sr     TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);
```

**Примеры:**
- `SMARTPHONE` → атрибуты: RAM, Storage, ScreenSize, OS
- `MILK` → жирность, объём, тип (kravlje/kozje), UHT/не-UHT
- `SNEAKERS` → размер, цвет, пол, тип (running/casual)

Связь с категориями (многие-ко-многим):
```sql
CREATE TABLE category_product_types (
    category_id     UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    product_type_id UUID NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    PRIMARY KEY (category_id, product_type_id)
);
```

В `products`:
```sql
ALTER TABLE products
    ADD COLUMN product_type_id UUID NULL REFERENCES product_types(id);
```

### 3. Attributes (Универсальные атрибуты)

Универсальная метамодель атрибутов, значения хранятся в `specs_json`.

**Справочник атрибутов:**
```sql
CREATE TABLE attributes (
    id              UUID PRIMARY KEY,
    code            TEXT NOT NULL UNIQUE,
    name_sr         TEXT NOT NULL,
    data_type       TEXT NOT NULL,   -- "int", "float", "string", "bool", "enum"
    unit_sr         TEXT NULL,
    is_filterable   BOOLEAN NOT NULL DEFAULT TRUE,
    is_sortable     BOOLEAN NOT NULL DEFAULT FALSE
);
```

**Связь ProductType ↔ Attributes:**
```sql
CREATE TABLE product_type_attributes (
    product_type_id UUID NOT NULL REFERENCES product_types(id) ON DELETE CASCADE,
    attribute_id    UUID NOT NULL REFERENCES attributes(id) ON DELETE CASCADE,
    is_required     BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      INTEGER NOT NULL DEFAULT 100,
    PRIMARY KEY (product_type_id, attribute_id)
);
```

**Хранение значений:**
Значения атрибутов хранятся в `products.specs_json`, ключи = `attributes.code`.

**Примеры:**

Телефон:
```json
{
  "RAM": 8,
  "INTERNAL_STORAGE": 256,
  "SCREEN_SIZE": 6.7,
  "OS": "Android"
}
```

Молоко:
```json
{
  "FAT_PERCENT": 2.8,
  "VOLUME_L": 1,
  "MILK_TYPE": "kravlje"
}
```

### 4. Cities (География)

```sql
CREATE TABLE cities (
    id          UUID PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE,
    name_sr     TEXT NOT NULL,
    region_sr   TEXT NULL,
    sort_order  INTEGER NOT NULL DEFAULT 100,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);
```

В `shops`:
```sql
ALTER TABLE shops
    ADD COLUMN default_city_id UUID NULL REFERENCES cities(id);
```

В `product_prices`:
```sql
ALTER TABLE product_prices
    ADD COLUMN city_id UUID NULL REFERENCES cities(id);
```

**Сценарии:**
- Если товар доступен по всей стране → `city_id IS NULL`
- Если цена/наличие отличаются по городам — отдельные `product_prices` на каждый `city_id`

## Изменения в pipeline

### Scraper → raw_products

Сейчас уже сохраняем:
- `name`, `brand`, `category` (сырая строка), `specs_json`, `price`, `currency`, `shop_id`, `external_id`

Для универсального ядра:
- `category_raw` → поле `category` (как сейчас)
- `city` пока можно не парсить (часто не нужно — для онлайн магазинов это страна)

### Processor

В `processor` добавляем 2 новые компонента:

1. **CategoryMapper**
   - Вход: `raw_product.name + raw_product.category + shop_id`
   - Выход: `category_id`, `product_type_id`
   - На старте это может быть простая таблица `category_mapping`

2. **SpecsNormalizer**
   - Вход: `raw_product.specs_json`, `product_type_id`
   - Смотрит в `product_type_attributes` и `attributes`
   - Приводит ключи/формат к единому виду (`"ram"`, `"RAM"` → `RAM`, `"8GB"` → `8`)

Итог: любой товар пройдёт через pipeline, просто разные типы будут иметь разные наборы атрибутов.

## Serbia-first

1. **Справочники заполняем под Сербию:**
   - `categories` — с реальными сербскими названиями и slug'ами
   - `cities` — только города Сербии
   - `product_types` — под те сегменты, которые реально есть на рынке
   - `attributes` — тоже с `name_sr` и `unit_sr`

2. **Язык ядра — сербский (латиница), без перевода:**
   - названия в справочниках: `name_sr`
   - slug'и — латиница, по-сербски: `bela-tehnika`, `patike`, `mleko`

3. **Маркетинг строится поверх этого:**
   - коллекции «Telefoni do 20.000 RSD», «Popularno u Beogradu», «Akcije nedelje»
   - просто используют `category_id`, `city_id`, `price` и атрибуты

## Миграционный путь

1. Этап 1: Categories (текущий)
   - Создать `categories` с улучшенной структурой
   - Добавить `category_id` в `products`
   - Обновить processor для маппинга

2. Этап 2: Product Types & Attributes
   - Создать `product_types`, `attributes`, связи
   - Добавить `product_type_id` в `products`
   - Реализовать SpecsNormalizer

3. Этап 3: Cities
   - Создать `cities`
   - Добавить `city_id` в `shops` и `product_prices`
   - Обновить API для фильтрации по городам

4. Этап 4: CategoryMapper
   - Создать таблицу `category_mapping`
   - Интегрировать в processor
   - Заполнить маппинги для существующих магазинов

