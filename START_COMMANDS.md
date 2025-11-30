# Команды для запуска проекта

## 🚀 Быстрый старт

### 1. Backend (API сервер)

**В PowerShell (из корня проекта):**
```powershell
cd backend
$env:DB_HOST="localhost"
$env:DB_PORT="5433"
$env:DB_USER="postgres"
$env:DB_PASSWORD="postgres"
$env:DB_NAME="izborator"
$env:SERVER_PORT="8081"
go run cmd/api/main.go
```

**Или через .env файл (создай `backend/.env`):**
```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=izborator
SERVER_PORT=8081
```

Затем просто:
```powershell
cd backend
go run cmd/api/main.go
```

**Проверка:** http://localhost:8081/api/health

---

### 2. Frontend (Next.js)

**В PowerShell (из корня проекта):**
```powershell
cd frontend
npm run dev
```

**Проверка:** http://localhost:3000

---

## 📝 Полный запуск (оба сервера)

### Вариант 1: Два отдельных терминала

**Терминал 1 (Backend):**
```powershell
cd backend
$env:DB_HOST="localhost"; $env:DB_PORT="5433"; $env:DB_USER="postgres"; $env:DB_PASSWORD="postgres"; $env:DB_NAME="izborator"; $env:SERVER_PORT="8081"
go run cmd/api/main.go
```

**Терминал 2 (Frontend):**
```powershell
cd frontend
npm run dev
```

---

### Вариант 2: Один терминал (фоновые процессы)

**Backend в фоне:**
```powershell
cd backend
Start-Process powershell -ArgumentList "-NoExit", "-Command", "`$env:DB_HOST='localhost'; `$env:DB_PORT='5433'; `$env:DB_USER='postgres'; `$env:DB_PASSWORD='postgres'; `$env:DB_NAME='izborator'; `$env:SERVER_PORT='8081'; go run cmd/api/main.go"
```

**Frontend:**
```powershell
cd frontend
npm run dev
```

---

## 🔍 Проверка работы

1. **Backend Health Check:**
   ```powershell
   Invoke-WebRequest -Uri "http://localhost:8081/api/health" -Method GET
   ```

2. **Frontend:**
   - Открой http://localhost:3000
   - Должен открыться каталог (с редиректом на `/en/catalog`)

3. **API Browse:**
   ```powershell
    Invoke-WebRequest -Uri "http://localhost:8081/api/v1/products/browse?category=mobilni-telefoni" -Method GET | Select-Object -ExpandProperty Content
   ```

---

## ⚠️ Важно

- **PostgreSQL** должен быть запущен (через Docker или локально)
- **Порт 3000** должен быть свободен для фронтенда
- **Порт 8081** должен быть свободен для бэкенда
- Если порты заняты, измени их в конфигурации

---

## 🐳 Если используешь Docker для PostgreSQL

Убедись, что контейнер запущен:
```powershell
docker ps --filter "name=postgres"
```

Если не запущен:
```powershell
docker-compose up -d postgres
```

