# Тестовые URL для парсинга по категориям

## Структура

Для каждой категории нужно найти один валидный URL товара на gigatron.rs.

## Категории

### 1. Mobilni telefoni (Мобильные телефоны)
**Категория ID:** `bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb`  
**Slug:** `mobilni-telefoni`

**Примеры URL для поиска:**
- iPhone: `https://gigatron.rs/mobilni-telefoni/apple-iphone-15-128gb-black-mtp03zda-573380`
- Samsung: `https://gigatron.rs/mobilni-telefoni/samsung-galaxy-s24-128gb-black-sm-s921bzkgeue`
- Xiaomi: `https://gigatron.rs/mobilni-telefoni/xiaomi-redmi-note-13-pro-256gb-black`

**Как найти:**
1. Открой https://gigatron.rs
2. Перейди в раздел "Мобилни телефони"
3. Выбери любой товар
4. Скопируй URL из адресной строки

---

### 2. Laptopovi (Ноутбуки)
**Категория ID:** `cccccccc-cccc-cccc-cccc-cccccccccccc`  
**Slug:** `laptopovi`

**Примеры URL для поиска:**
- Lenovo: `https://gigatron.rs/laptopovi/lenovo-ideapad-3-15-82h7000vra`
- HP: `https://gigatron.rs/laptopovi/hp-15s-eq2000nm-6k7k1ea`
- ASUS: `https://gigatron.rs/laptopovi/asus-vivobook-15-x1504za-nj024w`

**Как найти:**
1. Открой https://gigatron.rs
2. Перейди в раздел "Лаптопови"
3. Выбери любой товар
4. Скопируй URL из адресной строки

---

## Проверка URL

Перед использованием URL убедитесь:
1. ✅ Страница доступна (не 404)
2. ✅ Товар есть в наличии
3. ✅ Цена отображается на странице
4. ✅ Есть изображение товара

## Тестирование

После добавления URL в скрипт, запусти парсинг:

```powershell
# Для мобильных телефонов
go run cmd/worker/main.go -url "URL_ДЛЯ_ТЕЛЕФОНА" -shop "shop-001"

# Для ноутбуков
go run cmd/worker/main.go -url "URL_ДЛЯ_НОУТБУКА" -shop "shop-001"
```

---

## Обновление URL

Если URL перестал работать:
1. Найди новый валидный URL
2. Обнови его в `backend/scripts/seed_test_urls.sql`
3. Перезапусти парсинг

