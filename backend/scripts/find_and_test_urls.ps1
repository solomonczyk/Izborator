# Скрипт для поиска и тестирования URL по категориям
# Использование: .\find_and_test_urls.ps1

Write-Host "=== Поиск валидных URL для парсинга ===" -ForegroundColor Green
Write-Host "`nКатегории для тестирования:" -ForegroundColor Cyan
Write-Host "1. Mobilni telefoni (Мобильные телефоны)" -ForegroundColor White
Write-Host "2. Laptopovi (Ноутбуки)" -ForegroundColor White

Write-Host "`nИнструкции:" -ForegroundColor Yellow
Write-Host "1. Открой https://gigatron.rs в браузере" -ForegroundColor White
Write-Host "2. Найди товар в нужной категории" -ForegroundColor White
Write-Host "3. Скопируй URL из адресной строки" -ForegroundColor White
Write-Host "4. Вставь URL ниже" -ForegroundColor White

Write-Host "`n=== Ввод URL ===" -ForegroundColor Green

# Мобильные телефоны
$phoneURL = Read-Host "Введи URL для мобильного телефона (или нажми Enter для пропуска)"
if ($phoneURL) {
    Write-Host "`nТестирую парсинг телефона..." -ForegroundColor Yellow
    Set-Location backend
    go run cmd/worker/main.go -url $phoneURL -shop "shop-001"
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✅ Парсинг телефона успешен!" -ForegroundColor Green
    } else {
        Write-Host "`n❌ Ошибка парсинга телефона" -ForegroundColor Red
    }
    Set-Location ..
}

# Ноутбуки
$laptopURL = Read-Host "`nВведи URL для ноутбука (или нажми Enter для пропуска)"
if ($laptopURL) {
    Write-Host "`nТестирую парсинг ноутбука..." -ForegroundColor Yellow
    Set-Location backend
    go run cmd/worker/main.go -url $laptopURL -shop "shop-001"
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✅ Парсинг ноутбука успешен!" -ForegroundColor Green
    } else {
        Write-Host "`n❌ Ошибка парсинга ноутбука" -ForegroundColor Red
    }
    Set-Location ..
}

Write-Host "`n=== Обработка сырых данных ===" -ForegroundColor Green
$process = Read-Host "Обработать сырые данные? (y/n)"
if ($process -eq "y" -or $process -eq "Y") {
    Write-Host "`nЗапускаю обработку..." -ForegroundColor Yellow
    Set-Location backend
    go run cmd/worker/main.go -process
    Set-Location ..
    Write-Host "`n✅ Обработка завершена!" -ForegroundColor Green
}

Write-Host "`n=== Готово ===" -ForegroundColor Green
Write-Host "Проверь результаты на фронтенде: http://localhost:3003/sr/catalog" -ForegroundColor Cyan

