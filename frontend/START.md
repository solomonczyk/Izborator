# Инструкция по запуску фронтенда

## Важно: проверьте порт 3000

Если вы видите 404 страницу от другого проекта, значит на порту 3000 запущен другой сервер.

## Решение:

1. **Остановите все процессы на порту 3000:**
   ```powershell
   Get-NetTCPConnection -LocalPort 3000 | Select-Object -ExpandProperty OwningProcess | ForEach-Object { Stop-Process -Id $_ -Force }
   ```

2. **Запустите Next.js сервер:**
   ```bash
   cd frontend
   npm run dev
   ```

3. **Откройте в браузере:**
   - http://localhost:3000/catalog
   - http://localhost:3000/catalog?q=motorola

## Проверка:

Убедитесь, что:
- ✅ Backend API запущен на http://localhost:8080
- ✅ Next.js сервер запущен на http://localhost:3000
- ✅ Файл `.env.local` содержит `NEXT_PUBLIC_API_BASE=http://localhost:8080`

## Если все еще видите 404:

1. Проверьте, что вы открыли правильный URL: `http://localhost:3000/catalog`
2. Проверьте консоль браузера (F12) на наличие ошибок
3. Проверьте логи Next.js сервера в терминале

