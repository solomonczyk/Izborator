# Как найти валидный URL для тестирования парсинга

## Метод 1: Через браузер

1. Откройте сайт https://gigatron.rs
2. Найдите любой товар (например, через поиск "iPhone")
3. Откройте страницу товара
4. Скопируйте URL из адресной строки
5. Пример формата: `https://gigatron.rs/mobilni-telefoni/apple-iphone-15-pro-max-256gb-titanium-natural`

## Метод 2: Через API (если доступен)

```bash
# Поиск товаров через поисковик
curl "https://gigatron.rs/search?q=iphone"
```

## Метод 3: Использовать существующий товар из БД

Если в базе уже есть товары, можно использовать их URL:

```sql
SELECT url FROM raw_products WHERE shop_id = 'shop-001' LIMIT 1;
```

## Проверка URL

Перед использованием URL убедитесь, что:
1. Страница доступна (не 404)
2. Товар есть в наличии
3. Цена отображается на странице

## Тестирование

```bash
cd backend
go run cmd/worker/main.go -url "ВАШ_URL_ЗДЕСЬ" -shop "shop-001"
```

Если парсинг успешен, вы увидите:
- ✅ SUCCESS! Product parsed & saved
- Данные товара в логах

