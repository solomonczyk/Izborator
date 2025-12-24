# 🧪 Быстрое тестирование API на сервере

## ✅ Проверка кода (выполнена)

**Проверено:**
- ✅ Browse handler правильно обрабатывает category slug → category_id
- ✅ Получает дочерние категории для фильтрации
- ✅ BrowseResult имеет правильную структуру: `items`, `total`, `page`, `per_page`, `total_pages`
- ✅ Фильтрация работает через Meilisearch и PostgreSQL fallback
- ✅ Fallback при несуществующей категории обрабатывается корректно

## 🚀 Команда для выполнения на сервере

**Одна команда для всех тестов:**

```bash
docker-compose exec backend sh -c "API='http://backend:8080' && echo '=== 1. Health ===' && curl -s \$API/api/health && echo -e '\n\n=== 2. Browse (без фильтра) ===' && curl -s \$API/api/v1/products/browse?page=1&per_page=2 && echo -e '\n\n=== 3. Browse (category=mobilni-telefoni) ===' && curl -s \$API/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=2 && echo -e '\n\n=== 4. Browse (category=laptopovi) ===' && curl -s \$API/api/v1/products/browse?category=laptopovi&page=1&per_page=2 && echo -e '\n\n=== 5. Browse (несуществующая категория) ===' && curl -s \$API/api/v1/products/browse?category=neexistujuca&page=1&per_page=2 && echo -e '\n\n✅ Готово!'"
```

**Или по шагам:**

```bash
# 1. Health check
docker-compose exec backend sh -c "curl -s http://backend:8080/api/health"

# 2. Browse без фильтра
docker-compose exec backend sh -c "curl -s 'http://backend:8080/api/v1/products/browse?page=1&per_page=2'"

# 3. Browse с категорией mobilni-telefoni
docker-compose exec backend sh -c "curl -s 'http://backend:8080/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=2'"

# 4. Browse с категорией laptopovi
docker-compose exec backend sh -c "curl -s 'http://backend:8080/api/v1/products/browse?category=laptopovi&page=1&per_page=2'"

# 5. Browse с несуществующей категорией (fallback)
docker-compose exec backend sh -c "curl -s 'http://backend:8080/api/v1/products/browse?category=neexistujuca&page=1&per_page=2'"
```

## 📊 Ожидаемые результаты

1. **Health check:** `{"status":"ok"}`
2. **Browse без фильтра:** JSON с полями `items`, `total`, `page`, `per_page`, `total_pages`
3. **Browse с категорией:** JSON с товарами из указанной категории (или пустой массив, если нет товаров)
4. **Browse с несуществующей категорией:** JSON с пустым массивом `items` или предупреждение в логах

## ✅ После успешного тестирования

1. Обновить `ROADMAP_CURRENT_STEP.md` - отметить выполненные тесты
2. Обновить `DEVELOPMENT_LOG.md` - записать результаты
3. Перейти к следующей задаче из `PLAN.md`

