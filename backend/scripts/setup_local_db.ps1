# Скрипт для создания базы данных на локальном PostgreSQL
# Используйте, если Docker недоступен, но есть локальный PostgreSQL

Write-Host "=== Создание базы данных на локальном PostgreSQL ===" -ForegroundColor Cyan
Write-Host ""

# Проверяем подключение к локальному PostgreSQL
Write-Host "Проверка подключения к PostgreSQL на порту 5432..." -ForegroundColor Yellow

$env:PGPASSWORD = "postgres"
$testConnection = psql -U postgres -h localhost -p 5432 -c "SELECT 1;" 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Не удалось подключиться к PostgreSQL" -ForegroundColor Red
    Write-Host "Убедитесь, что PostgreSQL запущен и доступен на localhost:5432" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Или используйте Docker:" -ForegroundColor Yellow
    Write-Host "  1. Запустите Docker Desktop"
    Write-Host "  2. Выполните: docker-compose up -d postgres"
    exit 1
}

Write-Host "✅ Подключение к PostgreSQL успешно" -ForegroundColor Green
Write-Host ""

# Создаём базу данных
Write-Host "Создание базы данных izborator..." -ForegroundColor Yellow
psql -U postgres -h localhost -p 5432 -f backend/scripts/create_test_db.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ База данных создана" -ForegroundColor Green
} else {
    Write-Host "⚠️ База данных уже существует или ошибка" -ForegroundColor Yellow
}

Write-Host ""

# Применяем схему
Write-Host "Применение схемы..." -ForegroundColor Yellow
psql -U postgres -h localhost -p 5432 -d izborator -f backend/scripts/create_test_db_in_izborator.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Схема применена" -ForegroundColor Green
} else {
    Write-Host "❌ Ошибка при применении схемы" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Добавляем тестовые данные
Write-Host "Добавление тестовых данных..." -ForegroundColor Yellow
psql -U postgres -h localhost -p 5432 -d izborator -f backend/scripts/seed_test_data.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Тестовые данные добавлены" -ForegroundColor Green
} else {
    Write-Host "⚠️ Предупреждение: возможно, данные уже существуют" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== Готово! ===" -ForegroundColor Cyan
Write-Host "База данных izborator готова к использованию" -ForegroundColor Green
Write-Host ""
Write-Host "Не забудьте обновить .env:" -ForegroundColor Yellow
Write-Host "  DB_PORT=5432" -ForegroundColor White
Write-Host "  DB_HOST=localhost" -ForegroundColor White

