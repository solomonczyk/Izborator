# Скрипты для инициализации данных

## seed_gigatron.sql

Добавляет конфигурацию магазина Gigatron в базу данных для тестирования парсинга.

### Использование

```bash
# Через docker exec
docker exec -i izborator_postgres psql -U postgres -d izborator < backend/scripts/seed_gigatron.sql

# Или через psql напрямую
psql -U postgres -d izborator -f backend/scripts/seed_gigatron.sql
```

### Что делает скрипт

- Добавляет магазин Gigatron с ID `a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11`
- Настраивает CSS-селекторы для парсинга:
  - `name`: `h1` - заголовок страницы
  - `price`: `.pp-price-new, .product-price-new` - цена товара
  - `image`: `.pp-img-wrap img` - изображение товара
  - `description`: `.pp-description` - описание товара
  - `brand`: `.pp-brand` - бренд товара
- Устанавливает rate limit: 2 запроса в секунду

### Примечание

Селекторы могут измениться, если сайт обновит верстку. В этом случае нужно обновить JSON в поле `selectors`.

