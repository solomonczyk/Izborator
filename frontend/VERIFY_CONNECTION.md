# ✅ Проверка подключения к API

## Backend API работает ✅

Health check показывает: `{"status":"ok"}` на `http://localhost:3002/api/health`

## Проверка фронтенда

### 1. Проверьте Network tab в DevTools

1. Откройте DevTools (F12)
2. Перейдите на вкладку **Network**
3. Обновите страницу каталога (F5)
4. Найдите запросы к API:
   - `/api/v1/products/browse`
   - `/api/v1/categories/tree`
   - `/api/v1/cities`

**Правильные запросы должны быть:**
```
GET http://localhost:3002/api/v1/products/browse?...
GET http://localhost:3002/api/v1/categories/tree?lang=sr
GET http://localhost:3002/api/v1/cities?lang=sr
```

**Неправильные запросы (если видите):**
```
GET http://localhost:8081/api/v1/...  ❌
```

### 2. Проверьте статус ответов

В Network tab проверьте:
- **Status:** должен быть `200 OK` (или `404` если нет данных)
- **Response:** должен содержать JSON данные

### 3. Проверьте переменные окружения в браузере

В Console (F12) выполните:
```javascript
// Проверка API base URL
console.log('API Base:', process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:3002')
```

Должно вывести: `http://localhost:3002`

### 4. Проверьте страницу каталога

Откройте: `http://localhost:3003/sr/catalog`

**Ожидаемое поведение:**
- ✅ Нет ошибки "fetch failed" с портом 8081
- ✅ Выпадающие списки "Категорија" и "Град" загружаются
- ✅ Если есть товары - они отображаются
- ✅ Если товаров нет - показывается "Ништа није пронађено"

## Тестирование API endpoints напрямую

### Категории
```powershell
Invoke-WebRequest -Uri "http://localhost:3002/api/v1/categories/tree?lang=sr" -Method GET | Select-Object -ExpandProperty Content
```

### Города
```powershell
Invoke-WebRequest -Uri "http://localhost:3002/api/v1/cities?lang=sr" -Method GET | Select-Object -ExpandProperty Content
```

### Каталог товаров
```powershell
Invoke-WebRequest -Uri "http://localhost:3002/api/v1/products/browse?page=1&per_page=10" -Method GET | Select-Object -ExpandProperty Content
```

## Если все еще видите порт 8081

1. **Остановите фронтенд** (Ctrl+C)
2. **Очистите кэш:**
   ```powershell
   cd frontend
   Remove-Item -Recurse -Force .next
   ```
3. **Создайте/проверьте `.env.local`:**
   ```
   NEXT_PUBLIC_API_BASE=http://localhost:3002
   PORT=3003
   ```
4. **Перезапустите:**
   ```powershell
   npm run dev
   ```

## Ошибки content.js

Ошибки `content.js:1 Uncaught (in promise) The message port closed before a response was received` - это **нормально**. Они связаны с расширениями браузера и не влияют на работу приложения.

Их можно игнорировать или отключить расширения для тестирования.

