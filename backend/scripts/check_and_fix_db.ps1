# Скрипт для проверки и исправления БД
Write-Host "=== Checking Database Status ===" -ForegroundColor Cyan

# Проверяем подключение к БД через backend
Write-Host "`n1. Testing backend health..." -ForegroundColor Yellow
try {
    $health = Invoke-WebRequest -Uri "http://localhost:8081/api/health" -UseBasicParsing
    Write-Host "   ✅ Backend is running" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Backend not running" -ForegroundColor Red
    exit 1
}

# Проверяем категории
Write-Host "`n2. Testing categories endpoint..." -ForegroundColor Yellow
try {
    $cats = Invoke-WebRequest -Uri "http://localhost:8081/api/v1/categories/tree" -UseBasicParsing
    $data = $cats.Content | ConvertFrom-Json
    Write-Host "   ✅ Categories endpoint works" -ForegroundColor Green
    Write-Host "   📊 Categories count: $($data.Length)" -ForegroundColor Gray
    if ($data.Length -eq 0) {
        Write-Host "   ⚠️  Categories table is empty - need to load seed data" -ForegroundColor Yellow
    }
} catch {
    Write-Host "   ❌ Categories error: $($_.Exception.Message)" -ForegroundColor Red
}

# Проверяем города
Write-Host "`n3. Testing cities endpoint..." -ForegroundColor Yellow
try {
    $cities = Invoke-WebRequest -Uri "http://localhost:8081/api/v1/cities" -UseBasicParsing
    $data = $cities.Content | ConvertFrom-Json
    Write-Host "   ✅ Cities endpoint works" -ForegroundColor Green
    Write-Host "   📊 Cities count: $($data.Length)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ Cities error (500): Table might not exist" -ForegroundColor Red
    Write-Host "   💡 Solution: Apply migrations (0005_catalog_core.up.sql)" -ForegroundColor Yellow
}

Write-Host "`n=== Summary ===" -ForegroundColor Cyan
Write-Host "If cities returns 500, the 'cities' table doesn't exist." -ForegroundColor Yellow
Write-Host "You need to:" -ForegroundColor Yellow
Write-Host "1. Make sure Docker Desktop is running" -ForegroundColor White
Write-Host "2. Apply migrations: cd backend; go run cmd/migrate/main.go" -ForegroundColor White
Write-Host "3. Load seed data (if needed)" -ForegroundColor White


