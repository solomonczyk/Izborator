# PowerShell скрипт для создания тестовой базы данных
# Запустите: .\backend\scripts\setup_test_db.ps1

Write-Host "=== Создание тестовой базы данных Izborator ===" -ForegroundColor Cyan
Write-Host ""

# Проверяем, запущен ли контейнер PostgreSQL
Write-Host "Проверка контейнера PostgreSQL..." -ForegroundColor Yellow
$postgresRunning = docker ps --filter "name=izborator_postgres" --format "{{.Names}}" 2>&1

if ($LASTEXITCODE -ne 0 -or -not $postgresRunning) {
    Write-Host "❌ Контейнер izborator_postgres не найден или Docker не запущен" -ForegroundColor Red
    Write-Host "Запустите: docker-compose up -d postgres" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ Контейнер найден: $postgresRunning" -ForegroundColor Green
Write-Host ""

# Создаём базу данных
Write-Host "Создание базы данных..." -ForegroundColor Yellow
Get-Content "backend\scripts\create_test_db.sql" | docker exec -i izborator_postgres psql -U postgres

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ База данных создана" -ForegroundColor Green
} else {
    Write-Host "⚠️ База данных уже существует или ошибка" -ForegroundColor Yellow
}

# Применяем схему
Write-Host "Применение схемы..." -ForegroundColor Yellow
Get-Content "backend\scripts\create_test_db_in_izborator.sql" | docker exec -i izborator_postgres psql -U postgres -d izborator

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Схема применена" -ForegroundColor Green
} else {
    Write-Host "❌ Ошибка при применении схемы" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Добавляем тестовые данные
Write-Host "Добавление тестовых данных..." -ForegroundColor Yellow
Get-Content "backend\scripts\seed_test_data.sql" | docker exec -i izborator_postgres psql -U postgres -d izborator

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Тестовые данные добавлены" -ForegroundColor Green
} else {
    Write-Host "⚠️ Предупреждение: возможно, данные уже существуют" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== Готово! ===" -ForegroundColor Cyan
Write-Host "База данных izborator готова к использованию" -ForegroundColor Green
Write-Host ""
Write-Host "Проверка:" -ForegroundColor Yellow
docker exec -i izborator_postgres psql -U postgres -d izborator -c 'SELECT COUNT(*) as products FROM products; SELECT COUNT(*) as prices FROM product_prices;'

