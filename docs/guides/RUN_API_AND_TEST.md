# 🚀 Инструкция: Запуск API и тестирование

## Шаг 1: Запуск API сервера

Открой **новый терминал PowerShell** и выполни:

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

**Ожидаемый результат:**
```
{"level":"info","message":"Successfully connected to PostgreSQL"}
{"level":"info","message":"Meilisearch connection established"}
{"level":"info","message":"Redis connection established"}
{"level":"info","port":8081,"message":"Starting API server"}
```

## Шаг 2: Проверка здоровья API

В **другом терминале** выполни:

```powershell
Invoke-WebRequest -Uri "http://localhost:8081/api/health" -Method GET
```

Должен вернуть: `{"status":"ok"}`

## Шаг 3: Запуск тестирования

После того, как API запустился, выполни:

```powershell
.\test-browse-api.ps1
```

Или если скрипт не найден, выполни тесты вручную:

```powershell
# Тест 1: Browse без фильтра
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/products/browse?page=1&per_page=5" -Method GET | Select-Object StatusCode, @{Name='Content';Expression={$_.Content | ConvertFrom-Json | ConvertTo-Json -Depth 3}}

# Тест 2: Browse с категорией mobilni-telefoni
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/products/browse?category=mobilni-telefoni&page=1&per_page=5" -Method GET | Select-Object StatusCode, @{Name='Content';Expression={$_.Content | ConvertFrom-Json | ConvertTo-Json -Depth 3}}

# Тест 3: Browse с категорией laptopovi
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/products/browse?category=laptopovi&page=1&per_page=5" -Method GET | Select-Object StatusCode, @{Name='Content';Expression={$_.Content | ConvertFrom-Json | ConvertTo-Json -Depth 3}}

# Тест 4: Browse с несуществующей категорией (fallback)
Invoke-WebRequest -Uri "http://localhost:8081/api/v1/products/browse?category=neexistujuca-kategorija&page=1&per_page=5" -Method GET | Select-Object StatusCode, @{Name='Content';Expression={$_.Content | ConvertFrom-Json | ConvertTo-Json -Depth 3}}
```

## Шаг 4: Обновление ROADMAP_CURRENT_STEP.md

После успешного тестирования обнови `ROADMAP_CURRENT_STEP.md`, отметив выполненные тесты.

## ⚠️ Если API не запускается

1. **Проверь PostgreSQL:**
   ```powershell
   docker ps --filter "name=postgres"
   ```
   
   Если не запущен:
   ```powershell
   docker-compose up -d postgres
   ```

2. **Проверь порт 8081:**
   ```powershell
   netstat -ano | findstr :8081
   ```
   
   Если занят, останови процесс или измени `SERVER_PORT`.

3. **Проверь логи API сервера** в терминале, где запущен `go run cmd/api/main.go`

